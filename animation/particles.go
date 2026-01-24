package animation

import (
	"math"
	"math/rand"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/graphics"
)

// Vector2 is a 2D vector.
type Vector2 = Vec2

// Color represents a terminal color.
type Color = backend.Color

// Particle represents a single particle.
type Particle struct {
	Position      Vector2
	Velocity      Vector2
	Gravity       Vector2
	Color         Color
	Size          float64
	Life          float64
	MaxLife       float64
	Alpha         float64
	Rotation      float64
	RotationSpeed float64
}

// Range represents a numeric range.
type Range struct {
	Min, Max float64
}

// ColorRange represents a start/end color range.
type ColorRange struct {
	Start, End Color
}

// ParticleConfig configures a burst.
type ParticleConfig struct {
	Speed         Range
	Life          Range
	Size          Range
	Spread        float64
	Direction     float64
	Color         ColorRange
	Gravity       Vector2
	RotationSpeed Range
}

// Emitter produces particles.
type Emitter struct {
	Position      Vector2
	Rate          float64
	Spread        float64
	Direction     float64
	Speed         Range
	Life          Range
	Size          Range
	Color         ColorRange
	Gravity       Vector2
	RotationSpeed Range
	Active        bool

	accumulator float64
}

// ParticleSystem manages particles and emitters.
type ParticleSystem struct {
	particles    []Particle
	emitters     []*Emitter
	maxParticles int
	forceFields  []ForceField
}

// NewParticleSystem creates a particle system.
func NewParticleSystem(maxParticles int) *ParticleSystem {
	if maxParticles <= 0 {
		maxParticles = 128
	}
	return &ParticleSystem{
		particles:    make([]Particle, 0, maxParticles),
		maxParticles: maxParticles,
	}
}

// AddEmitter adds an emitter.
func (ps *ParticleSystem) AddEmitter(e *Emitter) {
	if ps == nil || e == nil {
		return
	}
	ps.emitters = append(ps.emitters, e)
}

// AddForceField registers a force field.
func (ps *ParticleSystem) AddForceField(f ForceField) {
	if ps == nil || f == nil {
		return
	}
	ps.forceFields = append(ps.forceFields, f)
}

// Emit spawns a single particle.
func (ps *ParticleSystem) Emit(p Particle) {
	if ps == nil {
		return
	}
	if len(ps.particles) >= ps.maxParticles {
		return
	}
	ps.particles = append(ps.particles, p)
}

// Burst emits multiple particles at once.
func (ps *ParticleSystem) Burst(position Vector2, count int, cfg ParticleConfig) {
	if ps == nil || count <= 0 {
		return
	}
	for i := 0; i < count && len(ps.particles) < ps.maxParticles; i++ {
		angle := cfg.Direction + (rand.Float64()-0.5)*cfg.Spread
		speed := randRange(cfg.Speed)
		life := randRange(cfg.Life)
		if life <= 0 {
			life = 0.1
		}
		size := randRange(cfg.Size)
		rotationSpeed := randRange(cfg.RotationSpeed)
		ps.particles = append(ps.particles, Particle{
			Position:      position,
			Velocity:      Vector2{X: math.Cos(angle) * speed, Y: math.Sin(angle) * speed},
			Gravity:       cfg.Gravity,
			Color:         lerpColor(cfg.Color.Start, cfg.Color.End, rand.Float64()),
			Size:          size,
			Life:          life,
			MaxLife:       life,
			Alpha:         1,
			Rotation:      0,
			RotationSpeed: rotationSpeed,
		})
	}
}

// Update advances the particle system.
func (ps *ParticleSystem) Update(dt float64) {
	if ps == nil || dt <= 0 {
		return
	}
	for _, e := range ps.emitters {
		if e == nil || !e.Active {
			continue
		}
		e.accumulator += e.Rate * dt
		for e.accumulator >= 1 {
			e.accumulator--
			ps.emitFromEmitter(e)
		}
	}

	alive := ps.particles[:0]
	for i := range ps.particles {
		p := ps.particles[i]
		for _, field := range ps.forceFields {
			force := field.Apply(p.Position)
			p.Velocity.X += force.X * dt
			p.Velocity.Y += force.Y * dt
		}
		p.Velocity.X += p.Gravity.X * dt
		p.Velocity.Y += p.Gravity.Y * dt

		p.Position.X += p.Velocity.X * dt
		p.Position.Y += p.Velocity.Y * dt

		p.Rotation += p.RotationSpeed * dt
		p.Life -= dt
		if p.MaxLife > 0 {
			p.Alpha = p.Life / p.MaxLife
		}
		if p.Life > 0 {
			alive = append(alive, p)
		}
	}
	ps.particles = alive
}

// Render draws particles to a canvas.
func (ps *ParticleSystem) Render(canvas *graphics.Canvas) {
	if ps == nil || canvas == nil {
		return
	}
	for _, p := range ps.particles {
		x := int(math.Round(p.Position.X))
		y := int(math.Round(p.Position.Y))
		if p.Size <= 1 {
			canvas.SetPixel(x, y, p.Color)
			continue
		}
		radius := int(math.Round(p.Size / 2))
		canvas.SetFillColor(p.Color)
		canvas.FillCircle(x, y, radius)
	}
}

// Clear removes all particles.
func (ps *ParticleSystem) Clear() {
	if ps == nil {
		return
	}
	ps.particles = ps.particles[:0]
}

// ParticleCount returns the current number of live particles.
func (ps *ParticleSystem) ParticleCount() int {
	if ps == nil {
		return 0
	}
	return len(ps.particles)
}

// MaxParticles returns the particle limit.
func (ps *ParticleSystem) MaxParticles() int {
	if ps == nil {
		return 0
	}
	return ps.maxParticles
}

// SetMaxParticles adjusts the particle limit. Existing particles are preserved
// up to the new limit.
func (ps *ParticleSystem) SetMaxParticles(max int) {
	if ps == nil || max <= 0 {
		return
	}
	ps.maxParticles = max
	if len(ps.particles) > max {
		ps.particles = ps.particles[:max]
	}
}

// RemoveEmitter removes an emitter from the system.
func (ps *ParticleSystem) RemoveEmitter(e *Emitter) {
	if ps == nil || e == nil {
		return
	}
	for i, em := range ps.emitters {
		if em == e {
			ps.emitters = append(ps.emitters[:i], ps.emitters[i+1:]...)
			return
		}
	}
}

// ClearEmitters removes all emitters.
func (ps *ParticleSystem) ClearEmitters() {
	if ps == nil {
		return
	}
	ps.emitters = ps.emitters[:0]
}

// RemoveForceField removes a force field from the system.
func (ps *ParticleSystem) RemoveForceField(f ForceField) {
	if ps == nil || f == nil {
		return
	}
	for i, field := range ps.forceFields {
		if field == f {
			ps.forceFields = append(ps.forceFields[:i], ps.forceFields[i+1:]...)
			return
		}
	}
}

// ClearForceFields removes all force fields.
func (ps *ParticleSystem) ClearForceFields() {
	if ps == nil {
		return
	}
	ps.forceFields = ps.forceFields[:0]
}

func (ps *ParticleSystem) emitFromEmitter(e *Emitter) {
	if ps == nil || e == nil {
		return
	}
	if len(ps.particles) >= ps.maxParticles {
		return
	}
	angle := e.Direction + (rand.Float64()-0.5)*e.Spread
	speed := randRange(e.Speed)
	life := randRange(e.Life)
	if life <= 0 {
		life = 0.1
	}
	size := randRange(e.Size)
	rotationSpeed := randRange(e.RotationSpeed)
	ps.particles = append(ps.particles, Particle{
		Position:      e.Position,
		Velocity:      Vector2{X: math.Cos(angle) * speed, Y: math.Sin(angle) * speed},
		Gravity:       e.Gravity,
		Color:         lerpColor(e.Color.Start, e.Color.End, rand.Float64()),
		Size:          size,
		Life:          life,
		MaxLife:       life,
		Alpha:         1,
		Rotation:      0,
		RotationSpeed: rotationSpeed,
	})
}

func randRange(r Range) float64 {
	if r.Max <= r.Min {
		return r.Min
	}
	return r.Min + rand.Float64()*(r.Max-r.Min)
}

func lerpColor(a, b Color, t float64) Color {
	if !a.IsRGB() || !b.IsRGB() {
		if t >= 0.5 {
			return b
		}
		return a
	}
	ar, ag, ab := a.RGB()
	br, bg, bb := b.RGB()
	r := uint8(float64(ar) + (float64(br)-float64(ar))*t)
	g := uint8(float64(ag) + (float64(bg)-float64(ag))*t)
	bval := uint8(float64(ab) + (float64(bb)-float64(ab))*t)
	return backend.ColorRGB(r, g, bval)
}
