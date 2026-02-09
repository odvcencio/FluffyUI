package runtime

import "github.com/odvcencio/fluffyui/state"

// AutoBind observes multiple signals and invalidates the widget when any change.
// Nil signals in the variadic list are silently skipped.
// Returns a cleanup function that unsubscribes all observers; call it in Unbind.
//
// Usage:
//
//	func (w *MyWidget) Bind(s runtime.Services) {
//	    w.services = s
//	    w.cleanup = runtime.AutoBind(s, w.label, w.count, w.enabled)
//	}
//
//	func (w *MyWidget) Unbind() {
//	    w.cleanup()
//	    w.services = runtime.Services{}
//	}
func AutoBind(services Services, signals ...state.Subscribable) func() {
	var unsubs []func()
	for _, sig := range signals {
		if sig == nil {
			continue
		}
		unsub := sig.Subscribe(func() {
			services.Invalidate()
		})
		unsubs = append(unsubs, unsub)
	}
	return func() {
		for _, unsub := range unsubs {
			unsub()
		}
	}
}

// AutoBindRelayout is like AutoBind but triggers a full relayout instead of
// just a render invalidation. Use this for signals that affect widget size
// (e.g., visibility toggles, text content that changes dimensions).
func AutoBindRelayout(services Services, signals ...state.Subscribable) func() {
	var unsubs []func()
	for _, sig := range signals {
		if sig == nil {
			continue
		}
		unsub := sig.Subscribe(func() {
			services.Relayout()
		})
		unsubs = append(unsubs, unsub)
	}
	return func() {
		for _, unsub := range unsubs {
			unsub()
		}
	}
}
