package widgets

import (
	"time"

	"github.com/odvcencio/fluffy-ui/animation"
)

// Direction indicates a slide direction.
type Direction int

const (
	DirectionLeft Direction = iota
	DirectionRight
	DirectionUp
	DirectionDown
)

// AnimatedWidget is a base for widgets with animation support.
type AnimatedWidget struct {
	Component

	Opacity animation.Float64
	OffsetX animation.Float64
	OffsetY animation.Float64
	Scale   animation.Float64
}

// NewAnimatedWidget creates an AnimatedWidget with defaults.
func NewAnimatedWidget() AnimatedWidget {
	return AnimatedWidget{
		Opacity: animation.Float64(1),
		Scale:   animation.Float64(1),
	}
}

// Animate starts a property animation.
func (w *AnimatedWidget) Animate(property string, endValue animation.Animatable, cfg animation.TweenConfig) {
	if w == nil {
		return
	}
	animator := w.Services.Animator()
	if animator == nil {
		return
	}

	var getValue func() animation.Animatable
	var setValue func(animation.Animatable)

	switch property {
	case "Opacity":
		getValue = func() animation.Animatable { return w.Opacity }
		setValue = func(v animation.Animatable) {
			w.Opacity = v.(animation.Float64)
			w.Invalidate()
		}
	case "OffsetX":
		getValue = func() animation.Animatable { return w.OffsetX }
		setValue = func(v animation.Animatable) {
			w.OffsetX = v.(animation.Float64)
			w.Invalidate()
		}
	case "OffsetY":
		getValue = func() animation.Animatable { return w.OffsetY }
		setValue = func(v animation.Animatable) {
			w.OffsetY = v.(animation.Float64)
			w.Invalidate()
		}
	case "Scale":
		getValue = func() animation.Animatable { return w.Scale }
		setValue = func(v animation.Animatable) {
			w.Scale = v.(animation.Float64)
			w.Invalidate()
		}
	default:
		return
	}

	animator.Animate(w, property, getValue, setValue, endValue, cfg)
}

// FadeIn animates opacity from 0 to 1.
func (w *AnimatedWidget) FadeIn(duration time.Duration) {
	if w == nil {
		return
	}
	w.Opacity = 0
	w.Animate("Opacity", animation.Float64(1), animation.TweenConfig{
		Duration: duration,
		Easing:   animation.OutCubic,
	})
}

// FadeOut animates opacity to 0.
func (w *AnimatedWidget) FadeOut(duration time.Duration, onComplete func()) {
	if w == nil {
		return
	}
	w.Animate("Opacity", animation.Float64(0), animation.TweenConfig{
		Duration:   duration,
		Easing:     animation.InCubic,
		OnComplete: onComplete,
	})
}

// SlideIn animates the widget into place from a direction.
func (w *AnimatedWidget) SlideIn(from Direction, distance int, duration time.Duration) {
	if w == nil {
		return
	}
	switch from {
	case DirectionLeft:
		w.OffsetX = animation.Float64(-distance)
		w.Animate("OffsetX", animation.Float64(0), animation.TweenConfig{Duration: duration, Easing: animation.OutCubic})
	case DirectionRight:
		w.OffsetX = animation.Float64(distance)
		w.Animate("OffsetX", animation.Float64(0), animation.TweenConfig{Duration: duration, Easing: animation.OutCubic})
	case DirectionUp:
		w.OffsetY = animation.Float64(-distance)
		w.Animate("OffsetY", animation.Float64(0), animation.TweenConfig{Duration: duration, Easing: animation.OutCubic})
	case DirectionDown:
		w.OffsetY = animation.Float64(distance)
		w.Animate("OffsetY", animation.Float64(0), animation.TweenConfig{Duration: duration, Easing: animation.OutCubic})
	}
}
