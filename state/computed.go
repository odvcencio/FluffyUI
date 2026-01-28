package state

import "sync"

// Computed derives its value from other signals.
type Computed[T any] struct {
	signal    *Signal[T]
	compute   func() T
	mu        sync.Mutex
	unsubs    []func()
	scheduler Scheduler
	autoDeps  bool
}

// NewComputed creates a derived value from dependencies. If no deps are
// provided, dependencies are detected automatically by tracking signal reads.
func NewComputed[T any](compute func() T, deps ...Subscribable) *Computed[T] {
	return NewComputedWithScheduler(nil, compute, deps...)
}

// NewComputedWithScheduler creates a derived value and schedules recomputes.
// If no deps are provided, dependencies are detected automatically.
func NewComputedWithScheduler[T any](scheduler Scheduler, compute func() T, deps ...Subscribable) *Computed[T] {
	if compute == nil {
		compute = func() T {
			var zero T
			return zero
		}
	}
	c := &Computed[T]{compute: compute, scheduler: scheduler}
	if len(deps) == 0 {
		value, tracked := trackDependencies(compute)
		c.signal = NewSignal(value)
		c.autoDeps = true
		c.setDeps(tracked)
		return c
	}
	c.signal = NewSignal(compute())
	c.setDeps(deps)
	return c
}

// SetEqualFunc configures the equality check used to suppress redundant updates.
func (c *Computed[T]) SetEqualFunc(fn EqualFunc[T]) {
	if c == nil {
		return
	}
	c.signal.SetEqualFunc(fn)
}

// Get returns the current computed value.
func (c *Computed[T]) Get() T {
	if c == nil {
		var zero T
		return zero
	}
	return c.signal.Get()
}

// Subscribe registers a listener for change notifications.
func (c *Computed[T]) Subscribe(fn func()) func() {
	if c == nil {
		return func() {}
	}
	return c.signal.Subscribe(fn)
}

// SubscribeWithScheduler registers a listener using a scheduler.
// If scheduler is nil, callbacks run synchronously.
func (c *Computed[T]) SubscribeWithScheduler(scheduler Scheduler, fn func()) func() {
	if c == nil {
		return func() {}
	}
	return c.signal.SubscribeWithScheduler(scheduler, fn)
}

// Stop unsubscribes from dependency updates.
func (c *Computed[T]) Stop() {
	if c == nil {
		return
	}
	c.clearDeps()
}

func (c *Computed[T]) recompute() {
	if c == nil {
		return
	}
	if c.autoDeps {
		value, deps := trackDependencies(c.compute)
		c.replaceDeps(deps)
		c.signal.Set(value)
		return
	}
	c.signal.Set(c.compute())
}

func (c *Computed[T]) enqueueRecompute() {
	if c == nil {
		return
	}
	if c.scheduler == nil {
		c.recompute()
		return
	}
	c.scheduler.Schedule(c.recompute)
}

func (c *Computed[T]) setDeps(deps []Subscribable) {
	if c == nil {
		return
	}
	unsubs := make([]func(), 0, len(deps))
	for _, dep := range deps {
		if dep == nil {
			continue
		}
		unsub := dep.Subscribe(c.enqueueRecompute)
		if unsub != nil {
			unsubs = append(unsubs, unsub)
		}
	}
	c.mu.Lock()
	c.unsubs = unsubs
	c.mu.Unlock()
}

func (c *Computed[T]) clearDeps() {
	if c == nil {
		return
	}
	c.mu.Lock()
	unsubs := c.unsubs
	c.unsubs = nil
	c.mu.Unlock()
	for _, unsub := range unsubs {
		if unsub != nil {
			unsub()
		}
	}
}

func (c *Computed[T]) replaceDeps(deps []Subscribable) {
	c.clearDeps()
	c.setDeps(deps)
}
