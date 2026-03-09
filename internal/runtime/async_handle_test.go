package runtime

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"avenir/internal/value"
)

func TestAsyncHandleResolve(t *testing.T) {
	ah := NewAsyncHandle()

	_, _, ready := ah.Poll()
	if ready {
		t.Fatal("expected not ready before resolve")
	}

	ah.Resolve(value.Int(42))

	res, err, ready := ah.Poll()
	if !ready {
		t.Fatal("expected ready after resolve")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Kind != value.KindInt || res.Int != 42 {
		t.Fatalf("expected Int(42), got %v", res)
	}
}

func TestAsyncHandleReject(t *testing.T) {
	ah := NewAsyncHandle()
	ah.Reject(fmt.Errorf("test error"))

	_, err, ready := ah.Poll()
	if !ready {
		t.Fatal("expected ready after reject")
	}
	if err == nil || err.Error() != "test error" {
		t.Fatalf("expected 'test error', got %v", err)
	}
}

func TestAsyncHandleDoubleResolve(t *testing.T) {
	ah := NewAsyncHandle()
	ah.Resolve(value.Int(1))
	ah.Resolve(value.Int(2))

	res, _, _ := ah.Poll()
	if res.Int != 1 {
		t.Fatalf("expected first resolve value 1, got %d", res.Int)
	}
}

func TestAsyncHandleWait(t *testing.T) {
	ah := NewAsyncHandle()
	go func() {
		time.Sleep(10 * time.Millisecond)
		ah.Resolve(value.Str("done"))
	}()

	ah.Wait()
	res, err, ready := ah.Poll()
	if !ready {
		t.Fatal("expected ready after Wait")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Str != "done" {
		t.Fatalf("expected 'done', got %q", res.Str)
	}
}

func TestAsyncHandleOnComplete(t *testing.T) {
	ah := NewAsyncHandle()

	var called bool
	var mu sync.Mutex

	ah.OnComplete(func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	ah.Resolve(value.Int(1))
	time.Sleep(5 * time.Millisecond)

	mu.Lock()
	if !called {
		t.Fatal("OnComplete callback was not called")
	}
	mu.Unlock()
}

func TestAsyncHandleOnCompleteAlreadyReady(t *testing.T) {
	ah := NewAsyncHandle()
	ah.Resolve(value.Int(1))

	called := false
	ah.OnComplete(func() {
		called = true
	})

	if !called {
		t.Fatal("OnComplete should fire immediately if already resolved")
	}
}

func TestWireToFutureResolve(t *testing.T) {
	ah := NewAsyncHandle()
	fut := NewFuture()
	ah.WireToFuture(fut)

	ah.Resolve(value.Int(99))
	time.Sleep(5 * time.Millisecond)

	if !fut.Ready {
		t.Fatal("future should be ready")
	}
	if fut.Err != nil {
		t.Fatalf("unexpected error: %v", fut.Err)
	}
	if fut.Result.Int != 99 {
		t.Fatalf("expected 99, got %d", fut.Result.Int)
	}
}

func TestWireToFutureReject(t *testing.T) {
	ah := NewAsyncHandle()
	fut := NewFuture()
	ah.WireToFuture(fut)

	ah.Reject(fmt.Errorf("io error"))
	time.Sleep(5 * time.Millisecond)

	if !fut.Ready {
		t.Fatal("future should be ready")
	}
	if fut.Err == nil || fut.Err.Error() != "io error" {
		t.Fatalf("expected 'io error', got %v", fut.Err)
	}
}

func TestRunAsync(t *testing.T) {
	ah := RunAsync(func() (value.Value, error) {
		time.Sleep(10 * time.Millisecond)
		return value.Str("hello"), nil
	})

	ah.Wait()
	res, err, ready := ah.Poll()
	if !ready {
		t.Fatal("expected ready")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Str != "hello" {
		t.Fatalf("expected 'hello', got %q", res.Str)
	}
}

func TestRunAsyncPanicRecovery(t *testing.T) {
	ah := RunAsync(func() (value.Value, error) {
		panic("boom")
	})

	ah.Wait()
	_, err, ready := ah.Poll()
	if !ready {
		t.Fatal("expected ready")
	}
	if err == nil {
		t.Fatal("expected error from panic")
	}
	if err.Error() != "async panic: boom" {
		t.Fatalf("expected 'async panic: boom', got %q", err.Error())
	}
}

func TestSchedulerWakeup(t *testing.T) {
	sched := NewScheduler()
	fut := NewFuture()

	task := sched.NewTask(fut, func() (TaskStatus, error) {
		return TaskDone, nil
	})

	sched.Suspend(task)

	if sched.HasTasks() {
		t.Fatal("should not have ready tasks")
	}
	if !sched.HasSuspended() {
		t.Fatal("should have suspended tasks")
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		sched.Schedule(task)
	}()

	sched.WaitForWakeup()

	if !sched.HasTasks() {
		t.Fatal("should have ready tasks after schedule")
	}
}

func TestEventLoopWithAsyncIO(t *testing.T) {
	sched := NewScheduler()
	mainFut := NewFuture()

	var result value.Value

	step := 0
	ioFuture := NewFuture()

	var task *Task
	task = sched.NewTask(mainFut, func() (TaskStatus, error) {
		switch step {
		case 0:
			ah := RunAsync(func() (value.Value, error) {
				time.Sleep(20 * time.Millisecond)
				return value.Str("async result"), nil
			})
			ah.WireToFuture(ioFuture)

			if !ioFuture.Ready {
				ioFuture.AddWaiter(task)
				step = 1
				return TaskSuspended, nil
			}
			result = ioFuture.Result
			mainFut.Resolve(result)
			return TaskDone, nil

		case 1:
			if ioFuture.Err != nil {
				return TaskFailed, ioFuture.Err
			}
			result = ioFuture.Result
			mainFut.Resolve(result)
			return TaskDone, nil
		}
		return TaskFailed, fmt.Errorf("unexpected step %d", step)
	})

	sched.Schedule(task)

	err := RunEventLoop(sched)
	if err != nil {
		t.Fatalf("event loop error: %v", err)
	}

	if !mainFut.Ready {
		t.Fatal("main future should be ready")
	}
	if mainFut.Err != nil {
		t.Fatalf("unexpected error: %v", mainFut.Err)
	}
	if mainFut.Result.Str != "async result" {
		t.Fatalf("expected 'async result', got %q", mainFut.Result.Str)
	}
}
