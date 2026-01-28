package state

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestResource_FetchSuccess(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	res := NewResource(func() (int, error) {
		close(started)
		<-release
		return 7, nil
	})

	<-started
	state := res.Get()
	if !state.Loading {
		t.Fatal("expected Loading=true while fetch is running")
	}

	done := make(chan struct{})
	var once sync.Once
	res.Subscribe(func() {
		if !res.Get().Loading {
			once.Do(func() { close(done) })
		}
	})

	close(release)
	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("timeout waiting for resource to finish loading")
	}

	state = res.Get()
	if state.Loading {
		t.Fatal("expected Loading=false after fetch")
	}
	if state.Error != nil {
		t.Fatalf("unexpected error: %v", state.Error)
	}
	if state.Data != 7 {
		t.Fatalf("unexpected data: %d", state.Data)
	}
}

func TestResource_FetchError(t *testing.T) {
	release := make(chan struct{})
	res := NewResource(func() (int, error) {
		<-release
		return 0, errors.New("boom")
	})

	close(release)
	deadline := time.After(250 * time.Millisecond)
	for {
		state := res.Get()
		if !state.Loading {
			if state.Error == nil {
				t.Fatal("expected error to be set")
			}
			return
		}
		select {
		case <-deadline:
			t.Fatal("timeout waiting for resource error")
		default:
		}
	}
}

func TestResource_RefetchOnDeps(t *testing.T) {
	dep := NewSignal(0)
	var count int32
	fetchCh := make(chan int, 4)

	res := NewResource(func() (int, error) {
		n := int(atomic.AddInt32(&count, 1))
		fetchCh <- n
		return n, nil
	}, dep)

	_ = res
	first := <-fetchCh
	if first != 1 {
		t.Fatalf("expected first fetch count 1, got %d", first)
	}

	dep.Set(1)
	second := <-fetchCh
	if second != 2 {
		t.Fatalf("expected second fetch count 2, got %d", second)
	}
}
