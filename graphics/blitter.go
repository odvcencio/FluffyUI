package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// Blitter converts pixel data to terminal cells.
type Blitter interface {
	// PixelsPerCell returns the resolution within a single cell.
	PixelsPerCell() (width, height int)

	// BlitCell converts a region of pixels to a cell.
	BlitCell(pixels *PixelBuffer, cellX, cellY int) (rune, backend.Style)

	// SupportsColor returns true if this blitter supports per-pixel color.
	SupportsColor() bool

	// SupportsAlpha returns true if this blitter can render alpha.
	SupportsAlpha() bool

	// Name returns the blitter identifier.
	Name() string
}

// BlitterType identifies a blitter configuration.
type BlitterType int

const (
	BlitterAuto BlitterType = iota
	BlitterHalfBlock
	BlitterQuadrant
	BlitterSextant
	BlitterBraille
	BlitterKitty
	BlitterSixel
	BlitterASCII
)

// Capabilities describes terminal graphics support.
type Capabilities struct{}

// NewBlitter creates a blitter by type.
func NewBlitter(typ BlitterType) Blitter {
	switch typ {
	case BlitterHalfBlock:
		return &HalfBlockBlitter{}
	case BlitterQuadrant:
		return &QuadrantBlitter{}
	case BlitterSextant:
		return &SextantBlitter{}
	case BlitterBraille:
		return &BrailleBlitter{}
	case BlitterAuto:
		return BestBlitter(nil)
	default:
		return &SextantBlitter{}
	}
}

// BestBlitter returns the best blitter for the current terminal.
func BestBlitter(_ *Capabilities) Blitter {
	return &SextantBlitter{}
}

func dominantColor(colors []Color) Color {
	if len(colors) == 0 {
		return backend.ColorDefault
	}
	counts := make(map[Color]int, len(colors))
	bestColor := backend.ColorDefault
	bestCount := 0
	for _, c := range colors {
		counts[c]++
	}
	for c, count := range counts {
		if c == backend.ColorDefault {
			continue
		}
		if count > bestCount {
			bestCount = count
			bestColor = c
		}
	}
	if bestCount == 0 {
		return backend.ColorDefault
	}
	return bestColor
}
