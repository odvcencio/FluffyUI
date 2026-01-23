package widgets

import (
	"math"

	"github.com/odvcencio/fluffy-ui/animation"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/effects"
	"github.com/odvcencio/fluffy-ui/graphics"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// GaugeColors defines the colors used for the animated gauge.
type GaugeColors struct {
	Background backend.Color
	Fill       backend.Color
	Glow       backend.Color
}

// AnimatedGauge renders a radial gauge with spring animation.
type AnimatedGauge struct {
	CanvasWidget

	value    float64
	min, max float64
	spring   *animation.Spring
	colors   GaugeColors
	services runtime.Services
}

// NewAnimatedGauge creates a new animated gauge.
func NewAnimatedGauge(minValue, maxValue float64) *AnimatedGauge {
	g := &AnimatedGauge{
		min: minValue,
		max: maxValue,
		colors: GaugeColors{
			Background: backend.ColorRGB(40, 40, 40),
			Fill:       backend.ColorRGB(0, 200, 100),
			Glow:       backend.ColorRGB(0, 255, 150),
		},
	}
	cfg := animation.SpringDefault
	cfg.OnUpdate = func(value float64) {
		g.Invalidate()
	}
	g.spring = animation.NewSpring(0, cfg)
	g.CanvasWidget = *NewCanvasWidget(g.drawGauge)
	return g
}

// StyleType returns the selector type name.
func (g *AnimatedGauge) StyleType() string { return "AnimatedGauge" }

// Bind attaches services and registers the spring.
func (g *AnimatedGauge) Bind(services runtime.Services) {
	if g == nil {
		return
	}
	g.services = services
	g.CanvasWidget.Bind(services)
	if animator := services.Animator(); animator != nil {
		animator.AnimateSpring(g, "value", g.spring, g.spring.Target)
	}
}

// Unbind releases services.
func (g *AnimatedGauge) Unbind() {
	if g == nil {
		return
	}
	g.services = runtime.Services{}
	g.CanvasWidget.Unbind()
}

// SetValue updates the gauge target value.
func (g *AnimatedGauge) SetValue(value float64) {
	if g == nil || g.spring == nil {
		return
	}
	g.value = value
	rangeSpan := g.max - g.min
	if rangeSpan == 0 {
		return
	}
	ratio := (value - g.min) / rangeSpan
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	if animator := g.services.Animator(); animator != nil {
		animator.AnimateSpring(g, "value", g.spring, ratio)
	} else {
		g.spring.SetTarget(ratio)
	}
	g.Invalidate()
}

func (g *AnimatedGauge) drawGauge(canvas *graphics.Canvas) {
	if g == nil || canvas == nil || g.spring == nil {
		return
	}
	w, h := canvas.Size()
	if w <= 0 || h <= 0 {
		return
	}
	cx, cy := w/2, h/2
	radius := minInt(w, h)/2 - 3
	if radius <= 0 {
		return
	}
	start := math.Pi * 0.75
	end := math.Pi * 2.25

	canvas.SetStrokeColor(g.colors.Background)
	canvas.DrawArc(cx, cy, radius, start, end)

	progress := g.spring.Value
	if progress <= 0 {
		return
	}
	angle := start + progress*(end-start)
	canvas.SetStrokeColor(g.colors.Fill)
	canvas.DrawArc(cx, cy, radius, start, angle)

	endX := cx + int(math.Round(float64(radius)*math.Cos(angle)))
	endY := cy + int(math.Round(float64(radius)*math.Sin(angle)))
	effects.Glow(canvas, endX, endY, 3, g.colors.Glow, 0.5)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
