package runtime

// TaskStatus represents the current state of an async task.
type TaskStatus int

const (
	TaskReady     TaskStatus = iota
	TaskRunning
	TaskSuspended
	TaskDone
	TaskFailed
)

// Task represents an async function's execution context.
type Task struct {
	ID        int
	Status    TaskStatus
	Future    *Future
	Scheduler *Scheduler
	StepFn    func() (TaskStatus, error)
}
