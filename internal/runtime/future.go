package runtime

import (
	"sync"

	"avenir/internal/value"
)

// Future represents a value that will be available in the future.
type Future struct {
	mu      sync.Mutex
	Ready   bool
	Result  value.Value
	Err     error
	waiters []*Task
}

// NewFuture creates a new unresolved Future.
func NewFuture() *Future {
	return &Future{}
}

// Resolve marks the future as ready with the given value and schedules all waiting tasks.
func (f *Future) Resolve(v value.Value) {
	f.mu.Lock()

	f.Ready = true
	f.Result = v

	waiters := f.waiters
	f.waiters = nil

	f.mu.Unlock()

	for _, t := range waiters {
		t.Scheduler.Schedule(t)
	}
}

// Reject marks the future as ready with an error and schedules all waiting tasks.
func (f *Future) Reject(err error) {
	f.mu.Lock()

	f.Ready = true
	f.Err = err

	waiters := f.waiters
	f.waiters = nil

	f.mu.Unlock()

	for _, t := range waiters {
		t.Scheduler.Schedule(t)
	}
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
