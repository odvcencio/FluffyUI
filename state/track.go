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

var (
	trackerStackMu  sync.Mutex
	trackerStack    []*dependencyTracker
	activeTracker   atomic.Pointer[dependencyTracker]
)

// trackDependencies runs fn while recording any signal reads.
// Note: tracking is global and assumes compute runs on a single goroutine.
// Uses a stack-based approach to support nested tracking calls.
func trackDependencies[T any](fn func() T) (T, []Subscribable) {
	if fn == nil {
		var zero T
		return zero, nil
	}
	tracker := &dependencyTracker{deps: make(map[Subscribable]struct{})}
	trackerStackMu.Lock()
	trackerStack = append(trackerStack, tracker)
	// Update active tracker to top of stack
	activeTracker.Store(tracker)
	trackerStackMu.Unlock()

	result := fn()

	trackerStackMu.Lock()
	if len(trackerStack) > 0 {
		trackerStack = trackerStack[:len(trackerStack)-1]
	}
	// Update active tracker to new top of stack (or nil)
	if len(trackerStack) > 0 {
		activeTracker.Store(trackerStack[len(trackerStack)-1])
	} else {
		activeTracker.Store(nil)
	}
	trackerStackMu.Unlock()
	return result, tracker.list()
}

func recordDependency(dep Subscribable) {
	// Fast path: check atomic pointer first to avoid mutex
	tracker := activeTracker.Load()
	if tracker == nil {
		return
	}
	tracker.add(dep)
}
