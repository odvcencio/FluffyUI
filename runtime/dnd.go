package runtime

// DragData represents data being dragged.
type DragData struct {
	Source Widget      // the widget being dragged from
	Kind   string      // type identifier (e.g., "list-item", "tree-node")
	Value  interface{} // the dragged data
	Label  string      // display label for the drag indicator
}

// DragStartMsg is a command sent when a drag begins.
type DragStartMsg struct{ Data DragData }

// Command implements the Command interface.
func (DragStartMsg) Command() {}

// DragEndMsg is a command sent when a drag completes or is cancelled.
type DragEndMsg struct {
	Data      DragData
	Target    Widget // nil if cancelled
	Cancelled bool
}

// Command implements the Command interface.
func (DragEndMsg) Command() {}

// DropTarget is implemented by widgets that accept drops.
type DropTarget interface {
	// CanDrop returns true if this widget accepts the given drag data.
	CanDrop(data DragData) bool
	// OnDrop handles the drop, returning true if accepted.
	OnDrop(data DragData) bool
}

// DragSource is implemented by widgets that can initiate drags.
type DragSource interface {
	// IsDraggable returns true if dragging is enabled on this widget.
	IsDraggable() bool
}
