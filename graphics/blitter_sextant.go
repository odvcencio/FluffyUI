package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// SextantBlitter renders 2x3 pixels per cell using sextant blocks.
type SextantBlitter struct{}

func (s *SextantBlitter) PixelsPerCell() (int, int) { return 2, 3 }
func (s *SextantBlitter) SupportsColor() bool       { return true }
func (s *SextantBlitter) SupportsAlpha() bool       { return true }
func (s *SextantBlitter) Name() string              { return "sextant" }

const (
	sextantBase      = 0x1FB00
	sextantLeftHalf  = 0x15 // 010101
	sextantRightHalf = 0x2A // 101010
	sextantFull      = 0x3F // 111111
)

var sextantTable = func() [64]rune {
	var table [64]rune
	idx := 0
	for pattern := 0; pattern < 64; pattern++ {
		switch pattern {
		case 0:
			table[pattern] = ' '
		case sextantLeftHalf:
			table[pattern] = '\u258C'
		case sextantRightHalf:
			table[pattern] = '\u2590'
		case sextantFull:
			table[pattern] = '\u2588'
		default:
			table[pattern] = rune(sextantBase + idx)
			idx++
		}
	}
	return table
}()

func (s *SextantBlitter) BlitCell(pixels *PixelBuffer, cx, cy int) (rune, backend.Style) {
	if pixels == nil {
		return ' ', backend.DefaultStyle()
	}
	var pattern int
	fgColors := make([]Color, 0, 6)
	bgColors := make([]Color, 0, 6)
	for i := 0; i < 6; i++ {
		dx := i % 2
		dy := i / 2
		px := cx*2 + dx
		py := cy*3 + dy
		if p := pixels.Get(px, py); p.Set {
			pattern |= 1 << i
			fgColors = append(fgColors, p.Color)
		} else {
			bgColors = append(bgColors, backend.ColorBlack)
		}
	}
	char := sextantTable[pattern]
	style := backend.DefaultStyle()
	fg := dominantColor(fgColors)
	bg := dominantColor(bgColors)
	if fg != backend.ColorDefault {
		style = style.Foreground(fg)
	}
	// Always set background to ensure proper rendering in recordings/GIFs
	if bg == backend.ColorDefault {
		bg = backend.ColorBlack
	}
	style = style.Background(bg)
	return char, style
}
