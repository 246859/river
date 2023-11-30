package wal

import (
	"sync"
)

func (w *Wal) Pending(capacity int) *Pending {
	return &Pending{
		wal:          w,
		pendingSize:  0,
		pendingBytes: make([][]byte, 0, capacity),
	}
}

// Pending provides a way to write data in batches
type Pending struct {
	wal *Wal

	pendingSize  int64
	pendingBytes [][]byte
	mutex        sync.Mutex
}

func (p *Pending) Write(data []byte) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if int64(len(data)) >= p.wal.option.MaxFileSize {
		return ErrDataExceedFile
	}

	p.pendingSize += estimateBlockSize(int64(len(data)))
	p.pendingBytes = append(p.pendingBytes, data)

	return nil
}

func (p *Pending) Flush(sync bool) ([]ChunkPos, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.pendingSize == 0 {
		return []ChunkPos{}, nil
	}

	defer p.rest()

	if p.pendingSize >= p.wal.option.MaxFileSize {
		return nil, ErrDataExceedFile
	}

	p.wal.mutex.Lock()
	defer p.wal.mutex.Unlock()

	// if no left space, rotate a new one
	if p.pendingSize+p.wal.active.Size() >= p.wal.option.MaxFileSize {
		if err := p.wal.rotate(); err != nil {
			return nil, err
		}
	}

	allpos, err := p.wal.active.WriteAll(p.pendingBytes)
	if err != nil {
		return nil, err
	}

	if err = p.wal.sync(p.pendingSize, sync); err != nil {
		return nil, err
	}

	return allpos, nil
}

func (p *Pending) rest() {
	p.pendingBytes = make([][]byte, 0, len(p.pendingBytes)/2)
	p.pendingSize = 0
}

func (p *Pending) Reset() {
	p.mutex.Lock()
	p.rest()
	p.mutex.Unlock()
}
