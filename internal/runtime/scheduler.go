package runtime

// Scheduler manages the ready queue and suspended tasks for the async event loop.
type Scheduler struct {
	readyQueue []*Task
	suspended  map[int]*Task
	nextID     int
}

// NewScheduler creates a new empty Scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		suspended: make(map[int]*Task),
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
func (s *Scheduler) Schedule(t *Task) {
	delete(s.suspended, t.ID)
	t.Status = TaskReady
	s.readyQueue = append(s.readyQueue, t)
}

// Next dequeues and returns the first ready task, or nil if the queue is empty.
func (s *Scheduler) Next() *Task {
	if len(s.readyQueue) == 0 {
		return nil
	}
	t := s.readyQueue[0]
	s.readyQueue = s.readyQueue[1:]
	return t
}

// HasTasks returns true if there are tasks in the ready queue.
func (s *Scheduler) HasTasks() bool {
	return len(s.readyQueue) > 0
}

// Suspend moves a task to the suspended set.
func (s *Scheduler) Suspend(t *Task) {
	t.Status = TaskSuspended
	s.suspended[t.ID] = t
}

// HasSuspended returns true if there are suspended tasks.
func (s *Scheduler) HasSuspended() bool {
	return len(s.suspended) > 0
}
