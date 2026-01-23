package graphics

import (
	"bytes"
	"testing"

	"github.com/odvcencio/fluffy-ui/backend"
)

func TestEncodeKittyBasic(t *testing.T) {
	img := backend.Image{
		Width:      1,
		Height:     1,
		CellWidth:  1,
		CellHeight: 1,
		Format:     backend.ImageFormatRGBA,
		Protocol:   backend.ImageProtocolKitty,
		Pixels:     []byte{255, 0, 0, 255},
	}
	data := EncodeKitty(img)
	if len(data) == 0 {
		t.Fatalf("expected kitty payload")
	}
	if !bytes.HasPrefix(data, []byte("\x1b_G")) {
		t.Fatalf("kitty payload missing prefix")
	}
	if !bytes.Contains(data, []byte("s=1")) || !bytes.Contains(data, []byte("v=1")) {
		t.Fatalf("kitty payload missing size params")
	}
	if !bytes.HasSuffix(data, []byte("\x1b\\")) {
		t.Fatalf("kitty payload missing terminator")
	}
}

func TestEncodeSixelBasic(t *testing.T) {
	img := backend.Image{
		Width:      1,
		Height:     1,
		CellWidth:  1,
		CellHeight: 1,
		Format:     backend.ImageFormatRGBA,
		Protocol:   backend.ImageProtocolSixel,
		Pixels:     []byte{0, 255, 0, 255},
	}
	data := EncodeSixel(img)
	if len(data) == 0 {
		t.Fatalf("expected sixel payload")
	}
	if !bytes.HasPrefix(data, []byte("\x1bPq")) {
		t.Fatalf("sixel payload missing prefix")
	}
	if !bytes.HasSuffix(data, []byte("\x1b\\")) {
		t.Fatalf("sixel payload missing terminator")
	}
}
