package state

import (
	"sync"
	"sync/atomic"
)

// Resource represents asynchronously loaded data with status.
type Resource[T any] struct {
	Data    T
	Loading bool
	Error   error

	mu          sync.Mutex
	subs        map[int]subscriber
	next        int
	fetcher     func() (T, error)
	subscribers Subscriptions
	fetchID     uint64
}

// NewResource creates a resource that refetches when deps change.
func NewResource[T any](fetcher func() (T, error), deps ...Signalish) *Resource[T] {
	return NewResourceWithScheduler(nil, fetcher, deps...)
}

// NewResourceWithScheduler creates a resource with dependency scheduling.
func NewResourceWithScheduler[T any](scheduler Scheduler, fetcher func() (T, error), deps ...Signalish) *Resource[T] {
	r := &Resource[T]{
		fetcher: fetcher,
	}
	r.subscribers.SetScheduler(scheduler)
	for _, dep := range deps {
		if dep == nil {
			continue
		}
		r.subscribers.Observe(dep, r.Refetch)
	}
	r.Refetch()
	return r
}

// Get returns a snapshot of the resource state.
func (r *Resource[T]) Get() Resource[T] {
	if r == nil {
		var zero T
		return Resource[T]{Data: zero}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return Resource[T]{
		Data:    r.Data,
		Loading: r.Loading,
		Error:   r.Error,
	}
}

// Subscribe registers a listener for state changes.
func (r *Resource[T]) Subscribe(fn func()) func() {
	return r.SubscribeWithScheduler(nil, fn)
}

// SubscribeWithScheduler registers a listener using a scheduler.
func (r *Resource[T]) SubscribeWithScheduler(scheduler Scheduler, fn func()) func() {
	if r == nil || fn == nil {
		return func() {}
	}
	r.mu.Lock()
	if r.subs == nil {
		r.subs = make(map[int]subscriber)
	}
	id := r.next
	r.next++
	r.subs[id] = subscriber{fn: fn, scheduler: scheduler}
	r.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			r.mu.Lock()
			delete(r.subs, id)
			r.mu.Unlock()
		})
	}
}

// Refetch triggers a new fetch cycle.
func (r *Resource[T]) Refetch() {
	if r == nil || r.fetcher == nil {
		return
	}
	id := atomic.AddUint64(&r.fetchID, 1)
	var subs []subscriber
	r.mu.Lock()
	r.Loading = true
	r.Error = nil
	subs = r.copySubscribersLocked()
	r.mu.Unlock()
	r.notify(subs)

	go func(fetchID uint64) {
		data, err := r.fetcher()
		r.mu.Lock()
		if fetchID != r.fetchID {
			r.mu.Unlock()
			return
		}
		r.Data = data
		r.Error = err
		r.Loading = false
		subs = r.copySubscribersLocked()
		r.mu.Unlock()
		r.notify(subs)
	}(id)
}

// Dispose unsubscribes from dependencies.
func (r *Resource[T]) Dispose() {
	if r == nil {
		return
	}
	r.subscribers.Clear()
}

func (r *Resource[T]) copySubscribersLocked() []subscriber {
	if len(r.subs) == 0 {
		return nil
	}
	subs := acquireSubscribers(len(r.subs))
	for _, sub := range r.subs {
		subs = append(subs, sub)
	}
	return subs
}

func (r *Resource[T]) notify(subs []subscriber) {
	if len(subs) == 0 {
		return
	}
	if enqueueBatch(subs) {
		releaseSubscribers(subs)
		return
	}
	for _, sub := range subs {
		if sub.fn == nil {
			continue
		}
		if sub.scheduler == nil {
			sub.fn()
			continue
		}
		sub.scheduler.Schedule(sub.fn)
	}
	releaseSubscribers(subs)
}
