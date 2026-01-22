package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// BrailleBlitter renders 2x4 pixels per cell using braille patterns.
type BrailleBlitter struct{}

func (b *BrailleBlitter) PixelsPerCell() (int, int) { return 2, 4 }
func (b *BrailleBlitter) SupportsColor() bool       { return true }
func (b *BrailleBlitter) SupportsAlpha() bool       { return false }
func (b *BrailleBlitter) Name() string              { return "braille" }

func (b *BrailleBlitter) BlitCell(pixels *PixelBuffer, cx, cy int) (rune, backend.Style) {
	if pixels == nil {
		return ' ', backend.DefaultStyle()
	}
	var pattern int
	colors := make([]Color, 0, 8)
	offsets := []struct {
		dx  int
		dy  int
		bit int
	}{
		{0, 0, 0}, {0, 1, 1}, {0, 2, 2}, {0, 3, 6},
		{1, 0, 3}, {1, 1, 4}, {1, 2, 5}, {1, 3, 7},
	}
	for _, off := range offsets {
		px := cx*2 + off.dx
		py := cy*4 + off.dy
		if p := pixels.Get(px, py); p.Set {
			pattern |= 1 << off.bit
			colors = append(colors, p.Color)
		}
	}
	if pattern == 0 {
		return ' ', backend.DefaultStyle()
	}
	fg := dominantColor(colors)
	style := backend.DefaultStyle()
	if fg != backend.ColorDefault {
		style = style.Foreground(fg)
	}
	return rune(0x2800 + pattern), style
}
