package tcell

// WriteFrame writes raw encoded bytes to the terminal.
func (b *Backend) WriteFrame(encoded []byte) {
	if b == nil || b.raw == nil || len(encoded) == 0 {
		return
	}
	_ = b.raw.WriteRaw(encoded)
}
