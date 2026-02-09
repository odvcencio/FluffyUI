package backend

// SyncOutputWriter is an optional interface for backends that support
// DEC private mode 2026 (synchronized output). When the terminal
// supports this mode, wrapping a frame update in BeginSync/EndSync
// causes the terminal to buffer all output and apply it atomically,
// eliminating screen tearing.
//
// The runtime detects terminal capabilities at startup and calls
// SetSyncOutput(true) when the terminal is known to support mode 2026.
// The render loop then checks SyncOutputSupported before each frame
// and wraps the update in BeginSync/EndSync.
//
// Usage:
//
//	if sw, ok := backend.(SyncOutputWriter); ok && sw.SyncOutputSupported() {
//	    sw.BeginSync()
//	    defer sw.EndSync()
//	}
//	// ... send frame data ...
//	backend.Show()
type SyncOutputWriter interface {
	// SetSyncOutput enables or disables synchronized output support.
	// Call this after detecting terminal capabilities.
	SetSyncOutput(enabled bool)

	// SyncOutputSupported reports whether the terminal supports
	// synchronized output (DEC mode 2026).
	SyncOutputSupported() bool

	// BeginSync writes the DEC mode 2026 enable sequence (\x1b[?2026h)
	// to the terminal, starting a synchronized update block.
	BeginSync()

	// EndSync writes the DEC mode 2026 disable sequence (\x1b[?2026l)
	// to the terminal, ending a synchronized update block. The terminal
	// flushes all buffered output at this point.
	EndSync()
}
