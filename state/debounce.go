package state

import (
	"sync"
	"time"
)

// Debounce returns a function that delays calling fn until after the specified
// duration has elapsed since the last invocation. Each call resets the timer.
// Useful for search-as-you-type, resize handlers, and similar scenarios where
// only the final event in a burst matters.
func Debounce(duration time.Duration, fn func()) func() {
	var mu sync.Mutex
	var timer *time.Timer

	return func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(duration, fn)
	}
}

// Throttle returns a function that calls fn at most once per duration.
// The first call executes immediately; subsequent calls within the duration
// window are silently dropped. After the duration elapses, the next call
// will execute immediately again.
func Throttle(duration time.Duration, fn func()) func() {
	var mu sync.Mutex
	var lastCall time.Time

	return func() {
		mu.Lock()
		defer mu.Unlock()
		now := time.Now()
		if now.Sub(lastCall) >= duration {
			lastCall = now
			fn()
		}
	}
}

// DebouncedSignal creates a new signal that follows the source signal but
// only updates after the source value has been stable for the given delay.
// This is useful for debouncing user input before triggering expensive
// operations like search queries or validation.
//
// The returned signal starts with the current value of source.
func DebouncedSignal[T any](source *Signal[T], delay time.Duration) *Signal[T] {
	debounced := NewSignal(source.Get())
	var mu sync.Mutex
	var timer *time.Timer

	source.Subscribe(func() {
		value := source.Get()
		mu.Lock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(delay, func() {
			debounced.Set(value)
		})
		mu.Unlock()
	})

	return debounced
}
