package agent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestTaskStatusString(t *testing.T) {
	cases := map[TaskStatus]string{
		TaskPending:    "pending",
		TaskRunning:    "running",
		TaskPaused:     "paused",
		TaskCompleted:  "completed",
		TaskFailed:     "failed",
		TaskCancelled:  "cancelled",
		TaskStatus(99): "unknown",
	}
	for status, expected := range cases {
		if got := status.String(); got != expected {
			t.Fatalf("status %v = %q", status, got)
		}
	}
}

func TestBackgroundTaskLifecycle(t *testing.T) {
	task := NewBackgroundTask("id-1", "name", "desc", "session", func(ctx context.Context, task *BackgroundTask) error {
		task.SetProgress(10)
		task.SetProgress(200)
		return nil
	})

	task.SetProgress(-5)
	if task.Progress() != 0 {
		t.Fatalf("expected progress to clamp to 0, got %d", task.Progress())
	}

	if err := task.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := task.Start(); err == nil {
		t.Fatal("expected start to fail when already running")
	}

	if err := task.WaitTimeout(time.Second); err != nil {
		t.Fatalf("wait: %v", err)
	}

	if task.Status() != TaskCompleted {
		t.Fatalf("status = %v", task.Status())
	}
	if !task.IsDone() {
		t.Fatal("expected task to be done")
	}
	if task.Progress() != 100 {
		t.Fatalf("progress = %d", task.Progress())
	}
	if task.Error() != nil {
		t.Fatalf("unexpected error: %v", task.Error())
	}
	if task.StartedAt().IsZero() {
		t.Fatal("expected started time")
	}
	if task.CompletedAt().IsZero() {
		t.Fatal("expected completed time")
	}
	if task.Duration() <= 0 {
		t.Fatalf("duration = %v", task.Duration())
	}

	stats := task.Stats()
	if stats.Status != "completed" || stats.Progress != 100 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.CompletedAt == nil || stats.CompletedAt.IsZero() {
		t.Fatal("expected stats completed time")
	}
}

func TestBackgroundTaskCancelAndWait(t *testing.T) {
	task := NewBackgroundTask("id-2", "cancel", "", "", func(ctx context.Context, task *BackgroundTask) error {
		<-ctx.Done()
		return ctx.Err()
	})

	if err := task.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	waitCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := task.Wait(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}

	task.Cancel()
	if err := task.WaitTimeout(time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error, got %v", err)
	}
	if task.Status() != TaskCancelled {
		t.Fatalf("status = %v", task.Status())
	}
	if !task.IsDone() {
		t.Fatal("expected task to be done after cancel")
	}
}

func TestBackgroundTaskManager(t *testing.T) {
	block := make(chan struct{})
	taskFn := func(ctx context.Context, task *BackgroundTask) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-block:
			return nil
		}
	}

	mgr := NewBackgroundTaskManager(2, 1)
	var started atomic.Bool
	var done atomic.Bool
	mgr.SetTaskStartCallback(func(t *BackgroundTask) { started.Store(true) })
	mgr.SetTaskDoneCallback(func(t *BackgroundTask) { done.Store(true) })

	task1, err := mgr.Submit("id-1", "task1", "", "session-1", taskFn)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if !started.Load() {
		t.Fatal("expected start callback")
	}

	if mgr.Count() != 1 {
		t.Fatalf("count = %d", mgr.Count())
	}
	if mgr.Get("id-1") == nil {
		t.Fatal("expected task in manager")
	}
	if len(mgr.List()) != 1 {
		t.Fatalf("list = %v", mgr.List())
	}
	if len(mgr.ListSession("session-1")) != 1 {
		t.Fatalf("list session = %v", mgr.ListSession("session-1"))
	}
	if len(mgr.Stats()) != 1 {
		t.Fatalf("stats len = %d", len(mgr.Stats()))
	}

	if _, err := mgr.Submit("id-2", "task2", "", "session-1", taskFn); err == nil {
		t.Fatal("expected session limit error")
	}
	if _, err := mgr.Submit("id-1", "dup", "", "session-2", taskFn); err == nil {
		t.Fatal("expected duplicate id error")
	}

	if !mgr.Cancel("id-1") {
		t.Fatal("expected cancel to return true")
	}
	if mgr.Cancel("missing") {
		t.Fatal("expected cancel missing to return false")
	}

	if err := task1.WaitTimeout(time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancel error, got %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for mgr.Count() != 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if mgr.Count() != 0 {
		t.Fatalf("expected tasks to be cleaned up, count=%d", mgr.Count())
	}
	if !done.Load() {
		t.Fatal("expected done callback")
	}
}

func TestBackgroundTaskManagerCancelSession(t *testing.T) {
	block := make(chan struct{})
	taskFn := func(ctx context.Context, task *BackgroundTask) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-block:
			return nil
		}
	}

	mgr := NewBackgroundTaskManager(3, 2)
	taskA, err := mgr.Submit("id-a", "taskA", "", "session-a", taskFn)
	if err != nil {
		t.Fatalf("submit A: %v", err)
	}
	taskB, err := mgr.Submit("id-b", "taskB", "", "session-a", taskFn)
	if err != nil {
		t.Fatalf("submit B: %v", err)
	}
	taskC, err := mgr.Submit("id-c", "taskC", "", "session-b", taskFn)
	if err != nil {
		t.Fatalf("submit C: %v", err)
	}

	if mgr.CancelSession("") != 0 {
		t.Fatal("expected empty session cancel to return 0")
	}

	if count := mgr.CancelSession("session-a"); count != 2 {
		t.Fatalf("expected cancel count 2, got %d", count)
	}

	if err := taskA.WaitTimeout(time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("task A error = %v", err)
	}
	if err := taskB.WaitTimeout(time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("task B error = %v", err)
	}

	close(block)
	if err := taskC.WaitTimeout(time.Second); err != nil {
		t.Fatalf("task C error = %v", err)
	}
}

func TestBackgroundJobAndGenerateID(t *testing.T) {
	mgr := NewBackgroundTaskManager(5, 2)
	job, err := mgr.SubmitSimple("simple", func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("submit simple: %v", err)
	}
	if job == nil {
		t.Fatal("expected job")
	}
	if err := job.Wait(context.Background()); err != nil {
		t.Fatalf("job wait: %v", err)
	}
	if job.Progress() != 100 {
		t.Fatalf("job progress = %d", job.Progress())
	}

	var nilJob *BackgroundJob
	if nilJob.IsRunning() {
		t.Fatal("expected nil job to not be running")
	}
	if err := nilJob.Wait(context.Background()); err == nil {
		t.Fatal("expected error for nil job wait")
	}
	nilJob.Cancel()
	if nilJob.Progress() != 0 {
		t.Fatal("expected nil job progress to be 0")
	}

	first := generateTaskID()
	second := generateTaskID()
	if first == "" || second == "" || first == second {
		t.Fatalf("unexpected task ids: %q %q", first, second)
	}
}
