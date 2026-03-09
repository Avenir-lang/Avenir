package runtime

// RunEventLoop runs all scheduled tasks until completion.
// When no ready tasks exist but suspended tasks remain (waiting for async I/O),
// the loop blocks on the scheduler's wakeup channel until a goroutine signals
// that a future has been resolved/rejected.
func RunEventLoop(sched *Scheduler) error {
	for {
		for !sched.HasTasks() {
			if sched.IsIdle() {
				return nil
			}
			sched.WaitForWakeup()
		}

		task := sched.Next()
		if task == nil {
			continue
		}

		task.Status = TaskRunning
		newStatus, err := task.StepFn()

		if err != nil {
			task.Status = TaskFailed
			task.Future.Reject(err)
			continue
		}

		task.Status = newStatus

		if newStatus == TaskSuspended {
			sched.Suspend(task)
		}
	}
}
