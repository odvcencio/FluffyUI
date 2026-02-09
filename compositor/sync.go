package compositor

// Synchronized output support (DEC private mode 2026).
//
// Modern terminals support "synchronized updates" which eliminates screen
// tearing during frame rendering. When enabled, the terminal buffers all
// output between the begin and end markers, then applies the changes
// atomically in a single
//
// Sequence: CSI ? 2026 h  (begin)  /  CSI ? 2026 l  (end)
//
// Terminals that don't recognize the sequence ignore it harmlessly, so
// emitting the markers unconditionally is safe. However, FluffyUI's
// Renderer can be configured to skip them when the terminal is known
// to lack support (see Renderer.SetSyncOutput).

// SyncBegin returns the escape sequence to begin a synchronized update.
// DEC mode 2026: \x1b[?2026h
func SyncBegin() string { return ANSISyncStart }

// SyncEnd returns the escape sequence to end a synchronized update.
// DEC mode 2026: \x1b[?2026l
func SyncEnd() string { return ANSISyncEnd }

// WrapSync wraps content in synchronized output markers.
// If syncEnabled is false, the content is returned unchanged.
func WrapSync(content string, syncEnabled bool) string {
	if !syncEnabled || content == "" {
		return content
	}
	return SyncBegin() + content + SyncEnd()
}
