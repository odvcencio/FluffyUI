# Tutorial 03: Graphics and Animation

This tutorial draws with the pixel canvas and animates particles.

## Canvas Drawing

```go
widget := widgets.NewCanvasWidget(func(c *graphics.Canvas) {
    c.SetStrokeColor(backend.ColorRGB(255, 183, 77))
    c.DrawLine(0, 0, 40, 12)
    c.FillCircle(24, 10, 6)
})
```

## Particles + Force Fields

```go
ps := animation.NewParticleSystem(256)
ps.AddForceField(&animation.RadialField{
    Center:   animation.Vector2{X: 40, Y: 12},
    Strength: 80,
})

ps.Update(dt)
ps.Render(canvas)
```

## Reference

See `examples/animation-demo` for a complete scene.
