# Graphics API

FluffyUI exposes a pixel-level canvas that can render via Unicode cell blitters
or terminal image protocols (Kitty/Sixel).

## Canvas

Use `graphics.Canvas` for pixel drawing:

```go
canvas := graphics.NewCanvasWithBlitter(widthCells, heightCells, &graphics.SextantBlitter{})
canvas.SetStrokeColor(backend.ColorRGB(255, 183, 77))
canvas.DrawLine(0, 0, 40, 20)
canvas.FillCircle(30, 12, 6)
canvas.Render(ctx.Buffer, x, y)
```

Key methods:

- `SetPixel`, `GetPixel`, `Clear`
- `DrawLine`, `DrawRect`, `DrawCircle`, `DrawBezier`
- `DrawText` with `graphics.PixelFont`
- `DrawImage` and `DrawImageScaled` for `image.Image`

## Blitters and Image Protocols

Blitters map pixels to terminal cells. Use `graphics.BestBlitter(nil)` to pick
Kitty/Sixel when supported, otherwise Unicode or ASCII.

Image protocol blitters (`KittyBlitter`, `SixelBlitter`) implement
`graphics.ImageBlitter`, which allows the canvas to emit full images for
supported backends.

## Image Loading

Load common image formats (PNG/JPEG/GIF/WebP) with helpers:

```go
img, err := graphics.LoadImage("assets/logo.png")
scaled := graphics.ScaleImage(img, 320, 180)
```

Helpers:

- `LoadImage(path)`
- `LoadImageScaled(path, maxWidth, maxHeight)`
- `ScaleImage(img, maxWidth, maxHeight)`

## Tips

- Use `DrawImageScaled` to fit frames to the canvas size.
- Pair a `CanvasWidget` with a custom draw callback for reusable renderers.
- For best fidelity, prefer Kitty or Sixel blitters when terminals support them.
