package agent

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRequestIsExpiredAndWait(t *testing.T) {
	req := &Request{Deadline: time.Now().Add(-time.Second), done: make(chan struct{})}
	if !req.IsExpired(time.Now()) {
		t.Fatalf("expected expired")
	}
	close(req.done)
	if _, err := req.Wait(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := (*Request)(nil).Wait(context.Background()); err == nil {
		t.Fatalf("expected error for nil request")
	}
}

func TestRequestQueueFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := &RequestQueue{
		maxSize:        1,
		maxPerPriority: 1,
		ctx:            ctx,
		cancel:         cancel,
		notEmpty:       make(chan struct{}, 1),
	}

	req1 := &Request{Priority: RequestPriorityNormal}
	if err := q.Enqueue(req1); err != nil {
		t.Fatalf("unexpected enqueue error: %v", err)
	}
	if err := q.Enqueue(&Request{Priority: RequestPriorityNormal}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected queue full")
	}
	if q.TryEnqueue(&Request{Priority: RequestPriorityNormal}) {
		t.Fatalf("expected TryEnqueue false")
	}
	if err := q.EnqueueBackground(&Request{Priority: RequestPriorityLow}); err != nil {
		t.Fatalf("expected background enqueue despite full total size, got %v", err)
	}
	if err := q.EnqueueBackground(&Request{Priority: RequestPriorityLow}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected background queue full")
	}
}

func TestRequestQueueProcessAndCallbacks(t *testing.T) {
	q := NewRequestQueue(QueueOptions{MaxSize: 10, MaxPerPriority: 10, Workers: 1})
	defer q.Stop()

	started := make(chan struct{}, 1)
	done := make(chan struct{}, 1)
	q.SetRequestStartCallback(func(req *Request) { started <- struct{}{} })
	q.SetRequestDoneCallback(func(req *Request, _ time.Duration, err error) { done <- struct{}{} })

	req := &Request{Priority: RequestPriorityHigh}
	req.Execute = func(ctx context.Context) error {
		return nil
	}

	if err := q.Enqueue(req); err != nil {
		t.Fatalf("enqueue error: %v", err)
	}
	if q.Size() < 0 || q.ActiveCount() < 0 {
		t.Fatalf("unexpected queue counts: size=%d active=%d", q.Size(), q.ActiveCount())
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for start callback")
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for done callback")
	}

	if _, err := req.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected wait error: %v", err)
	}
	if q.Size() < 0 || q.ActiveCount() < 0 {
		t.Fatalf("unexpected queue counts after wait: size=%d active=%d", q.Size(), q.ActiveCount())
	}

	stats := q.Stats()
	if stats.TotalQueued == 0 || stats.TotalDone == 0 {
		t.Fatalf("expected queue stats to update")
	}
}

func TestAsyncExecute(t *testing.T) {
	q := NewRequestQueue(QueueOptions{MaxSize: 10, MaxPerPriority: 10, Workers: 1})
	defer q.Stop()

	result := q.AsyncExecute(func(ctx context.Context) (any, error) {
		return "ok", nil
	}, RequestPriorityNormal)
	value, err := result.WaitTimeout(2 * time.Second)
	if err != nil {
		t.Fatalf("unexpected async error: %v", err)
	}
	if value != "ok" {
		t.Fatalf("unexpected async value")
	}
	if !result.IsDone() {
		t.Fatalf("expected async result done")
	}

	bg := q.AsyncExecuteBackground(func(ctx context.Context) (any, error) {
		return 42, nil
	})
	value, err = bg.WaitTimeout(2 * time.Second)
	if err != nil {
		t.Fatalf("unexpected async background error: %v", err)
	}
	if value != 42 {
		t.Fatalf("unexpected async background value")
	}
	if !bg.IsDone() {
		t.Fatalf("expected background result done")
	}
}

func TestAsyncExecuteQueueNil(t *testing.T) {
	var q *RequestQueue
	res := q.AsyncExecute(func(ctx context.Context) (any, error) { return nil, nil }, RequestPriorityLow)
	if _, err := res.WaitTimeout(10 * time.Millisecond); err == nil {
		t.Fatalf("expected error for nil queue")
	}
}
