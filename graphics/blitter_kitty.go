package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// KittyBlitter currently falls back to sextant rendering.
type KittyBlitter struct {
	fallback SextantBlitter
}

func (k *KittyBlitter) PixelsPerCell() (int, int) { return k.fallback.PixelsPerCell() }
func (k *KittyBlitter) SupportsColor() bool       { return k.fallback.SupportsColor() }
func (k *KittyBlitter) SupportsAlpha() bool       { return k.fallback.SupportsAlpha() }
func (k *KittyBlitter) Name() string              { return "kitty" }

func (k *KittyBlitter) BlitCell(pixels *PixelBuffer, cellX, cellY int) (rune, backend.Style) {
	return k.fallback.BlitCell(pixels, cellX, cellY)
}

// BuildImage builds a kitty image payload from the pixel buffer.
func (k *KittyBlitter) BuildImage(pixels *PixelBuffer, cellWidth, cellHeight int) backend.Image {
	return buildRGBAImage(pixels, cellWidth, cellHeight, backend.ImageProtocolKitty)
}
