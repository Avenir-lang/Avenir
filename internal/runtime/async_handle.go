package runtime

import (
	"fmt"
	"sync"

	"avenir/internal/value"
)

// AsyncHandle represents a pending asynchronous operation.
// The actual I/O runs in a Go goroutine; on completion, the result is
// delivered via Resolve/Reject and any registered callbacks fire.
type AsyncHandle struct {
	mu         sync.Mutex
	done       chan struct{}
	ready      bool
	result     value.Value
	err        error
	onComplete func()
}

// NewAsyncHandle creates a new unresolved AsyncHandle.
func NewAsyncHandle() *AsyncHandle {
	return &AsyncHandle{
		done: make(chan struct{}),
	}
}

// Resolve completes the handle with a successful result.
func (h *AsyncHandle) Resolve(v value.Value) {
	h.mu.Lock()
	if h.ready {
		h.mu.Unlock()
		return
	}
	h.ready = true
	h.result = v
	cb := h.onComplete
	h.mu.Unlock()

	close(h.done)
	if cb != nil {
		cb()
	}
}

// Reject completes the handle with an error.
func (h *AsyncHandle) Reject(err error) {
	h.mu.Lock()
	if h.ready {
		h.mu.Unlock()
		return
	}
	h.ready = true
	h.err = err
	cb := h.onComplete
	h.mu.Unlock()

	close(h.done)
	if cb != nil {
		cb()
	}
}

// Poll checks whether the handle has completed.
// Returns (result, err, ready).
func (h *AsyncHandle) Poll() (value.Value, error, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.ready {
		return value.Value{}, nil, false
	}
	return h.result, h.err, true
}

// Wait blocks until the handle completes.
func (h *AsyncHandle) Wait() {
	<-h.done
}

// OnComplete registers a callback that fires when the handle completes.
// If the handle is already complete, the callback fires immediately.
func (h *AsyncHandle) OnComplete(fn func()) {
	h.mu.Lock()
	if h.ready {
		h.mu.Unlock()
		fn()
		return
	}
	h.onComplete = fn
	h.mu.Unlock()
}

// WireToFuture connects this AsyncHandle to a Future:
// when the handle completes, the future is resolved or rejected.
func (h *AsyncHandle) WireToFuture(fut *Future) {
	h.OnComplete(func() {
		res, err, _ := h.Poll()
		if err != nil {
			fut.Reject(err)
		} else {
			fut.Resolve(res)
		}
	})
}

// RunAsync is a convenience helper that starts a goroutine performing fn
// and wires the result to the returned AsyncHandle. Panics are recovered.
func RunAsync(fn func() (value.Value, error)) *AsyncHandle {
	ah := NewAsyncHandle()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ah.Reject(fmt.Errorf("async panic: %v", r))
			}
		}()
		result, err := fn()
		if err != nil {
			ah.Reject(err)
		} else {
			ah.Resolve(result)
		}
	}()
	return ah
}
