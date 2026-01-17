package runtime

// RegisterFocusables registers focusable widgets from the tree into the scope.
func RegisterFocusables(scope *FocusScope, root Widget) {
	if scope == nil || root == nil {
		return
	}
	registerFocusable(scope, root)
}

func registerFocusable(scope *FocusScope, widget Widget) {
	if widget == nil {
		return
	}
	if focusable, ok := widget.(Focusable); ok {
		scope.Register(focusable)
	}
	if container, ok := widget.(ChildProvider); ok {
		for _, child := range container.ChildWidgets() {
			registerFocusable(scope, child)
		}
	}
}
