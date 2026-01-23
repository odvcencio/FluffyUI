package tcell

import (
	"bytes"
	"fmt"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/graphics"
)

// DrawImage renders a pixel image using Kitty or Sixel protocols.
func (b *Backend) DrawImage(x, y int, img backend.Image) {
	if b == nil || b.raw == nil {
		return
	}
	var payload []byte
	switch img.Protocol {
	case backend.ImageProtocolKitty:
		payload = graphics.EncodeKitty(img)
	case backend.ImageProtocolSixel:
		payload = graphics.EncodeSixel(img)
	default:
		return
	}
	if len(payload) == 0 {
		return
	}
	var out bytes.Buffer
	out.WriteString("\x1b[s")
	out.WriteString(cursorTo(x, y))
	out.Write(payload)
	out.WriteString("\x1b[u")
	_ = b.raw.WriteRaw(out.Bytes())
}

func cursorTo(x, y int) string {
	return fmt.Sprintf("\x1b[%d;%dH", y+1, x+1)
}
