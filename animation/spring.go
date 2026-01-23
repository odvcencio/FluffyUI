package animation

import "math"

// Spring provides physics-based animation.
type Spring struct {
	Value    float64
	Target   float64
	Velocity float64

	Tension  float64
	Friction float64
	Mass     float64

	RestSpeed    float64
	RestDistance float64

	onUpdate   func(value float64)
	onComplete func()

	atRest bool
}

// SpringConfig configures a spring.
type SpringConfig struct {
	Tension    float64
	Friction   float64
	Mass       float64
	OnUpdate   func(value float64)
	OnComplete func()
}

// Presets for common spring behaviors.
var (
	SpringDefault = SpringConfig{Tension: 170, Friction: 26, Mass: 1}
	SpringGentle  = SpringConfig{Tension: 120, Friction: 14, Mass: 1}
	SpringWobbly  = SpringConfig{Tension: 180, Friction: 12, Mass: 1}
	SpringStiff   = SpringConfig{Tension: 210, Friction: 20, Mass: 1}
	SpringSlow    = SpringConfig{Tension: 280, Friction: 60, Mass: 1}
	SpringMolassy = SpringConfig{Tension: 280, Friction: 120, Mass: 1}
)

// NewSpring creates a spring.
func NewSpring(initial float64, cfg SpringConfig) *Spring {
	if cfg.Tension == 0 {
		cfg.Tension = SpringDefault.Tension
	}
	if cfg.Friction == 0 {
		cfg.Friction = SpringDefault.Friction
	}
	if cfg.Mass == 0 {
		cfg.Mass = 1
	}
	return &Spring{
		Value:        initial,
		Target:       initial,
		Tension:      cfg.Tension,
		Friction:     cfg.Friction,
		Mass:         cfg.Mass,
		RestSpeed:    0.001,
		RestDistance: 0.001,
		onUpdate:     cfg.OnUpdate,
		onComplete:   cfg.OnComplete,
	}
}

// SetTarget sets the target value.
func (s *Spring) SetTarget(target float64) {
	if s == nil {
		return
	}
	s.Target = target
	s.atRest = false
}

// Update advances the spring simulation.
func (s *Spring) Update(dt float64) bool {
	if s == nil || s.atRest {
		return s != nil && s.atRest
	}
	if dt <= 0 {
		return false
	}
	displacement := s.Value - s.Target
	springForce := -s.Tension * displacement
	dampingForce := -s.Friction * s.Velocity
	acceleration := (springForce + dampingForce) / s.Mass

	s.Velocity += acceleration * dt
	s.Value += s.Velocity * dt

	speed := math.Abs(s.Velocity)
	distance := math.Abs(s.Value - s.Target)

	if speed < s.RestSpeed && distance < s.RestDistance {
		s.Value = s.Target
		s.Velocity = 0
		s.atRest = true
		if s.onComplete != nil {
			s.onComplete()
		}
	}

	if s.onUpdate != nil {
		s.onUpdate(s.Value)
	}

	return s.atRest
}

// AtRest reports whether the spring has settled.
func (s *Spring) AtRest() bool {
	if s == nil {
		return true
	}
	return s.atRest
}
