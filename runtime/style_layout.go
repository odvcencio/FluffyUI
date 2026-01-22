package runtime

import "github.com/odvcencio/fluffy-ui/style"

// StyleApplier receives resolved stylesheet styles for layout.
type StyleApplier interface {
	ApplyStyle(style.Style)
}

func applyLayoutStyles(root Widget, resolver *StyleResolver, focused bool) {
	if root == nil {
		return
	}
	var walk func(node Widget)
	walk = func(node Widget) {
		if node == nil {
			return
		}
		if applier, ok := node.(StyleApplier); ok {
			var resolved style.Style
			if resolver != nil {
				resolved = resolver.Resolve(node, focused)
			}
			applier.ApplyStyle(resolved)
		}
		if container, ok := node.(ChildProvider); ok {
			for _, child := range container.ChildWidgets() {
				walk(child)
			}
		}
	}
	walk(root)
}
