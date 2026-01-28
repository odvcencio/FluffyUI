package state

import "sync"

// Effect runs side effects when dependencies change.
type Effect func()

// Signalish describes a reactive dependency.
type Signalish = Subscribable

// EffectHandle manages an effect's lifecycle.
type EffectHandle struct {
	fn   Effect
	subs Subscriptions
	mu   sync.Mutex
	dead bool
}

// NewEffect creates an effect that runs when deps change.
func NewEffect(fn Effect, deps ...Signalish) *EffectHandle {
	return NewEffectWithScheduler(nil, fn, deps...)
}

// NewEffectWithScheduler creates an effect with a scheduler.
func NewEffectWithScheduler(scheduler Scheduler, fn Effect, deps ...Signalish) *EffectHandle {
	if fn == nil {
		fn = func() {}
	}
	e := &EffectHandle{fn: fn}
	e.subs.SetScheduler(scheduler)
	for _, dep := range deps {
		if dep == nil {
			continue
		}
		e.subs.Observe(dep, e.run)
	}
	e.Trigger()
	return e
}

// Trigger runs the effect once.
func (e *EffectHandle) Trigger() {
	if e == nil {
		return
	}
	scheduler := e.subs.Scheduler()
	if scheduler != nil {
		scheduler.Schedule(e.run)
		return
	}
	e.run()
}

// Dispose stops the effect and unsubscribes from dependencies.
func (e *EffectHandle) Dispose() {
	if e == nil {
		return
	}
	e.mu.Lock()
	if e.dead {
		e.mu.Unlock()
		return
	}
	e.dead = true
	e.mu.Unlock()
	e.subs.Clear()
}

func (e *EffectHandle) run() {
	if e == nil {
		return
	}
	e.mu.Lock()
	if e.dead {
		e.mu.Unlock()
		return
	}
	fn := e.fn
	e.mu.Unlock()
	if fn != nil {
		fn()
	}
}
