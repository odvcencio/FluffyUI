package graphics

import "github.com/odvcencio/fluffyui/backend"

// QuadrantBlitter renders 2x2 pixels per cell using block quadrants.
type QuadrantBlitter struct{}

func (q *QuadrantBlitter) PixelsPerCell() (int, int) { return 2, 2 }
func (q *QuadrantBlitter) SupportsColor() bool       { return true }
func (q *QuadrantBlitter) SupportsAlpha() bool       { return false }
func (q *QuadrantBlitter) Name() string              { return "quadrant" }

var quadrantTable = [16]rune{
	' ',      // 0000
	'\u2596', // 0001 - lower left
	'\u2597', // 0010 - lower right
	'\u2584', // 0011 - lower half
	'\u2598', // 0100 - upper left
	'\u258C', // 0101 - left half
	'\u259E', // 0110 - diagonal
	'\u259B', // 0111 - missing lower right
	'\u259D', // 1000 - upper right
	'\u259A', // 1001 - diagonal
	'\u2590', // 1010 - right half
	'\u259C', // 1011 - missing lower left
	'\u2580', // 1100 - upper half
	'\u259F', // 1101 - missing upper right
	'\u259E', // 1110 - missing upper left
	'\u2588', // 1111 - full block
}

func (q *QuadrantBlitter) BlitCell(pixels *PixelBuffer, cx, cy int) (rune, backend.Style) {
	if pixels == nil {
		return ' ', backend.DefaultStyle()
	}
	var pattern int
	fgColors := make([]Color, 0, 4)
	bgColors := make([]Color, 0, 4)
	for dy := 0; dy < 2; dy++ {
		for dx := 0; dx < 2; dx++ {
			px := cx*2 + dx
			py := cy*2 + dy
			bit := 0
			if dy == 1 {
				bit = dx
			} else {
				bit = 2 + dx
			}
			if p := pixels.Get(px, py); p.Set {
				pattern |= 1 << bit
				fgColors = append(fgColors, p.Color)
			} else {
				bgColors = append(bgColors, backend.ColorDefault)
			}
		}
	}
	char := quadrantTable[pattern]
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
