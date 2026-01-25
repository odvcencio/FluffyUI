//go:build linux || darwin || windows

package ghostty

type ghosttyInputKey struct {
	Action             int32
	Mods               int32
	ConsumedMods       int32
	Keycode            uint32
	Text               *byte
	UnshiftedCodepoint uint32
	Composing          bool
}

type ghosttySurfaceSize struct {
	Columns      uint16
	Rows         uint16
	WidthPx      uint32
	HeightPx     uint32
	CellWidthPx  uint32
	CellHeightPx uint32
}
