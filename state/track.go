package state

import (
	"sync"
	"sync/atomic"
)

type dependencyTracker struct {
	mu   sync.Mutex
	deps map[Subscribable]struct{}
}

func (t *dependencyTracker) add(dep Subscribable) {
	if t == nil || dep == nil {
		return
	}
	t.mu.Lock()
	if t.deps == nil {
		t.deps = make(map[Subscribable]struct{})
	}
	t.deps[dep] = struct{}{}
	t.mu.Unlock()
}

func (t *dependencyTracker) list() []Subscribable {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.deps) == 0 {
		return nil
	}
	out := make([]Subscribable, 0, len(t.deps))
	for dep := range t.deps {
		out = append(out, dep)
	}
	return out
}

var currentTracker atomic.Pointer[dependencyTracker]

// trackDependencies runs fn while recording any signal reads.
// Note: tracking is global and assumes compute runs on a single goroutine.
func trackDependencies[T any](fn func() T) (T, []Subscribable) {
	if fn == nil {
		var zero T
		return zero, nil
	}
	tracker := &dependencyTracker{deps: make(map[Subscribable]struct{})}
	prev := currentTracker.Load()
	currentTracker.Store(tracker)
	result := fn()
	currentTracker.Store(prev)
	return result, tracker.list()
}

func recordDependency(dep Subscribable) {
	tracker := currentTracker.Load()
	if tracker == nil {
		return
	}
	tracker.add(dep)
}
