# Tutorial 03: Graphics and Animation

This tutorial draws with the pixel canvas and animates particles.

<a href="../demos/graphics.gif" target="_blank"><img src="../demos/graphics.gif" alt="Canvas graphics"></a>

## Canvas Drawing

```go
widget := widgets.NewCanvasWidget(func(c *graphics.Canvas) {
    c.SetStrokeColor(backend.ColorRGB(255, 183, 77))
    c.DrawLine(0, 0, 40, 12)
    c.FillCircle(24, 10, 6)
})
```

<a href="../demos/easing.gif" target="_blank"><img src="../demos/easing.gif" alt="Animation easing functions"></a>

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

<a href="../demos/fireworks.gif" target="_blank"><img src="../demos/fireworks.gif" alt="3D particle fireworks"></a>

## Reference

See `examples/animation-demo` for a complete scene.
