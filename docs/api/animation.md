# Animation API

FluffyUI includes tweens, springs, and a particle system for terminal effects.

## Tweens

Tweens animate values over time using easing functions.

```go
tween := animation.NewTween(animation.Float64(0), animation.Float64(100), 2*time.Second)
value := tween.Update(dt)
```

Use `animation.EaseIn`, `animation.EaseOut`, and friends to change timing.

## Springs

Springs model natural motion with stiffness and damping.

```go
spring := animation.NewSpring(animation.SpringConfig{Stiffness: 180, Damping: 22})
spring.SetTarget(1)
value := spring.Update(dt)
```

## Particles

```go
ps := animation.NewParticleSystem(256)
ps.AddEmitter(&animation.Emitter{
    Position:  animation.Vector2{X: 10, Y: 10},
    Rate:      40,
    Spread:    math.Pi,
    Speed:     animation.Range{Min: 2, Max: 6},
    Life:      animation.Range{Min: 0.5, Max: 1.5},
    Size:      animation.Range{Min: 1, Max: 2},
    Color:     animation.ColorRange{Start: backend.ColorRed, End: backend.ColorYellow},
    Active:    true,
})

ps.Update(dt)
ps.Render(canvas)
```

## Force Fields

Force fields influence particle velocity each update:

- `GravityField` (constant force)
- `RadialField` (attract/repel from a center)
- `VortexField` (rotational flow)
- `TurbulenceField` (oscillating noise)

```go
ps.AddForceField(&animation.RadialField{Center: animation.Vector2{X: 40, Y: 12}, Strength: 80})
ps.AddForceField(&animation.VortexField{Center: animation.Vector2{X: 40, Y: 12}, Strength: 25})
```
