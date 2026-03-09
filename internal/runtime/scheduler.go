package runtime

import "sync"

// Scheduler manages the ready queue and suspended tasks for the async event loop.
type Scheduler struct {
	mu         sync.Mutex
	readyQueue []*Task
	suspended  map[int]*Task
	nextID     int
	wakeup     chan struct{}
}

// NewScheduler creates a new empty Scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		suspended: make(map[int]*Task),
		wakeup:    make(chan struct{}, 1),
	}
}

// NewTask creates a new Task with a unique ID, associated with this scheduler.
func (s *Scheduler) NewTask(future *Future, stepFn func() (TaskStatus, error)) *Task {
	id := s.nextID
	s.nextID++
	return &Task{
		ID:        id,
		Status:    TaskReady,
		Future:    future,
		Scheduler: s,
		StepFn:    stepFn,
	}
}

// Schedule adds a task to the ready queue and removes it from suspended if present.
// It also signals the wakeup channel so the event loop unblocks.
func (s *Scheduler) Schedule(t *Task) {
	s.mu.Lock()
	delete(s.suspended, t.ID)
	t.Status = TaskReady
	s.readyQueue = append(s.readyQueue, t)
	s.mu.Unlock()
	s.Signal()
}

// Signal sends a non-blocking signal on the wakeup channel.
// This wakes the event loop when it is waiting for I/O completion.
func (s *Scheduler) Signal() {
	select {
	case s.wakeup <- struct{}{}:
	default:
	}
}

// WaitForWakeup blocks until a signal arrives on the wakeup channel.
func (s *Scheduler) WaitForWakeup() {
	<-s.wakeup
}

// Next dequeues and returns the first ready task, or nil if the queue is empty.
func (s *Scheduler) Next() *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.readyQueue) == 0 {
		return nil
	}
	t := s.readyQueue[0]
	s.readyQueue = s.readyQueue[1:]
	return t
}

// HasTasks returns true if there are tasks in the ready queue.
func (s *Scheduler) HasTasks() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.readyQueue) > 0
}

// Suspend moves a task to the suspended set.
func (s *Scheduler) Suspend(t *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.Status = TaskSuspended
	s.suspended[t.ID] = t
}

// HasSuspended returns true if there are suspended tasks.
func (s *Scheduler) HasSuspended() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.suspended) > 0
}

// IsIdle atomically checks whether the scheduler has no ready and no suspended tasks.
func (s *Scheduler) IsIdle() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.readyQueue) == 0 && len(s.suspended) == 0
}
