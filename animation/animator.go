package animation

import (
	"reflect"
	"sync"
	"time"
)

type animationKey struct {
	target   uintptr
	property string
}

// Animator manages active animations.
type Animator struct {
	tweens    map[animationKey]*Tween
	springs   map[animationKey]*Spring
	particles []*ParticleSystem

	mu sync.Mutex
}

// NewAnimator creates an animator.
func NewAnimator() *Animator {
	return &Animator{
		tweens:  make(map[animationKey]*Tween),
		springs: make(map[animationKey]*Spring),
	}
}

// Animate starts a tween animation.
func (a *Animator) Animate(target any, property string, getValue func() Animatable, setValue func(Animatable), endValue Animatable, cfg TweenConfig) *Tween {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	key := animationKey{target: targetPointer(target), property: property}
	if existing, ok := a.tweens[key]; ok {
		existing.Stop()
	}
	tween := NewTween(getValue, setValue, endValue, cfg)
	tween.Start()
	a.tweens[key] = tween
	return tween
}

// AnimateSpring starts a spring animation.
func (a *Animator) AnimateSpring(target any, property string, spring *Spring, targetValue float64) *Spring {
	if a == nil || spring == nil {
		return spring
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	key := animationKey{target: targetPointer(target), property: property}
	spring.SetTarget(targetValue)
	a.springs[key] = spring
	return spring
}

// AddParticleSystem registers a particle system.
func (a *Animator) AddParticleSystem(ps *ParticleSystem) {
	if a == nil || ps == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.particles = append(a.particles, ps)
}

// Update advances all animations, returns true if any are active.
func (a *Animator) Update(dt float64) bool {
	if a == nil {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	active := false

	for key, tween := range a.tweens {
		if tween.Update(now) {
			delete(a.tweens, key)
		} else {
			active = true
		}
	}
	for key, spring := range a.springs {
		if spring.Update(dt) {
			delete(a.springs, key)
		} else {
			active = true
		}
	}
	for _, ps := range a.particles {
		ps.Update(dt)
		if len(ps.particles) > 0 {
			active = true
		}
	}
	return active
}

// HasActiveAnimations returns true if animations are running.
func (a *Animator) HasActiveAnimations() bool {
	if a == nil {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.tweens) > 0 || len(a.springs) > 0 {
		return true
	}
	for _, ps := range a.particles {
		if len(ps.particles) > 0 {
			return true
		}
	}
	return false
}

// Clear stops all animations.
func (a *Animator) Clear() {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tweens = make(map[animationKey]*Tween)
	a.springs = make(map[animationKey]*Spring)
	for _, ps := range a.particles {
		ps.Clear()
	}
}

func targetPointer(target any) uintptr {
	if target == nil {
		return 0
	}
	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Pointer || val.Kind() == reflect.Map || val.Kind() == reflect.Slice ||
		val.Kind() == reflect.Func || val.Kind() == reflect.Chan || val.Kind() == reflect.UnsafePointer {
		return val.Pointer()
	}
	return reflect.ValueOf(&target).Pointer()
}
