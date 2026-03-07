package runtime

import "fmt"

// RunEventLoop runs all scheduled tasks until completion or deadlock.
func RunEventLoop(sched *Scheduler) error {
	for {
		if !sched.HasTasks() {
			if sched.HasSuspended() {
				return fmt.Errorf(
					"deadlock: %d suspended tasks, no ready tasks",
					len(sched.suspended),
				)
			}
			break
		}

		task := sched.Next()

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

	return nil
}
