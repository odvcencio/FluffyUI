package graphics

import (
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/terminal"
)

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

// ImageBlitter can emit a full image for a canvas.
type ImageBlitter interface {
	Blitter
	BuildImage(pixels *PixelBuffer, cellWidth, cellHeight int) backend.Image
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
type Capabilities = terminal.Capabilities

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
	case BlitterKitty:
		return &KittyBlitter{}
	case BlitterSixel:
		return &SixelBlitter{}
	case BlitterASCII:
		return &ASCIIBlitter{}
	case BlitterAuto:
		return BestBlitter(nil)
	default:
		return &SextantBlitter{}
	}
}

// BestBlitter returns the best blitter for the current terminal.
func BestBlitter(caps *Capabilities) Blitter {
	if caps == nil {
		detected := terminal.DetectCapabilities()
		caps = &detected
	}
	if caps.Kitty {
		return &KittyBlitter{}
	}
	if caps.Sixel {
		return &SixelBlitter{}
	}
	if caps.Unicode {
		return &SextantBlitter{}
	}
	return &ASCIIBlitter{}
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
