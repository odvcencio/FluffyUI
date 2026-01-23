package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// SixelBlitter currently falls back to sextant rendering.
type SixelBlitter struct {
	fallback SextantBlitter
}

func (s *SixelBlitter) PixelsPerCell() (int, int) { return s.fallback.PixelsPerCell() }
func (s *SixelBlitter) SupportsColor() bool       { return s.fallback.SupportsColor() }
func (s *SixelBlitter) SupportsAlpha() bool       { return s.fallback.SupportsAlpha() }
func (s *SixelBlitter) Name() string              { return "sixel" }

func (s *SixelBlitter) BlitCell(pixels *PixelBuffer, cellX, cellY int) (rune, backend.Style) {
	return s.fallback.BlitCell(pixels, cellX, cellY)
}

// BuildImage builds a sixel image payload from the pixel buffer.
func (s *SixelBlitter) BuildImage(pixels *PixelBuffer, cellWidth, cellHeight int) backend.Image {
	return buildRGBAImage(pixels, cellWidth, cellHeight, backend.ImageProtocolSixel)
}
