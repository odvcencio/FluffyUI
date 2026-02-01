package animation

import (
	"math"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/graphics"
)

func TestAnimatableLerp(t *testing.T) {
	if got := Float64(0).Lerp(Float64(10), 0.5).(Float64); got != 5 {
		t.Fatalf("Float64 lerp = %v", got)
	}
	if got := Int(0).Lerp(Int(3), 0.5).(Int); got != 2 {
		t.Fatalf("Int lerp = %v", got)
	}
	vec := Vec2{X: 0, Y: 0}.Lerp(Vec2{X: 2, Y: 4}, 0.5).(Vec2)
	if vec.X != 1 || vec.Y != 2 {
		t.Fatalf("Vec2 lerp = %+v", vec)
	}
	color := AnimColor{R: 0, G: 0, B: 0}.Lerp(AnimColor{R: 10, G: 20, B: 30}, 0.5).(AnimColor)
	if color.R != 5 || color.G != 10 || color.B != 15 {
		t.Fatalf("AnimColor lerp = %+v", color)
	}
	rect := AnimRect{X: 0, Y: 0, W: 2, H: 2}.Lerp(AnimRect{X: 2, Y: 2, W: 4, H: 6}, 0.5).(AnimRect)
	if rect.X != 1 || rect.Y != 1 || rect.W != 3 || rect.H != 4 {
		t.Fatalf("AnimRect lerp = %+v", rect)
	}
}

func TestTweenLifecycle(t *testing.T) {
	var current Animatable = Float64(0)
	updated := false
	completed := false

	tween := NewTween(func() Animatable { return current }, func(v Animatable) { current = v }, Float64(10), TweenConfig{
		Duration:   10 * time.Millisecond,
		OnUpdate:   func(Animatable) { updated = true },
		OnComplete: func() { completed = true },
	})
	if tween == nil {
		t.Fatalf("expected tween")
	}
	tween.Start()
	tween.Pause()
	if tween.Update(time.Now().Add(20 * time.Millisecond)) {
		t.Fatalf("expected paused tween to not complete")
	}
	tween.Resume()
	if !tween.Update(time.Now().Add(20 * time.Millisecond)) {
		t.Fatalf("expected tween to complete")
	}
	if !updated || !completed {
		t.Fatalf("expected update and completion")
	}

	completed = false
	tween2 := NewTween(nil, nil, Float64(5), TweenConfig{Duration: 0, OnComplete: func() { completed = true }})
	tween2.Start()
	tween2.Complete()
	if !completed {
		t.Fatalf("expected complete to trigger")
	}
	tween2.Stop()
}

func TestAnimatorAndParticles(t *testing.T) {
	anim := NewAnimator()
	value := Float64(0)
	anim.Animate(&value, "x", func() Animatable { return value }, func(v Animatable) { value = v.(Float64) }, Float64(1), TweenConfig{Duration: time.Millisecond})
	if !anim.Update(0.01) {
		t.Fatalf("expected animator to have active animations")
	}

	spring := NewSpring(0, SpringConfig{Tension: 120, Friction: 12, Mass: 1})
	anim.AnimateSpring(&value, "y", spring, 1)
	anim.AddParticleSystem(NewParticleSystem(4))
	if !anim.HasActiveAnimations() {
		t.Fatalf("expected active animations")
	}
	anim.Clear()
}

func TestParticleSystemRenderAndClear(t *testing.T) {
	ps := NewParticleSystem(4)
	ps.Emit(Particle{Life: 0.2, MaxLife: 0.2, Size: 1})

	canvas := graphics.NewCanvas(2, 2)
	ps.Render(canvas)
	ps.Clear()
	if len(ps.particles) != 0 {
		t.Fatalf("expected cleared particles")
	}

	// Exercise emitFromEmitter indirectly
	emitter := &Emitter{Active: true, Rate: 2, Life: Range{Min: 0.2, Max: 0.2}, Size: Range{Min: 1, Max: 1}}
	emitter.Active = true
	emitter.Rate = 1
	emitter.Speed = Range{Min: 1, Max: 1}
	emitter.Spread = math.Pi / 2
	ps.AddEmitter(emitter)
	ps.Update(1)
}
