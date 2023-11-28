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
	if db.mask.CheckAny(closed) {
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

func newWatcher(maxsize int, expected ...EventType) *watcher {
	return &watcher{
		maxsize:  maxsize,
		expected: expected,
		events:   queues.NewArrayQueue[*Event](maxsize),
		eventCh:  make(chan *Event, 200),
	}
}

// watcher has the responsibility of maintaining events queue and channel
// it will send events while the behavior of database is changed
type watcher struct {
	maxsize  int
	expected []EventType
	events   *queues.ArrayQueue[*Event]
	eventCh  chan *Event
	mu       sync.Mutex
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
	if w.events.Size() >= w.maxsize {
		return
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
watch:
	for {
		select {
		case <-ctx.Done():
			break watch
		default:
			event := w.pop()
			if event != nil {
				w.eventCh <- event
			} else {
				time.Sleep(5 * time.Millisecond)
			}
		}
	}
}

func (w *watcher) close() {
	w.events = nil
	close(w.eventCh)
	w.eventCh = nil
}
