package recording

import (
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/compositor"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/style"
)

// ANSIEncoder encodes buffer updates as ANSI sequences.
type ANSIEncoder struct {
	styleCache map[backend.Style]compositor.Style
}

// NewANSIEncoder creates a new encoder.
func NewANSIEncoder() *ANSIEncoder {
	return &ANSIEncoder{
		styleCache: make(map[backend.Style]compositor.Style),
	}
}

// Encode builds ANSI output for the buffer.
func (e *ANSIEncoder) Encode(buffer *runtime.Buffer, full bool) string {
	if buffer == nil {
		return ""
	}
	if !full && buffer.DirtyCount() == 0 {
		return ""
	}
	writer := compositor.NewANSIWriter()
	if full {
		writer.WriteString(compositor.ANSIClearScreen)
		writer.WriteString(compositor.ANSICursorHome)
	}
	writer.HideCursor()
	if full {
		w, h := buffer.Size()
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				cell := buffer.Get(x, y)
				e.writeCell(writer, x, y, cell)
			}
		}
	} else {
		buffer.ForEachDirtyCell(func(x, y int, cell runtime.Cell) {
			e.writeCell(writer, x, y, cell)
		})
	}
	writer.Reset()
	return writer.String()
}

func (e *ANSIEncoder) writeCell(writer *compositor.ANSIWriter, x, y int, cell runtime.Cell) {
	writer.MoveTo(x, y)
	writer.SetStyle(e.toCompositor(cell.Style))
	r := cell.Rune
	if r == 0 {
		r = ' '
	}
	writer.WriteRune(r)
}

func (e *ANSIEncoder) toCompositor(bs backend.Style) compositor.Style {
	if cached, ok := e.styleCache[bs]; ok {
		return cached
	}
	cs := style.ToCompositor(bs)
	e.styleCache[bs] = cs
	return cs
}
