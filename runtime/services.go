package runtime

import (
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/animation"
	"github.com/odvcencio/fluffy-ui/audio"
	"github.com/odvcencio/fluffy-ui/clipboard"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/style"
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

// Animator returns the app animator.
func (s Services) Animator() *animation.Animator {
	if s.app == nil {
		return nil
	}
	return s.app.animator
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
