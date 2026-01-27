package backend

// FrameWriter writes raw frame bytes to the backend.
type FrameWriter interface {
	WriteFrame(encoded []byte)
}
