package graphics

import "github.com/odvcencio/fluffyui/backend"

// Default Kitty cell dimensions (typical for most terminal fonts)
const (
	DefaultKittyCellWidth  = 10
	DefaultKittyCellHeight = 20
)

// KittyBlitter renders images using the Kitty graphics protocol.
// Typical terminal cells are ~10x20 pixels, providing high-resolution graphics.
type KittyBlitter struct {
	// CellWidth overrides the default cell width in pixels.
	// Set to 0 to use the default (10 pixels).
	CellWidth int
	// CellHeight overrides the default cell height in pixels.
	// Set to 0 to use the default (20 pixels).
	CellHeight int
}

// PixelsPerCell returns the pixel resolution per cell.
// Returns configured dimensions or defaults (10x20).
func (k *KittyBlitter) PixelsPerCell() (int, int) {
	w, h := DefaultKittyCellWidth, DefaultKittyCellHeight
	if k != nil {
		if k.CellWidth > 0 {
			w = k.CellWidth
		}
		if k.CellHeight > 0 {
			h = k.CellHeight
		}
	}
	return w, h
}

// SupportsColor returns true - Kitty protocol supports full 24-bit color.
func (k *KittyBlitter) SupportsColor() bool { return true }

// SupportsAlpha returns true - Kitty protocol supports alpha transparency.
func (k *KittyBlitter) SupportsAlpha() bool { return true }

// Name returns the blitter identifier.
func (k *KittyBlitter) Name() string { return "kitty" }

// BlitCell returns a space - actual rendering uses BuildImage.
// When Kitty protocol is used, the Canvas.Render method will call BuildImage
// and send the full image to the terminal, then fill cells with spaces.
func (k *KittyBlitter) BlitCell(pixels *PixelBuffer, cellX, cellY int) (rune, backend.Style) {
	return ' ', backend.DefaultStyle()
}

// BuildImage builds a Kitty graphics protocol image from the pixel buffer.
func (k *KittyBlitter) BuildImage(pixels *PixelBuffer, cellWidth, cellHeight int) backend.Image {
	return buildRGBAImage(pixels, cellWidth, cellHeight, backend.ImageProtocolKitty)
}
