// Fireworks Demo - 3D particle effects with perspective projection
//
// Demonstrates the graphics package capabilities inspired by willmcgugan/ny2026
//
// Usage:
//
//	go run ./examples/fireworks-demo
package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/graphics"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	root := NewFireworksDemo()
	bundle, err := demo.NewApp(root, demo.Options{
		TickRate: time.Second / 60,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "backend init failed: %v\n", err)
		os.Exit(1)
	}
	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

// Particle3D represents a particle with 3D position and velocity
type Particle3D struct {
	X, Y, Z       float64 // 3D position
	VX, VY, VZ    float64 // 3D velocity
	R, G, B       uint8   // Color
	Life, MaxLife float64 // Lifetime tracking
}

// Update applies physics to the particle
func (p *Particle3D) Update(dt, gravity, airResistance float64) {
	// Apply gravity (downward in Y)
	p.VY += gravity * dt

	// Apply air resistance (damping)
	p.VX *= airResistance
	p.VY *= airResistance
	p.VZ *= airResistance

	// Update position
	p.X += p.VX * dt
	p.Y += p.VY * dt
	p.Z += p.VZ * dt

	// Age the particle
	p.Life += dt
}

// IsAlive returns true if the particle hasn't expired
func (p *Particle3D) IsAlive() bool {
	return p.Life < p.MaxLife
}

// Project3D converts 3D coordinates to 2D screen coordinates with perspective
func (p *Particle3D) Project3D(cameraZ, cameraDist, centerX, centerY float64) (int, int, bool) {
	// Z relative to camera
	zRel := p.Z - cameraZ
	zOffset := zRel + cameraDist

	if zOffset <= 0 {
		return 0, 0, false // Behind camera
	}

	// Perspective scale
	scale := cameraDist / zOffset

	// Project to screen
	screenX := centerX + (p.X-centerX)*scale
	screenY := centerY + (p.Y-centerY)*scale

	return int(screenX), int(screenY), true
}

// Alpha returns the particle's alpha based on remaining life
func (p *Particle3D) Alpha() float32 {
	remaining := 1.0 - (p.Life / p.MaxLife)
	return float32(remaining)
}

// Firework represents a single firework that launches and explodes
type Firework struct {
	X, Y, Z    float64   // Position
	VX, VY, VZ float64   // Velocity
	R, G, B    uint8     // Color
	Exploded   bool      // Has it exploded?
	Particles  []Particle3D
	Trail      [][2]float64 // Launch trail
	TargetY    float64      // Explosion height
	ApexTime   float64      // Time since reaching apex
}

// NewFirework creates a firework that will launch from the bottom
func NewFirework(canvasW, canvasH int, cameraZ float64) *Firework {
	// Random neon/firework colors
	colors := [][3]uint8{
		{255, 50, 50},    // Red
		{255, 140, 0},    // Orange
		{255, 215, 0},    // Gold
		{50, 255, 50},    // Green
		{100, 150, 255},  // Blue
		{200, 100, 255},  // Purple
		{255, 192, 203},  // Pink
		{0, 255, 255},    // Cyan
		{255, 255, 255},  // White
	}
	c := colors[rand.Intn(len(colors))]

	// Launch position - bottom of screen, random X
	x := float64(canvasW)*0.2 + rand.Float64()*float64(canvasW)*0.6
	y := float64(canvasH) - 1
	z := cameraZ + 50 + rand.Float64()*250

	// Target explosion height (top 10-33% of screen)
	targetY := float64(canvasH)*0.1 + rand.Float64()*float64(canvasH)*0.23

	// Calculate launch velocity to reach target
	gravity := 100.0
	dist := targetY - y // Negative (going up)
	requiredV := math.Sqrt(-2 * gravity * dist)

	return &Firework{
		X: x, Y: y, Z: z,
		VX: rand.Float64()*40 - 20, // Slight horizontal drift
		VY: -requiredV,
		VZ: 0,
		R: c[0], G: c[1], B: c[2],
		TargetY: targetY,
	}
}

// Update advances the firework simulation
func (f *Firework) Update(dt float64) {
	if !f.Exploded {
		// Launch phase
		gravity := 100.0
		f.VY += gravity * dt
		f.X += f.VX * dt
		f.Y += f.VY * dt

		// Store trail
		f.Trail = append(f.Trail, [2]float64{f.X, f.Y})
		if len(f.Trail) > 15 {
			f.Trail = f.Trail[1:]
		}

		// Check for apex (velocity becomes positive/downward)
		if f.VY > 0 {
			f.ApexTime += dt
			if f.ApexTime >= 0.3 {
				f.Explode()
			}
		}
	} else {
		// Update explosion particles
		alive := f.Particles[:0]
		for i := range f.Particles {
			f.Particles[i].Update(dt, 50.0, 0.97)
			if f.Particles[i].IsAlive() {
				alive = append(alive, f.Particles[i])
			}
		}
		f.Particles = alive
	}
}

// Explode creates the particle burst
func (f *Firework) Explode() {
	f.Exploded = true

	// Generate 300-500 particles in all directions (spherical)
	numParticles := 300 + rand.Intn(200)
	speed := 100.0 + rand.Float64()*80

	for i := 0; i < numParticles; i++ {
		// Random direction on a sphere
		theta := rand.Float64() * 2 * math.Pi // Azimuthal angle
		phi := rand.Float64() * math.Pi       // Polar angle

		// Convert spherical to Cartesian velocity
		vx := speed * math.Sin(phi) * math.Cos(theta)
		vy := speed * math.Cos(phi)
		vz := speed * math.Sin(phi) * math.Sin(theta)

		// Random lifetime
		life := 1.5 + rand.Float64()*1.0

		f.Particles = append(f.Particles, Particle3D{
			X: f.X, Y: f.Y, Z: f.Z,
			VX: vx, VY: vy, VZ: vz,
			R: f.R, G: f.G, B: f.B,
			MaxLife: life,
		})
	}
}

// IsFinished returns true when the firework is done
func (f *Firework) IsFinished() bool {
	return f.Exploded && len(f.Particles) == 0
}

// FireworksDemo is the main demo widget
type FireworksDemo struct {
	*widgets.CanvasWidget
	fireworks    []*Firework
	cameraZ      float64
	lastSpawn    time.Time
	spawnDelay   time.Duration
	canvasW      int
	canvasH      int
}

func NewFireworksDemo() *FireworksDemo {
	demo := &FireworksDemo{
		spawnDelay: time.Millisecond * 800,
	}
	// Use Braille blitter for maximum resolution (2x4 pixels per cell)
	demo.CanvasWidget = widgets.NewCanvasWidget(demo.draw)
	demo.CanvasWidget.WithBlitter(&graphics.BrailleBlitter{})
	return demo
}

func (d *FireworksDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.TickMsg:
		dt := 1.0 / 60.0 // Assume 60fps

		// Update canvas size
		bounds := d.Bounds()
		blitter := &graphics.BrailleBlitter{}
		px, py := blitter.PixelsPerCell()
		d.canvasW = bounds.Width * px
		d.canvasH = bounds.Height * py

		// Move camera forward
		d.cameraZ += 15.0 * dt

		// Auto-spawn fireworks
		if m.Time.Sub(d.lastSpawn) > d.spawnDelay {
			if d.canvasW > 0 && d.canvasH > 0 {
				d.fireworks = append(d.fireworks, NewFirework(d.canvasW, d.canvasH, d.cameraZ))
				d.lastSpawn = m.Time
				d.spawnDelay = time.Duration(400+rand.Intn(600)) * time.Millisecond
			}
		}

		// Update fireworks
		alive := d.fireworks[:0]
		for _, fw := range d.fireworks {
			fw.Update(dt)
			if !fw.IsFinished() && fw.Z-d.cameraZ > -50 {
				alive = append(alive, fw)
			}
		}
		d.fireworks = alive

		return runtime.Handled()

	case runtime.KeyMsg:
		switch m.Rune {
		case 'q':
			return runtime.WithCommand(runtime.Quit{})
		case ' ':
			// Manual launch on space
			if d.canvasW > 0 && d.canvasH > 0 {
				d.fireworks = append(d.fireworks, NewFirework(d.canvasW, d.canvasH, d.cameraZ))
			}
			return runtime.Handled()
		}
	}
	if d.CanvasWidget != nil {
		return d.CanvasWidget.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (d *FireworksDemo) draw(canvas *graphics.Canvas) {
	w, h := canvas.Size()
	if w == 0 || h == 0 {
		return
	}

	cameraDist := 200.0
	centerX := float64(w) / 2
	centerY := float64(h) / 2

	// Draw each firework
	for _, fw := range d.fireworks {
		color := backend.ColorRGB(fw.R, fw.G, fw.B)
		if !fw.Exploded {
			// Draw launch trail
			for _, pt := range fw.Trail {
				zRel := fw.Z - d.cameraZ
				zOffset := zRel + cameraDist
				if zOffset > 0 {
					scale := cameraDist / zOffset
					sx := int(centerX + (pt[0]-centerX)*scale)
					sy := int(centerY + (pt[1]-centerY)*scale)
					if sx >= 0 && sx < w && sy >= 0 && sy < h {
						canvas.Blend(sx, sy, color, 0.8)
					}
				}
			}
		} else {
			// Draw explosion particles
			for i := range fw.Particles {
				p := &fw.Particles[i]
				sx, sy, visible := p.Project3D(d.cameraZ, cameraDist, centerX, centerY)
				if visible && sx >= 0 && sx < w && sy >= 0 && sy < h {
					pcolor := backend.ColorRGB(p.R, p.G, p.B)
					alpha := p.Alpha()
					canvas.Blend(sx, sy, pcolor, alpha)
				}
			}
		}
	}

	// Draw instructions
	canvas.SetStrokeColor(backend.ColorRGB(180, 180, 180))
	canvas.DrawText(2, 2, "SPACE:LAUNCH Q:QUIT", graphics.DefaultFont)
}
