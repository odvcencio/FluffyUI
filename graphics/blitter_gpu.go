package graphics

import (
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/gpu"
)

// TerminalEncoder converts raw RGBA pixels into terminal protocol bytes.
type TerminalEncoder interface {
	Encode(pixels []byte, width, height int) []byte
	SupportsAlpha() bool
	Name() string
}

// GPUBlitter converts GPU canvas frames into terminal images.
type GPUBlitter struct {
	canvas  *gpu.GPUCanvas
	encoder TerminalEncoder
	cellW   int
	cellH   int
}

// NewGPUBlitter creates a GPU blitter.
func NewGPUBlitter(canvas *gpu.GPUCanvas, encoder TerminalEncoder, cellW, cellH int) *GPUBlitter {
	return &GPUBlitter{canvas: canvas, encoder: encoder, cellW: cellW, cellH: cellH}
}

// SetCanvas updates the canvas.
func (b *GPUBlitter) SetCanvas(canvas *gpu.GPUCanvas) {
	if b == nil {
		return
	}
	b.canvas = canvas
}

// SetEncoder updates the encoder.
func (b *GPUBlitter) SetEncoder(encoder TerminalEncoder) {
	if b == nil {
		return
	}
	b.encoder = encoder
}

// SetCellSize updates the cell dimensions.
func (b *GPUBlitter) SetCellSize(cellW, cellH int) {
	if b == nil {
		return
	}
	b.cellW = cellW
	b.cellH = cellH
}

// Image builds a backend image from the canvas.
func (b *GPUBlitter) Image() backend.Image {
	if b == nil || b.canvas == nil {
		return backend.Image{}
	}
	pixels := b.canvas.End()
	w, h := b.canvas.Size()
	if w <= 0 || h <= 0 || len(pixels) == 0 {
		return backend.Image{}
	}
	protocol := backend.ImageProtocolKitty
	if b.encoder != nil && b.encoder.Name() == "sixel" {
		protocol = backend.ImageProtocolSixel
	}
	return backend.Image{
		Width:      w,
		Height:     h,
		CellWidth:  b.cellW,
		CellHeight: b.cellH,
		Format:     backend.ImageFormatRGBA,
		Protocol:   protocol,
		Pixels:     pixels,
	}
}

// Encode encodes the canvas frame into terminal bytes.
func (b *GPUBlitter) Encode() []byte {
	if b == nil || b.encoder == nil || b.canvas == nil {
		return nil
	}
	pixels := b.canvas.End()
	w, h := b.canvas.Size()
	if w <= 0 || h <= 0 || len(pixels) == 0 {
		return nil
	}
	return b.encoder.Encode(pixels, w, h)
}

// KittyEncoder encodes using the Kitty graphics protocol.
type KittyEncoder struct {
	CellWidth  int
	CellHeight int
}

func (k KittyEncoder) Encode(pixels []byte, width, height int) []byte {
	img := backend.Image{
		Width:      width,
		Height:     height,
		CellWidth:  k.cellWidth(),
		CellHeight: k.cellHeight(),
		Format:     backend.ImageFormatRGBA,
		Protocol:   backend.ImageProtocolKitty,
		Pixels:     pixels,
	}
	return EncodeKitty(img)
}

func (k KittyEncoder) SupportsAlpha() bool { return true }
func (k KittyEncoder) Name() string        { return "kitty" }

func (k KittyEncoder) cellWidth() int {
	if k.CellWidth > 0 {
		return k.CellWidth
	}
	return DefaultKittyCellWidth
}

func (k KittyEncoder) cellHeight() int {
	if k.CellHeight > 0 {
		return k.CellHeight
	}
	return DefaultKittyCellHeight
}

// SixelEncoder encodes using the Sixel graphics protocol.
type SixelEncoder struct {
	CellWidth  int
	CellHeight int
}

func (s SixelEncoder) Encode(pixels []byte, width, height int) []byte {
	img := backend.Image{
		Width:      width,
		Height:     height,
		CellWidth:  s.cellWidth(),
		CellHeight: s.cellHeight(),
		Format:     backend.ImageFormatRGBA,
		Protocol:   backend.ImageProtocolSixel,
		Pixels:     pixels,
	}
	return EncodeSixel(img)
}

func (s SixelEncoder) SupportsAlpha() bool { return false }
func (s SixelEncoder) Name() string        { return "sixel" }

func (s SixelEncoder) cellWidth() int {
	if s.CellWidth > 0 {
		return s.CellWidth
	}
	return DefaultSixelCellWidth
}

func (s SixelEncoder) cellHeight() int {
	if s.CellHeight > 0 {
		return s.CellHeight
	}
	return DefaultSixelCellHeight
}
