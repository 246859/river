package riverdb

import (
	"context"
	"github.com/246859/containers/queues"
	"github.com/pkg/errors"
	"slices"
	"sync"
	"time"
)

type EventType uint

const (
	PutEvent EventType = 1 + iota
	DelEvent
	RollbackEvent
	MergeEvent
	BackupEvent
	RecoverEvent
)

var (
	ErrWatcherDisabled = errors.New("event watcher disabled")
)

func (db *DB) Watch() (<-chan *Event, error) {
	if db.flag.Check(closed) {
		return nil, ErrDBClosed
	}

	if db.watcher == nil {
		return nil, ErrWatcherDisabled
	}
	return db.watcher.eventCh, nil
}

// Event represents a push event
type Event struct {
	Type  EventType
	Value any
}

func newWatcher(queueSize int, expected ...EventType) *watcher {
	w := &watcher{
		expected:  expected,
		queueSize: queueSize,
		events:    queues.NewArrayQueue[*Event](queueSize),
	}

	eventsize := queueSize / 5
	if eventsize <= 0 {
		eventsize = 20
	}
	w.eventSize = eventsize
	// chan buffer size a little greater than eventsize is to prevent to block watch goroutine if chan is full
	w.eventCh = make(chan *Event, w.eventSize+5)

	return w
}

// watcher has the responsibility of maintaining events queue and channel
// it will send events while the behavior of database is changed
type watcher struct {
	expected  []EventType
	queueSize int
	events    *queues.ArrayQueue[*Event]

	eventSize int
	eventCh   chan *Event

	mu sync.Mutex
}

func (w *watcher) expect(et EventType) bool {
	return slices.Contains(w.expected, et)
}

func (w *watcher) pop() *Event {
	w.mu.Lock()
	defer w.mu.Unlock()
	e, _ := w.events.Pop()
	return e
}

func (w *watcher) push(eve *Event) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.events.Size() >= w.queueSize {
		w.events.Pop()
	}
	if !w.expect(eve.Type) {
		return
	}
	w.events.Push(eve)
}

func (w *watcher) clear() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events.Clear()
}

func (w *watcher) watch(ctx context.Context) {
	defer close(w.eventCh)

	for {
		select {
		case <-ctx.Done():
			return
		default:

			// if channel will be full soon, consume event in channel by self
			for len(w.eventCh) >= w.eventSize {
				_, ok := <-w.eventCh
				if !ok {
					return
				}
			}

			if event := w.pop(); event != nil {
				w.eventCh <- event
			} else {
				time.Sleep(5 * time.Millisecond)
			}
		}
	}
}

func (w *watcher) close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events = nil
}
