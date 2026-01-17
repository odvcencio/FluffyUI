// Package dragdrop provides drag-and-drop interfaces.
package dragdrop

import "github.com/odvcencio/fluffy-ui/runtime"

// DropPosition describes a drop location.
type DropPosition struct {
	X int
	Y int
}

// DragData contains drag payload metadata.
type DragData struct {
	Source  runtime.Widget
	Kind    string
	Payload any
}

// Draggable provides drag behavior for widgets.
type Draggable interface {
	DragStart() DragData
	DragEnd(cancelled bool)
}

// DropTarget receives drops.
type DropTarget interface {
	CanDrop(data DragData) bool
	Drop(data DragData, position DropPosition)
	DropPreview(data DragData, position DropPosition)
}
