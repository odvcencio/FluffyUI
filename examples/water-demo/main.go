// Water Demo - GPU-accelerated waterfall with particles and mist
//
// Demonstrates GPU canvas with particle physics for a waterfall effect
//
// Usage:
//
//	go run ./examples/water-demo
package main

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/gpu"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	view := NewWaterDemo()
	bundle, err := demo.NewApp(view, demo.Options{TickRate: time.Second / 60})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

// WaterParticle represents a single water droplet
type WaterParticle struct {
	X, Y       float32
	VX, VY     float32
	Life       float32
	MaxLife    float32
	Size       float32
	Alpha      float32
	IsMist     bool
	IsSplash   bool
}

// WaterDemo renders a GPU-accelerated waterfall
type WaterDemo struct {
	*widgets.GPUCanvasWidget
	particles  []WaterParticle
	phase      float32
	lastEmit   time.Time
	poolLevel  float32
	rocks      []Rock
}

// Rock represents a decorative rock in the scene
type Rock struct {
	X, Y, W, H float32
}

func NewWaterDemo() *WaterDemo {
	d := &WaterDemo{
		particles: make([]WaterParticle, 0, 2000),
		poolLevel: 0.85, // Pool starts at 85% down the screen
	}
	d.GPUCanvasWidget = widgets.NewGPUCanvasWidget(d.draw)

	// Check for GPU backend override
	backend := strings.TrimSpace(os.Getenv("FLUFFYUI_GPU_BACKEND"))
	switch strings.ToLower(backend) {
	case "opengl", "gl":
		d.GPUCanvasWidget.WithBackend(gpu.BackendOpenGL)
	case "metal":
		d.GPUCanvasWidget.WithBackend(gpu.BackendMetal)
	case "software", "cpu":
		d.GPUCanvasWidget.WithBackend(gpu.BackendSoftware)
	default:
		// Auto-detect: OpenGL on Linux/Windows, Metal on macOS, software fallback
		d.GPUCanvasWidget.WithBackend(gpu.BackendAuto)
	}
	return d
}

func (d *WaterDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.TickMsg:
		d.update(m.Time)
		d.Invalidate()
		return runtime.Handled()
	case runtime.KeyMsg:
		if m.Rune == 'q' {
			return runtime.WithCommand(runtime.Quit{})
		}
	}
	return runtime.Unhandled()
}

func (d *WaterDemo) update(now time.Time) {
	dt := float32(1.0 / 60.0)
	d.phase += 0.02
	if d.phase > math.Pi*2 {
		d.phase -= math.Pi * 2
	}

	bounds := d.Bounds()
	w := float32(bounds.Width * 2) // Approximate pixel width
	h := float32(bounds.Height * 4)
	if w <= 0 || h <= 0 {
		return
	}

	// Initialize rocks on first frame
	if len(d.rocks) == 0 {
		d.initRocks(w, h)
	}

	// Emit new water particles from the top
	if now.Sub(d.lastEmit) > time.Millisecond*8 {
		d.emitWater(w, h)
		d.lastEmit = now
	}

	poolY := h * d.poolLevel

	// Update particles
	alive := d.particles[:0]
	for i := range d.particles {
		p := &d.particles[i]

		if p.IsMist {
			// Mist rises and drifts
			p.VY -= 15 * dt // Rise
			p.VX += float32(math.Sin(float64(d.phase+p.X*0.1))) * 20 * dt
			p.VY *= 0.98
			p.VX *= 0.98
		} else if p.IsSplash {
			// Splash arcs outward
			p.VY += 180 * dt // Gravity
			p.VX *= 0.99
		} else {
			// Main waterfall - gravity + turbulence
			p.VY += 400 * dt // Strong gravity
			// Turbulence
			turbX := float32(math.Sin(float64(p.Y*0.05+d.phase*2))) * 30
			p.VX += turbX * dt
			p.VX *= 0.95 // Damping
		}

		// Update position
		p.X += p.VX * dt
		p.Y += p.VY * dt

		// Age particle
		p.Life -= dt
		p.Alpha = p.Life / p.MaxLife

		// Check for pool collision (main water only)
		if !p.IsMist && !p.IsSplash && p.Y >= poolY {
			// Spawn splash particles
			d.spawnSplash(p.X, poolY)
			// Spawn mist
			if rand.Float32() < 0.3 {
				d.spawnMist(p.X, poolY)
			}
			p.Life = 0 // Kill this particle
		}

		// Keep alive particles
		if p.Life > 0 && p.Y < h+20 && p.Y > -20 && p.X > -20 && p.X < w+20 {
			alive = append(alive, *p)
		}
	}
	d.particles = alive
}

func (d *WaterDemo) initRocks(w, h float32) {
	poolY := h * d.poolLevel
	// Rocks at the base of the waterfall
	d.rocks = []Rock{
		{X: w*0.35 - 15, Y: poolY - 8, W: 30, H: 16},
		{X: w*0.5 - 20, Y: poolY - 5, W: 40, H: 12},
		{X: w*0.65 - 12, Y: poolY - 6, W: 24, H: 14},
	}
}

func (d *WaterDemo) emitWater(w, h float32) {
	// Waterfall source - top center
	sourceX := w * 0.5
	sourceW := w * 0.15

	for i := 0; i < 8; i++ {
		x := sourceX + (rand.Float32()-0.5)*sourceW
		d.particles = append(d.particles, WaterParticle{
			X:       x,
			Y:       0,
			VX:      (rand.Float32() - 0.5) * 20,
			VY:      rand.Float32() * 50,
			Life:    2.0 + rand.Float32()*1.0,
			MaxLife: 3.0,
			Size:    1 + rand.Float32()*2,
			Alpha:   1.0,
		})
	}
}

func (d *WaterDemo) spawnSplash(x, y float32) {
	count := 2 + rand.Intn(3)
	for i := 0; i < count; i++ {
		angle := -math.Pi/2 + (rand.Float64()-0.5)*math.Pi*0.8
		speed := 60 + rand.Float32()*80
		d.particles = append(d.particles, WaterParticle{
			X:        x + (rand.Float32()-0.5)*10,
			Y:        y,
			VX:       float32(math.Cos(angle)) * speed,
			VY:       float32(math.Sin(angle)) * speed,
			Life:     0.3 + rand.Float32()*0.4,
			MaxLife:  0.7,
			Size:     1 + rand.Float32(),
			Alpha:    0.8,
			IsSplash: true,
		})
	}
}

func (d *WaterDemo) spawnMist(x, y float32) {
	count := 1 + rand.Intn(2)
	for i := 0; i < count; i++ {
		d.particles = append(d.particles, WaterParticle{
			X:       x + (rand.Float32()-0.5)*30,
			Y:       y - rand.Float32()*10,
			VX:      (rand.Float32() - 0.5) * 40,
			VY:      -20 - rand.Float32()*30,
			Life:    1.0 + rand.Float32()*1.5,
			MaxLife: 2.5,
			Size:    2 + rand.Float32()*3,
			Alpha:   0.4,
			IsMist:  true,
		})
	}
}

func (d *WaterDemo) draw(canvas *gpu.GPUCanvas) {
	w, h := canvas.Size()
	if w <= 0 || h <= 0 {
		return
	}

	fw, fh := float32(w), float32(h)
	poolY := fh * d.poolLevel

	// Draw sky gradient (dark blue to lighter blue)
	d.drawBackground(canvas, w, h)

	// Draw cliff/rock face on sides
	d.drawCliffs(canvas, fw, fh)

	// Draw pool water
	d.drawPool(canvas, fw, fh, poolY)

	// Draw rocks
	d.drawRocks(canvas)

	// Draw water particles with glow
	d.drawParticles(canvas, fw, fh)

	// Draw waterfall source (foam at top)
	d.drawSource(canvas, fw)

	// Instructions
	canvas.SetFillColor(color.RGBA{R: 200, G: 220, B: 255, A: 200})
	canvas.DrawText("Q:QUIT", 4, fh-12, nil)

	// Add subtle vignette
	canvas.ApplyEffect(gpu.VignetteEffect{Radius: 0.8, Softness: 0.3})
}

func (d *WaterDemo) drawBackground(canvas *gpu.GPUCanvas, w, h int) {
	// Night sky gradient
	for y := 0; y < h; y++ {
		t := float32(y) / float32(h)
		col := lerpColor(
			color.RGBA{R: 8, G: 15, B: 35, A: 255},   // Dark blue top
			color.RGBA{R: 20, G: 40, B: 60, A: 255},  // Lighter blue bottom
			t,
		)
		canvas.SetStrokeColor(col)
		canvas.DrawLine(0, float32(y), float32(w), float32(y))
	}
}

func (d *WaterDemo) drawCliffs(canvas *gpu.GPUCanvas, w, h float32) {
	cliffColor := color.RGBA{R: 40, G: 35, B: 45, A: 255}
	cliffHighlight := color.RGBA{R: 55, G: 50, B: 60, A: 255}

	// Left cliff
	canvas.SetFillColor(cliffColor)
	canvas.FillRect(0, 0, w*0.38, h)
	canvas.SetFillColor(cliffHighlight)
	canvas.FillRect(w*0.36, 0, w*0.02, h)

	// Right cliff
	canvas.SetFillColor(cliffColor)
	canvas.FillRect(w*0.62, 0, w*0.38, h)
	canvas.SetFillColor(cliffHighlight)
	canvas.FillRect(w*0.62, 0, w*0.02, h)
}

func (d *WaterDemo) drawPool(canvas *gpu.GPUCanvas, w, h, poolY float32) {
	// Pool base color
	poolColor := color.RGBA{R: 20, G: 50, B: 80, A: 255}
	canvas.SetFillColor(poolColor)
	canvas.FillRect(w*0.3, poolY, w*0.4, h-poolY)

	// Animated surface waves
	canvas.SetStrokeColor(color.RGBA{R: 60, G: 120, B: 180, A: 150})
	canvas.SetStrokeWidth(2)
	for i := 0; i < 3; i++ {
		waveY := poolY + float32(i)*8 + float32(math.Sin(float64(d.phase)+float64(i)*0.5))*3
		canvas.DrawLine(w*0.3, waveY, w*0.7, waveY)
	}

	// Foam where water hits
	canvas.PushLayer()
	foamColor := color.RGBA{R: 180, G: 200, B: 220, A: 180}
	canvas.SetFillColor(foamColor)
	foamX := w * 0.5
	foamW := w*0.1 + float32(math.Sin(float64(d.phase*2)))*w*0.02
	canvas.FillCircle(foamX, poolY+5, foamW)
	canvas.PopLayer(gpu.GlowEffect{Radius: 8, Intensity: 0.4, Color: color.RGBA{R: 150, G: 180, B: 220, A: 100}})
}

func (d *WaterDemo) drawRocks(canvas *gpu.GPUCanvas) {
	rockColor := color.RGBA{R: 50, G: 45, B: 55, A: 255}
	rockHighlight := color.RGBA{R: 70, G: 65, B: 75, A: 255}

	for _, rock := range d.rocks {
		canvas.SetFillColor(rockColor)
		canvas.FillRoundedRect(rock.X, rock.Y, rock.W, rock.H, 4)
		// Highlight
		canvas.SetFillColor(rockHighlight)
		canvas.FillRoundedRect(rock.X+2, rock.Y+2, rock.W*0.6, rock.H*0.4, 2)
	}
}

func (d *WaterDemo) drawParticles(canvas *gpu.GPUCanvas, w, h float32) {
	// Scale from particle coords to canvas coords
	scaleX := w / (float32(d.Bounds().Width) * 2)
	scaleY := h / (float32(d.Bounds().Height) * 4)

	// Draw mist first (behind water)
	for _, p := range d.particles {
		if !p.IsMist {
			continue
		}
		x := p.X * scaleX
		y := p.Y * scaleY
		alpha := uint8(p.Alpha * 100)
		mistColor := color.RGBA{R: 180, G: 200, B: 230, A: alpha}
		canvas.SetFillColor(mistColor)
		canvas.FillCircle(x, y, p.Size*scaleX*1.5)
	}

	// Draw main water particles
	canvas.PushLayer()
	for _, p := range d.particles {
		if p.IsMist || p.IsSplash {
			continue
		}
		x := p.X * scaleX
		y := p.Y * scaleY
		alpha := uint8(p.Alpha * 200)
		waterColor := color.RGBA{R: 100, G: 150, B: 220, A: alpha}
		canvas.SetFillColor(waterColor)
		canvas.FillCircle(x, y, p.Size*scaleX)
	}
	canvas.PopLayer(gpu.GlowEffect{Radius: 4, Intensity: 0.3, Color: color.RGBA{R: 80, G: 140, B: 200, A: 150}})

	// Draw splash particles
	for _, p := range d.particles {
		if !p.IsSplash {
			continue
		}
		x := p.X * scaleX
		y := p.Y * scaleY
		alpha := uint8(p.Alpha * 220)
		splashColor := color.RGBA{R: 180, G: 210, B: 240, A: alpha}
		canvas.SetFillColor(splashColor)
		canvas.FillCircle(x, y, p.Size*scaleX)
	}
}

func (d *WaterDemo) drawSource(canvas *gpu.GPUCanvas, w float32) {
	// Foam/turbulence at the waterfall source
	canvas.PushLayer()
	foamColor := color.RGBA{R: 200, G: 220, B: 240, A: 200}
	canvas.SetFillColor(foamColor)
	sourceX := w * 0.5
	sourceW := w * 0.08

	for i := 0; i < 5; i++ {
		offset := float32(math.Sin(float64(d.phase*3)+float64(i))) * sourceW * 0.3
		canvas.FillCircle(sourceX+offset, 8+float32(i)*3, 4+float32(i))
	}
	canvas.PopLayer(gpu.GlowEffect{Radius: 6, Intensity: 0.5, Color: color.RGBA{R: 150, G: 180, B: 220, A: 150}})
}

func lerpColor(a, b color.RGBA, t float32) color.RGBA {
	return color.RGBA{
		R: uint8(float32(a.R) + (float32(b.R)-float32(a.R))*t),
		G: uint8(float32(a.G) + (float32(b.G)-float32(a.G))*t),
		B: uint8(float32(a.B) + (float32(b.B)-float32(a.B))*t),
		A: uint8(float32(a.A) + (float32(b.A)-float32(a.A))*t),
	}
}
