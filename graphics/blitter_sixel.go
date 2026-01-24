package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// Default Sixel cell dimensions
const (
	DefaultSixelCellWidth  = 10
	DefaultSixelCellHeight = 12 // 2 sixel bands (6 pixels each)
)

// SixelBlitter renders images using the Sixel graphics protocol.
// Sixel uses 6-pixel-high bands and typically ~10 pixels per cell width.
type SixelBlitter struct {
	// CellWidth overrides the default cell width in pixels.
	// Set to 0 to use the default (10 pixels).
	CellWidth int
	// CellHeight overrides the default cell height in pixels.
	// Set to 0 to use the default (12 pixels = 2 sixel bands).
	CellHeight int
}

// PixelsPerCell returns the pixel resolution per cell.
// Returns configured dimensions or defaults (10x12).
func (s *SixelBlitter) PixelsPerCell() (int, int) {
	w, h := DefaultSixelCellWidth, DefaultSixelCellHeight
	if s != nil {
		if s.CellWidth > 0 {
			w = s.CellWidth
		}
		if s.CellHeight > 0 {
			h = s.CellHeight
		}
	}
	return w, h
}

// SupportsColor returns true - Sixel supports palette-based color (up to 256 colors).
func (s *SixelBlitter) SupportsColor() bool { return true }

// SupportsAlpha returns false - Sixel protocol does not support alpha transparency.
func (s *SixelBlitter) SupportsAlpha() bool { return false }

// Name returns the blitter identifier.
func (s *SixelBlitter) Name() string { return "sixel" }

// BlitCell returns a space - actual rendering uses BuildImage.
// When Sixel protocol is used, the Canvas.Render method will call BuildImage
// and send the full image to the terminal, then fill cells with spaces.
func (s *SixelBlitter) BlitCell(pixels *PixelBuffer, cellX, cellY int) (rune, backend.Style) {
	return ' ', backend.DefaultStyle()
}

// BuildImage builds a Sixel graphics protocol image from the pixel buffer.
func (s *SixelBlitter) BuildImage(pixels *PixelBuffer, cellWidth, cellHeight int) backend.Image {
	return buildRGBAImage(pixels, cellWidth, cellHeight, backend.ImageProtocolSixel)
}
