package graphics

import "github.com/odvcencio/fluffyui/backend"

// PixelBuffer stores sub-cell pixel data.
type PixelBuffer struct {
	width, height int
	pixels        []Pixel
	dirty         bool
}

// NewPixelBuffer creates a buffer for the given cell dimensions.
func NewPixelBuffer(cellWidth, cellHeight int, blitter Blitter) *PixelBuffer {
	pw, ph := 1, 1
	if blitter != nil {
		pw, ph = blitter.PixelsPerCell()
		if pw <= 0 {
			pw = 1
		}
		if ph <= 0 {
			ph = 1
		}
	}
	width := cellWidth * pw
	height := cellHeight * ph
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return &PixelBuffer{
		width:  width,
		height: height,
		pixels: make([]Pixel, width*height),
	}
}

// Size returns pixel dimensions.
func (b *PixelBuffer) Size() (width, height int) {
	if b == nil {
		return 0, 0
	}
	return b.width, b.height
}

// Get returns the pixel at (x, y).
func (b *PixelBuffer) Get(x, y int) Pixel {
	if b == nil || x < 0 || x >= b.width || y < 0 || y >= b.height {
		return Pixel{}
	}
	return b.pixels[y*b.width+x]
}

// Set sets a pixel.
func (b *PixelBuffer) Set(x, y int, p Pixel) {
	if b == nil || x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}
	b.pixels[y*b.width+x] = p
	b.dirty = true
}

// SetPixel sets an "on" pixel with color.
func (b *PixelBuffer) SetPixel(x, y int, color Color) {
	b.Set(x, y, Pixel{Set: true, Color: color, Alpha: 1})
}

// Clear resets all pixels.
func (b *PixelBuffer) Clear() {
	if b == nil {
		return
	}
	for i := range b.pixels {
		b.pixels[i] = Pixel{}
	}
	b.dirty = true
}

// Blend blends a pixel with alpha.
func (b *PixelBuffer) Blend(x, y int, color Color, alpha float32) {
	if b == nil || x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}
	existing := b.Get(x, y)
	if !existing.Set || existing.Alpha == 0 {
		b.Set(x, y, Pixel{Set: true, Color: color, Alpha: alpha})
		return
	}
	blended := blendColors(existing.Color, color, float64(alpha))
	newAlpha := existing.Alpha + alpha*(1-existing.Alpha)
	if newAlpha > 1 {
		newAlpha = 1
	}
	b.Set(x, y, Pixel{Set: true, Color: blended, Alpha: newAlpha})
}

func blendColors(base, top Color, alpha float64) Color {
	if alpha <= 0 {
		return base
	}
	if alpha >= 1 {
		return top
	}
	if !base.IsRGB() || !top.IsRGB() {
		if alpha >= 0.5 {
			return top
		}
		return base
	}
	br, bg, bb := base.RGB()
	tr, tg, tb := top.RGB()
	r := uint8(float64(br)*(1-alpha) + float64(tr)*alpha + 0.5)
	g := uint8(float64(bg)*(1-alpha) + float64(tg)*alpha + 0.5)
	b := uint8(float64(bb)*(1-alpha) + float64(tb)*alpha + 0.5)
	return backend.ColorRGB(r, g, b)
}
