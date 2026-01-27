package runtime

import (
	"sync"
	"sync/atomic"
)

// WidgetPool provides pooled widget reuse with optional reset logic.
// Size is best-effort and intended for limiting pool growth.
type WidgetPool[T any] struct {
	pool    sync.Pool
	newFn   func() T
	reset   func(T)
	maxSize int
	size    atomic.Int64
}

// NewWidgetPool creates a new widget pool.
// maxSize <= 0 means no explicit limit.
func NewWidgetPool[T any](newFn func() T, resetFn func(T), maxSize int) *WidgetPool[T] {
	if newFn == nil {
		return &WidgetPool[T]{}
	}
	p := &WidgetPool[T]{
		newFn:   newFn,
		reset:   resetFn,
		maxSize: maxSize,
	}
	p.pool.New = func() any {
		return newFn()
	}
	return p
}

// Acquire retrieves a widget instance, creating one if needed.
func (p *WidgetPool[T]) Acquire() T {
	var zero T
	if p == nil || p.newFn == nil {
		return zero
	}
	value := p.pool.Get()
	item, ok := value.(T)
	if !ok {
		return zero
	}
	if p.maxSize > 0 {
		p.decSize()
	}
	return item
}

// Release returns a widget instance to the pool.
func (p *WidgetPool[T]) Release(widget T) {
	if p == nil || p.newFn == nil {
		return
	}
	if p.reset != nil {
		p.reset(widget)
	}
	if p.maxSize > 0 {
		if !p.incSize() {
			return
		}
	}
	p.pool.Put(widget)
}

// Size returns the approximate number of pooled widgets.
func (p *WidgetPool[T]) Size() int {
	if p == nil {
		return 0
	}
	return int(p.size.Load())
}

// MaxSize returns the configured maximum pool size.
func (p *WidgetPool[T]) MaxSize() int {
	if p == nil {
		return 0
	}
	return p.maxSize
}

func (p *WidgetPool[T]) incSize() bool {
	for {
		current := p.size.Load()
		if p.maxSize > 0 && current >= int64(p.maxSize) {
			return false
		}
		if p.size.CompareAndSwap(current, current+1) {
			return true
		}
	}
}

func (p *WidgetPool[T]) decSize() {
	for {
		current := p.size.Load()
		if current == 0 {
			return
		}
		if p.size.CompareAndSwap(current, current-1) {
			return
		}
	}
}
