package runtime

import (
	"time"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/animation"
	"m31labs.dev/fluffyui/audio"
	"m31labs.dev/fluffyui/clipboard"
	"m31labs.dev/fluffyui/i18n"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/style"
	"m31labs.dev/fluffyui/theme"
)

// Services exposes app-level scheduling and messaging helpers.
type Services struct {
	app *App
}

// Services returns a service handle for the app.
func (a *App) Services() Services {
	return Services{app: a}
}

func (s Services) isZero() bool {
	return s.app == nil
}

// Announcer returns the accessibility announcer.
func (s Services) Announcer() accessibility.Announcer {
	if s.app == nil {
		return nil
	}
	return s.app.announcer
}

// FocusStyle returns the global focus style.
func (s Services) FocusStyle() *accessibility.FocusStyle {
	if s.app == nil {
		return nil
	}
	return s.app.focusStyle
}

// Clipboard returns the app clipboard.
func (s Services) Clipboard() clipboard.Clipboard {
	if s.app == nil {
		return nil
	}
	return s.app.clipboard
}

// Audio returns the app audio service.
func (s Services) Audio() audio.Service {
	if s.app == nil {
		return nil
	}
	return s.app.audio
}

// Stylesheet returns the active stylesheet.
func (s Services) Stylesheet() *style.Stylesheet {
	if s.app == nil {
		return nil
	}
	return s.app.stylesheet
}

// Theme returns the active theme, if set.
func (s Services) Theme() *theme.Theme {
	if s.app == nil {
		return nil
	}
	return s.app.theme
}

// Localizer returns the active localizer.
func (s Services) Localizer() i18n.Localizer {
	if s.app == nil {
		return nil
	}
	return s.app.localizer
}

// Animator returns the app animator.
func (s Services) Animator() *animation.Animator {
	if s.app == nil {
		return nil
	}
	return s.app.animator
}

// ReducedMotion reports whether motion should be minimized.
func (s Services) ReducedMotion() bool {
	if s.app == nil {
		return false
	}
	return s.app.reducedMotion
}

// SaveFocus pushes the currently focused widget onto the focus restoration stack.
// Call this before opening a modal dialog, popover, or dropdown so that focus
// can be returned to the triggering element when the overlay closes.
func (s Services) SaveFocus() {
	if s.app == nil {
		return
	}
	screen := s.app.Screen()
	if screen == nil {
		return
	}
	// Save the focused widget from the base layer, since overlay layers
	// get their own focus scopes.
	scope := screen.BaseFocusScope()
	if scope == nil {
		return
	}
	s.app.focusRestore.Push(scope.Current())
}

// RestoreFocus pops the most recently saved widget from the focus restoration
// stack and refocuses it. Call this when a modal dialog, popover, or dropdown
// closes to return focus to the element that triggered it.
func (s Services) RestoreFocus() {
	if s.app == nil {
		return
	}
	saved := s.app.focusRestore.Pop()
	if saved == nil {
		return
	}
	screen := s.app.Screen()
	if screen == nil {
		return
	}
	scope := screen.BaseFocusScope()
	if scope == nil {
		return
	}
	scope.SetFocus(saved)
}

// Scheduler returns the app state scheduler.
func (s Services) Scheduler() state.Scheduler {
	if s.app == nil {
		return nil
	}
	return s.app.StateScheduler()
}

// InvalidateScheduler returns the app invalidation scheduler.
func (s Services) InvalidateScheduler() state.Scheduler {
	if s.app == nil {
		return nil
	}
	return s.app.InvalidateScheduler()
}

// Invalidate requests a render pass.
func (s Services) Invalidate() {
	if s.app == nil {
		return
	}
	s.app.Invalidate()
}

// Relayout requests a layout pass followed by a render.
func (s Services) Relayout() {
	if s.app == nil {
		return
	}
	s.app.Relayout()
}

// Post sends a message into the app loop.
func (s Services) Post(msg Message) bool {
	if s.app == nil {
		return false
	}
	return s.app.tryPost(msg)
}

// Spawn starts an effect using the app task context.
func (s Services) Spawn(effect Effect) {
	if s.app == nil {
		return
	}
	s.app.Spawn(effect)
}

// After schedules a delayed message.
func (s Services) After(delay time.Duration, msg Message) {
	if s.app == nil {
		return
	}
	s.app.After(delay, msg)
}

// Every schedules a recurring message.
func (s Services) Every(interval time.Duration, fn func(time.Time) Message) {
	if s.app == nil {
		return
	}
	s.app.Every(interval, fn)
}
