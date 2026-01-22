package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// HalfBlockBlitter renders 1x2 pixels per cell using half blocks.
type HalfBlockBlitter struct{}

func (h *HalfBlockBlitter) PixelsPerCell() (int, int) { return 1, 2 }
func (h *HalfBlockBlitter) SupportsColor() bool       { return true }
func (h *HalfBlockBlitter) SupportsAlpha() bool       { return false }
func (h *HalfBlockBlitter) Name() string              { return "halfblock" }

var halfBlockTable = [4]rune{
	' ',      // 00 - empty
	'\u2584', // 01 - lower half
	'\u2580', // 10 - upper half
	'\u2588', // 11 - full block
}

func (h *HalfBlockBlitter) BlitCell(pixels *PixelBuffer, cx, cy int) (rune, backend.Style) {
	if pixels == nil {
		return ' ', backend.DefaultStyle()
	}
	var pattern int
	fgColors := make([]Color, 0, 2)
	bgColors := make([]Color, 0, 2)
	for dy := 0; dy < 2; dy++ {
		px := cx
		py := cy*2 + dy
		bit := 0
		if dy == 0 {
			bit = 1
		}
		if p := pixels.Get(px, py); p.Set {
			pattern |= 1 << bit
			fgColors = append(fgColors, p.Color)
		} else {
			bgColors = append(bgColors, backend.ColorDefault)
		}
	}
	char := halfBlockTable[pattern]
	style := backend.DefaultStyle()
	fg := dominantColor(fgColors)
	bg := dominantColor(bgColors)
	if fg != backend.ColorDefault {
		style = style.Foreground(fg)
	}
	if bg != backend.ColorDefault {
		style = style.Background(bg)
	}
	return char, style
}
