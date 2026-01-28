package keybind

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/runtime"
)

// WhenFocused returns true when a widget is focused.
func WhenFocused() Condition {
	return func(ctx Context) bool {
		return ctx.Focused != nil
	}
}

// WhenFocusedWidget checks a predicate against the focused widget.
func WhenFocusedWidget(fn func(runtime.Widget) bool) Condition {
	return func(ctx Context) bool {
		if ctx.Focused == nil || fn == nil {
			return false
		}
		return fn(ctx.Focused)
	}
}

// WhenFocusedAccessible checks a predicate against the focused accessible widget.
func WhenFocusedAccessible(fn func(accessibility.Accessible) bool) Condition {
	return func(ctx Context) bool {
		if ctx.FocusedWidget == nil || fn == nil {
			return false
		}
		return fn(ctx.FocusedWidget)
	}
}

// WhenFocusedRole matches a specific accessibility role.
func WhenFocusedRole(role accessibility.Role) Condition {
	return WhenFocusedAccessible(func(widget accessibility.Accessible) bool {
		return widget.AccessibleRole() == role
	})
}

// WhenFocusedClipboardTarget matches widgets that support clipboard operations.
func WhenFocusedClipboardTarget() Condition {
	return func(ctx Context) bool {
		_, ok := ctx.Focused.(clipboard.Target)
		return ok
	}
}

// WhenFocusedNotClipboardTarget matches widgets that do not support clipboard operations.
func WhenFocusedNotClipboardTarget() Condition {
	return func(ctx Context) bool {
		if ctx.Focused == nil {
			return true
		}
		_, ok := ctx.Focused.(clipboard.Target)
		return !ok
	}
}

// All combines conditions with logical AND.
func All(conditions ...Condition) Condition {
	return func(ctx Context) bool {
		for _, cond := range conditions {
			if cond == nil {
				continue
			}
			if !cond(ctx) {
				return false
			}
		}
		return true
	}
}

// Any combines conditions with logical OR.
func Any(conditions ...Condition) Condition {
	return func(ctx Context) bool {
		for _, cond := range conditions {
			if cond == nil {
				continue
			}
			if cond(ctx) {
				return true
			}
		}
		return false
	}
}

// Not negates a condition.
func Not(condition Condition) Condition {
	return func(ctx Context) bool {
		if condition == nil {
			return true
		}
		return !condition(ctx)
	}
}
