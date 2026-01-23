package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/animation"
	"github.com/odvcencio/fluffy-ui/backend"
	backendtcell "github.com/odvcencio/fluffy-ui/backend/tcell"
	"github.com/odvcencio/fluffy-ui/effects"
	"github.com/odvcencio/fluffy-ui/graphics"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	be, err := backendtcell.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "backend init failed: %v\n", err)
		os.Exit(1)
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:  be,
		TickRate: time.Second / 60,
		Animator: animation.NewAnimator(),
	})

	root := NewGraphicsDashboard()
	app.SetRoot(root)

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type GraphicsDemo struct {
	widgets.CanvasWidget
	particles *animation.ParticleSystem
	phase     float64
	lastBurst time.Time
	image     image.Image
}

func NewGraphicsDemo() *GraphicsDemo {
	sample := loadSampleImage()
	if sample == nil {
		sample = buildDemoImage()
	}
	demo := &GraphicsDemo{
		particles: animation.NewParticleSystem(256),
		image:     sample,
	}
	widget := widgets.NewCanvasWidget(demo.draw)
	widget.WithBlitter(graphics.BestBlitter(nil))
	demo.CanvasWidget = *widget
	return demo
}

func (d *GraphicsDemo) Bind(services runtime.Services) {
	d.CanvasWidget.Bind(services)
	if animator := services.Animator(); animator != nil {
		animator.AddParticleSystem(d.particles)
	}
}

func (d *GraphicsDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.TickMsg:
		d.phase += 0.015
		if d.phase > 1 {
			d.phase -= 1
		}
		if m.Time.Sub(d.lastBurst) > time.Second {
			bounds := d.Bounds()
			pixelW := bounds.Width * 2
			pixelH := bounds.Height * 3
			if pixelW > 0 && pixelH > 0 {
				effects.Confetti(d.particles, pixelW/2, pixelH/2, 40)
			}
			d.lastBurst = m.Time
		}
		return runtime.Handled()
	case runtime.KeyMsg:
		if m.Rune == 'q' {
			return runtime.WithCommand(runtime.Quit{})
		}
	}
	return runtime.Unhandled()
}

func (d *GraphicsDemo) draw(canvas *graphics.Canvas) {
	w, h := canvas.Size()
	if w == 0 || h == 0 {
		return
	}
	backgroundStart := backend.ColorRGB(12, 12, 16)
	backgroundEnd := backend.ColorRGB(32, 32, 48)
	effects.LinearGradient(canvas, 0, 0, w, h, backgroundStart, backgroundEnd, math.Pi/6)

	canvas.SetStrokeColor(backend.ColorRGB(240, 238, 232))
	canvas.DrawText(2, 2, "FLUFFY UI", graphics.DefaultFont)

	canvas.SetStrokeColor(backend.ColorRGB(255, 183, 77))
	canvas.DrawLineAA(0, 0, w-1, h-1)

	radius := min(w, h) / 4
	canvas.SetStrokeColor(backend.ColorRGB(79, 195, 247))
	canvas.DrawCircle(w/2, h/2, radius)
	effects.Glow(canvas, w/2, h/2, radius/2, backend.ColorRGB(79, 195, 247), 0.4)

	effects.Shimmer(canvas, 2, 2, max(4, w/3), max(2, h/6), d.phase, backend.ColorRGB(255, 255, 255))

	if d.image != nil {
		imageW := min(32, max(8, w/4))
		imageH := min(16, max(6, h/4))
		imageX := w - imageW - 2
		imageY := 2
		if imageX >= 0 && imageY >= 0 {
			canvas.DrawImageScaled(imageX, imageY, imageW, imageH, d.image)
		}
	}

	d.particles.Render(canvas)
}

type GraphicsDashboard struct {
	layout     *runtime.Flex
	demo       *GraphicsDemo
	chart      *widgets.LineChart
	gauge      *widgets.AnimatedGauge
	values     []float64
	phase      float64
	lastUpdate time.Time
}

func NewGraphicsDashboard() *GraphicsDashboard {
	demo := NewGraphicsDemo()
	chart := widgets.NewLineChart()
	gauge := widgets.NewAnimatedGauge(0, 100)
	gauge.SetValue(50)

	bottom := runtime.HBox(
		runtime.Expanded(chart),
		runtime.Sized(gauge, 24),
	).WithGap(2)

	layout := runtime.VBox(
		runtime.Expanded(demo),
		runtime.Sized(bottom, 12),
	).WithGap(1)

	dashboard := &GraphicsDashboard{
		layout: layout,
		demo:   demo,
		chart:  chart,
		gauge:  gauge,
		values: []float64{50, 55, 60, 58, 52, 48, 46, 50},
		phase:  0,
	}
	dashboard.refreshSeries()
	return dashboard
}

func (d *GraphicsDashboard) Measure(constraints runtime.Constraints) runtime.Size {
	if d == nil || d.layout == nil {
		return constraints.MinSize()
	}
	return d.layout.Measure(constraints)
}

func (d *GraphicsDashboard) Layout(bounds runtime.Rect) {
	if d == nil || d.layout == nil {
		return
	}
	d.layout.Layout(bounds)
}

func (d *GraphicsDashboard) Render(ctx runtime.RenderContext) {
	if d == nil || d.layout == nil {
		return
	}
	d.layout.Render(ctx)
}

func (d *GraphicsDashboard) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.TickMsg:
		d.updateSeries(m.Time)
	}
	if d != nil && d.layout != nil {
		return d.layout.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (d *GraphicsDashboard) ChildWidgets() []runtime.Widget {
	if d == nil || d.layout == nil {
		return nil
	}
	return []runtime.Widget{d.layout}
}

func (d *GraphicsDashboard) updateSeries(now time.Time) {
	if d == nil {
		return
	}
	if !d.lastUpdate.IsZero() && now.Sub(d.lastUpdate) < 120*time.Millisecond {
		return
	}
	d.lastUpdate = now
	d.phase += 0.15
	value := 50 + 50*math.Sin(d.phase)
	d.values = append(d.values, value)
	if len(d.values) > 40 {
		d.values = d.values[len(d.values)-40:]
	}
	d.refreshSeries()
	d.gauge.SetValue(value)
}

func (d *GraphicsDashboard) refreshSeries() {
	if d == nil || d.chart == nil {
		return
	}
	data := append([]float64(nil), d.values...)
	d.chart.SetSeries([]widgets.ChartSeries{{
		Data:   data,
		Color:  backend.ColorRGB(255, 183, 77),
		Smooth: true,
		Fill:   true,
	}})
}

const sampleImageBase64 = `
iVBORw0KGgoAAAANSUhEUgAAACAAAAAQCAIAAAD4YuoOAAAECElEQVR4nDzQQWsa6R8H
8B/xz5wemIuXvxLYiVa2G61sjVa2G61sjVa2GyUHo+QSJQej5JJ9H3sq9GKhp2IvUxkw
qQ6YpzJgUh0wT2XApGsvU+hxTp7EhW8gb+HzP7ec98gvJVr+cNo++YVEy29OJyg/l2hpO
d2o/Eyipen0E/JTiZaGM8jITyRa6s4wLz+WaKk5o5L8SKLle2dSkR9KtHzrTOvyA4mWr5
3Z3/JPa25vnhh9X2jr3j+J0dfFWcCbIUZfFnrY+wcx+ry4iHsTxOjTwkh5fyNGHxeXOW+
UGH1YmHveX4nRu8X1gTdIjN4srCPvz8To1eLmxOsjRmvE6IfT9igvidE3p+NTXhAjy+kG
lefEyHT6UeUZMTKcQUJ5Sox0Z5hRnhAjzRnllcfE6L0zKSmPiNFbZ1pRHhKj186srjwgR
v84/7pIov/7/3JJq++2tuHPuaTVV/t80592Sasvtr7lT7mk1Webb/t/d0mrT7ax44+7pN
VH+2rXH3FJqw+2ue8Pu6TVO1sc+n9xSas3tnXsD7ik1Sv79tSvrLlDILK19RCI7LNACES
2Hg6ByL6Ih0BkG6kQiOzLXAhEtrkXApF9fRACkW0dhUBk35yE7onmbU8ERPOOLwKieTcY
AdG8H42AaD5IREA0H2YiIJqP8hEQzSelCIjm00oERPNZPQKi+R1RDERC24iBSJxvxkAk9
K0YiATfjoFIGDsxEImr3RiIhLkfA5EQhzEQCes4BiJxexpT1txJEAltPQkicRZIgkjo4S
SIxEU8CSJhpJIgEpe5JIiEuZcEkbg+SIJIWEdJEImbk+Q90bjtSYNo3PGlQTTuBtMgGve
jaRCNB4k0iMbDTBpE41E+DaLxpJQG0XhaSYNoPKunQTS+I8qCiGsbWRDx880siLi+lQUR
59tZEHFjJwsifrWbBRE397Mg4uIwCyJuHWdBxG9Ps8qauwAirq0XQMTPAgUQcT1cABG/i
BdAxI1UAUT8MlcAETf3CiDi1wcFEHHrqAAifnNSuCfqtT1FEPU6viKIet1gEUS9frQIot
4gUQRRb5gpgqg3yhdB1JuUiiDqTStFEPVm9SKIendEZRCp2kYZROr5ZhlEqr5VBpHKt8sg
Uo2dMojUq90yiFRzvwwiVRyWQaRax2UQqbenZWXNXQWRqq1XQaSeBaogUvVwFUTqRbwKIt
VIVUGkXuaqIFLNvSqI1OuDKohU66gKIvXmpHpP1Gp7aiBqdXw1ELW6wRqIWv1oDUStQaIG
otYwUwNRa5Svgag1KdVA1JpWaiBqzeo1ELXuiBogamobDRA1zzcbIGrqWw0QNfl2A0RNY6
cBoubVbgNETXO/AaKmOGyAqGkdN0DUvD1tKP8FAAD//6nNQHUW99IiAAAAAElFTkSuQmCC
`

func loadSampleImage() image.Image {
	data, err := decodeBase64Image(sampleImageBase64)
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return img
}

func decodeBase64Image(encoded string) ([]byte, error) {
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t', ' ':
			return -1
		default:
			return r
		}
	}, encoded)
	return base64.StdEncoding.DecodeString(cleaned)
}

func buildDemoImage() image.Image {
	const w, h = 16, 8
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := uint8(40 + x*10)
			g := uint8(30 + y*18)
			b := uint8(180 - x*7)
			if (x+y)%2 == 0 {
				r /= 2
				g /= 2
				b /= 2
			}
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
