package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// ASCIIBlitter renders a single pixel as an ASCII character.
type ASCIIBlitter struct{}

func (a *ASCIIBlitter) PixelsPerCell() (int, int) { return 1, 1 }
func (a *ASCIIBlitter) SupportsColor() bool       { return true }
func (a *ASCIIBlitter) SupportsAlpha() bool       { return false }
func (a *ASCIIBlitter) Name() string              { return "ascii" }

func (a *ASCIIBlitter) BlitCell(pixels *PixelBuffer, cellX, cellY int) (rune, backend.Style) {
	if pixels == nil {
		return ' ', backend.DefaultStyle()
	}
	p := pixels.Get(cellX, cellY)
	char := ' '
	if p.Set {
		char = '#'
	}
	style := backend.DefaultStyle().Foreground(p.Color)
	return char, style
}
