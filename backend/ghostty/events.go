//go:build linux || darwin || windows

package ghostty

type ghosttyEventType int32

const (
	ghosttyEventNone        ghosttyEventType = 0
	ghosttyEventRender      ghosttyEventType = 1
	ghosttyEventResize      ghosttyEventType = 2
	ghosttyEventKey         ghosttyEventType = 3
	ghosttyEventMouseButton ghosttyEventType = 4
	ghosttyEventMouseMove   ghosttyEventType = 5
	ghosttyEventMouseScroll ghosttyEventType = 6
)

type ghosttyEventResizeData struct {
	Columns uint16
	Rows    uint16
}

type ghosttyEventKeyData struct {
	Action int32
	Key    int32
	Rune   uint32
	Mods   int32
}

type ghosttyEventMouseData struct {
	X       int32
	Y       int32
	Button  int32
	State   int32
	Mods    int32
	ScrollX float64
	ScrollY float64
}

type ghosttyEvent struct {
	Tag    ghosttyEventType
	Resize ghosttyEventResizeData
	Key    ghosttyEventKeyData
	Mouse  ghosttyEventMouseData
}
