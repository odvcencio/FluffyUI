package runtime

// Bindable widgets receive app services when mounted into a screen.
type Bindable interface {
	Bind(services Services)
}

// Unbindable widgets release app services when removed.
type Unbindable interface {
	Unbind()
}

// MountHook is implemented by widgets that support an onMount lifecycle callback.
// BindTree calls OnMountHook after Bind so hooks fire even when a widget
// defines its own Bind that does not chain to Base.
type MountHook interface {
	OnMountHook()
}

// UnmountHook is implemented by widgets that support an onUnmount lifecycle callback.
// UnbindTree calls OnUnmountHook before Unbind.
type UnmountHook interface {
	OnUnmountHook()
}

// BindTree calls Bind on widgets that implement Bindable.
func BindTree(root Widget, services Services) {
	if services.isZero() {
		return
	}
	bindWidget(root, services)
}

// UnbindTree calls Unbind on widgets that implement Unbindable.
func UnbindTree(root Widget) {
	unbindWidget(root)
}

func bindWidget(w Widget, services Services) {
	if w == nil {
		return
	}
	if b, ok := w.(Bindable); ok {
		b.Bind(services)
	}
	if h, ok := w.(MountHook); ok {
		h.OnMountHook()
	}
	if children, ok := w.(ChildProvider); ok {
		for _, child := range children.ChildWidgets() {
			bindWidget(child, services)
		}
	}
}

func unbindWidget(w Widget) {
	if w == nil {
		return
	}
	if children, ok := w.(ChildProvider); ok {
		for _, child := range children.ChildWidgets() {
			unbindWidget(child)
		}
	}
	if h, ok := w.(UnmountHook); ok {
		h.OnUnmountHook()
	}
	if u, ok := w.(Unbindable); ok {
		u.Unbind()
	}
}
