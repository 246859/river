package riverdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/246859/containers/queues"
	"slices"
	"sync"
)

var (
	ErrWatcherClosed   = errors.New("watcher is closed")
	ErrInvalidEvent    = errors.New("invalid event")
	ErrWatcherDisabled = errors.New("event watcher disabled")
)

type EventType uint

// Event represents a push event
type Event struct {
	Type  EventType
	Value any
}

const (
	PutEvent EventType = 1 + iota
	DelEvent
	RollbackEvent
	MergeEvent
	BackupEvent
	RecoverEvent
)

func validEvent(et EventType) bool {
	ets := []EventType{
		PutEvent,
		DelEvent,
		RollbackEvent,
		MergeEvent,
		BackupEvent,
		RecoverEvent,
	}
	return slices.Contains(ets, et)
}

// Watcher returns a new event watcher with the given event type, if events is empty, it will apply db.Option
func (db *DB) Watcher(events ...EventType) (*Watcher, error) {

	if db.flag.Check(closed) {
		return nil, ErrDBClosed
	}

	if db.option.WatchSize == 0 {
		return nil, ErrWatcherDisabled
	}

	for _, event := range events {
		if !validEvent(event) {
			return nil, ErrInvalidEvent
		}
	}

	w := db.watcher.newWatcher(events...)
	return w, nil
}

func newWatcherPool(ctx context.Context, queueSize int, events ...EventType) (*watcherPool, error) {
	for _, event := range events {
		if !validEvent(event) {
			return nil, ErrInvalidEvent
		}
	}

	w := &watcherPool{
		ctx:           ctx,
		expectedEvent: events,
		pool:          make([]*Watcher, 0, 20),
		queueSize:     queueSize,
		eventQueue:    queues.NewArrayQueue[*Event](queueSize),
	}

	w.chanSize = queueSize / 10
	if queueSize < 100 {
		w.chanSize = queueSize / 5
	} else if queueSize <= 0 {
		w.chanSize = 10
	}

	return w, nil
}

type watcherPool struct {
	ctx           context.Context
	mu            sync.Mutex
	pool          []*Watcher
	expectedEvent []EventType

	eventQueue *queues.ArrayQueue[*Event]

	queueSize int
	chanSize  int

	done chan struct{}
}

func (w *watcherPool) newWatcher(ets ...EventType) *Watcher {
	if len(ets) == 0 {
		ets = w.expectedEvent
	}
	wa := &Watcher{}
	wa.watch = w
	// chan buffer size a little greater than eventsize is to prevent to block watch goroutine if chan is full
	wa.ch = make(chan *Event, w.chanSize+5)
	wa.expectedEvent = ets

	w.pool = append(w.pool, wa)

	return wa
}

func (w *watcherPool) expected(e EventType) bool {
	return expectedEvents(w.expectedEvent, e)
}

func (w *watcherPool) pop() *Event {
	w.mu.Lock()
	defer w.mu.Unlock()
	e, has := w.eventQueue.Pop()
	if !has {
		return nil
	}
	return e
}

func (w *watcherPool) push(e *Event) {
	w.mu.Lock()
	size := w.eventQueue.Size()
	w.mu.Unlock()

	if size >= w.queueSize {
		w.pop()
	}

	w.mu.Lock()
	w.eventQueue.Push(e)
	w.mu.Unlock()
}

// watch the events in the event queue, and notify the watcher in pool
// it must be running in another goroutine, and it will return when db closed
func (w *watcherPool) watch() {
	defer func() {
		w.Close()
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			event := w.pop()
			if event == nil {
				break
			}

			// notify watcher in pool
			for _, wa := range w.pool {

				if w.eventQueue == nil {
					return
				}

				if wa.closed || !wa.expected(event.Type) {
					continue
				}

				for len(wa.ch) >= w.chanSize {
					fmt.Println("out", (<-wa.ch).Type)
				}

				wa.ch <- event
			}
		}
	}
}

// Close closes all the live watchers in pool and discard self
func (w *watcherPool) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, wa := range w.pool {
		if wa.closed {
			continue
		}
		close(wa.ch)
	}
	w.pool = nil
	w.eventQueue = nil
}

// only use for Watcher to call
func (w *watcherPool) closeCh(ch chan *Event) {
	for i, wa := range w.pool {
		if wa.closed {
			return
		}

		if wa.ch == ch {
			close(wa.ch)
			w.pool = slices.Delete(w.pool, i, i+1)
			wa.ch = nil
			break
		}
	}
}

// Watcher
// a watcher should be closed after it used up all, or it will be closed when db closed at last
type Watcher struct {
	watch         *watcherPool
	ch            chan *Event
	expectedEvent []EventType
	closed        bool
}

func (w *Watcher) Listen() (<-chan *Event, error) {
	if w.closed {
		return nil, ErrWatcherClosed
	}
	return w.ch, nil
}

func (w *Watcher) expected(e EventType) bool {
	return expectedEvents(w.expectedEvent, e)
}

// Close closes the watcher.
func (w *Watcher) Close() error {
	w.watch.mu.Lock()
	defer w.watch.mu.Unlock()

	if w.closed {
		return ErrWatcherClosed
	}

	w.watch.closeCh(w.ch)
	w.closed = true
	return nil
}

func expectedEvents(events []EventType, e EventType) bool {
	return slices.Contains(events, e)
}
