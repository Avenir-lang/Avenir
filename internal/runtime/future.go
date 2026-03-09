package runtime

import (
	"fmt"
	"sync"
	"time"

	"avenir/internal/value"
)

// Future represents a value that will be available in the future.
type Future struct {
	mu      sync.Mutex
	Ready   bool
	Result  value.Value
	Err     error
	waiters []*Task
	done    chan struct{}
}

// NewFuture creates a new unresolved Future.
func NewFuture() *Future {
	return &Future{done: make(chan struct{})}
}

// Resolve marks the future as ready with the given value and schedules all waiting tasks.
// No-op if the future is already resolved/rejected (first-write-wins).
func (f *Future) Resolve(v value.Value) {
	f.mu.Lock()
	if f.Ready {
		f.mu.Unlock()
		return
	}

	f.Ready = true
	f.Result = v

	waiters := f.waiters
	f.waiters = nil

	f.mu.Unlock()

	close(f.done)
	for _, t := range waiters {
		t.Scheduler.Schedule(t)
	}
}

// Reject marks the future as ready with an error and schedules all waiting tasks.
// No-op if the future is already resolved/rejected (first-write-wins).
func (f *Future) Reject(err error) {
	f.mu.Lock()
	if f.Ready {
		f.mu.Unlock()
		return
	}

	f.Ready = true
	f.Err = err

	waiters := f.waiters
	f.waiters = nil

	f.mu.Unlock()

	close(f.done)
	for _, t := range waiters {
		t.Scheduler.Schedule(t)
	}
}

// Wait blocks until the future is resolved or rejected.
func (f *Future) Wait() {
	<-f.done
}

// WithTimeout creates a new Future that races inner against a deadline.
// If inner resolves/rejects before durationNs nanoseconds, the result is forwarded.
// Otherwise the returned future is rejected with a timeout error.
func WithTimeout(inner *Future, durationNs int64) *Future {
	result := NewFuture()

	go func() {
		timer := time.NewTimer(time.Duration(durationNs))
		defer timer.Stop()

		select {
		case <-inner.done:
			inner.mu.Lock()
			res, err := inner.Result, inner.Err
			inner.mu.Unlock()
			if err != nil {
				result.Reject(err)
			} else {
				result.Resolve(res)
			}
		case <-timer.C:
			result.Reject(fmt.Errorf("timeout after %dms", durationNs/1000000))
		}
	}()

	return result
}

// AddWaiter registers a task as waiting for this future.
// Returns true if the task was added (future not ready yet, task should suspend).
// Returns false if the future is already ready (task should not suspend).
func (f *Future) AddWaiter(t *Task) bool {
	f.mu.Lock()

	if f.Ready {
		f.mu.Unlock()
		return false
	}

	f.waiters = append(f.waiters, t)
	f.mu.Unlock()

	return true
}
