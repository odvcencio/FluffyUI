package runtime

import "strings"

// SnapshotText returns a snapshot of the current screen buffer as plain text.
// The snapshot is taken under the render lock to avoid tearing.
func (a *App) SnapshotText() string {
	if a == nil {
		return ""
	}
	a.renderMu.Lock()
	defer a.renderMu.Unlock()

	if a.screen == nil {
		return ""
	}
	buf := a.screen.Buffer()
	if buf == nil {
		return ""
	}
	return buf.SnapshotText()
}

// SnapshotText returns the buffer content as plain text.
// Callers are responsible for external synchronization if needed.
func (b *Buffer) SnapshotText() string {
	if b == nil {
		return ""
	}
	w, h := b.Size()
	if w <= 0 || h <= 0 {
		return ""
	}

	var out strings.Builder
	out.Grow((w + 1) * h)

	for y := 0; y < h; y++ {
		if y > 0 {
			out.WriteByte('\n')
		}
		rowStart := y * w
		row := b.cells[rowStart : rowStart+w]
		for _, cell := range row {
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			if r < 0x80 {
				out.WriteByte(byte(r))
			} else {
				out.WriteRune(r)
			}
		}
	}

	return out.String()
}
