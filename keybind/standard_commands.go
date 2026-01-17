package keybind

import (
	"github.com/odvcencio/fluffy-ui/clipboard"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/scroll"
)

// RegisterStandardCommands registers common runtime commands.
func RegisterStandardCommands(registry *CommandRegistry) {
	if registry == nil {
		return
	}
	registry.RegisterAll(
		Command{
			ID:          "app.quit",
			Title:       "Quit",
			Description: "Exit the application",
			Category:    "App",
			Handler: func(ctx Context) {
				if ctx.App != nil {
					ctx.App.ExecuteCommand(runtime.Quit{})
				}
			},
		},
		Command{
			ID:          "app.refresh",
			Title:       "Refresh",
			Description: "Force a redraw",
			Category:    "App",
			Handler: func(ctx Context) {
				if ctx.App != nil {
					ctx.App.ExecuteCommand(runtime.Refresh{})
				}
			},
		},
		Command{
			ID:          "focus.next",
			Title:       "Next Focus",
			Description: "Move focus to the next widget",
			Category:    "Focus",
			Handler: func(ctx Context) {
				if ctx.App != nil {
					ctx.App.ExecuteCommand(runtime.FocusNext{})
				}
			},
		},
		Command{
			ID:          "focus.prev",
			Title:       "Previous Focus",
			Description: "Move focus to the previous widget",
			Category:    "Focus",
			Handler: func(ctx Context) {
				if ctx.App != nil {
					ctx.App.ExecuteCommand(runtime.FocusPrev{})
				}
			},
		},
		Command{
			ID:          "overlay.pop",
			Title:       "Close Overlay",
			Description: "Dismiss the top overlay",
			Category:    "Overlay",
			Handler: func(ctx Context) {
				if ctx.App != nil {
					ctx.App.ExecuteCommand(runtime.PopOverlay{})
				}
			},
		},
	)
}

// RegisterScrollCommands registers scroll commands for focused widgets.
func RegisterScrollCommands(registry *CommandRegistry) {
	if registry == nil {
		return
	}
	scrollBy := func(dx, dy int) func(Context) {
		return func(ctx Context) {
			if ctx.Focused == nil {
				return
			}
			if controller, ok := ctx.Focused.(scroll.Controller); ok {
				controller.ScrollBy(dx, dy)
				if ctx.App != nil {
					ctx.App.Invalidate()
				}
			}
		}
	}
	pageBy := func(pages int) func(Context) {
		return func(ctx Context) {
			if ctx.Focused == nil {
				return
			}
			if controller, ok := ctx.Focused.(scroll.Controller); ok {
				controller.PageBy(pages)
				if ctx.App != nil {
					ctx.App.Invalidate()
				}
			}
		}
	}
	registry.RegisterAll(
		Command{ID: "scroll.up", Title: "Scroll Up", Category: "Scroll", Handler: scrollBy(0, -1)},
		Command{ID: "scroll.down", Title: "Scroll Down", Category: "Scroll", Handler: scrollBy(0, 1)},
		Command{ID: "scroll.left", Title: "Scroll Left", Category: "Scroll", Handler: scrollBy(-1, 0)},
		Command{ID: "scroll.right", Title: "Scroll Right", Category: "Scroll", Handler: scrollBy(1, 0)},
		Command{ID: "scroll.pageUp", Title: "Page Up", Category: "Scroll", Handler: pageBy(-1)},
		Command{ID: "scroll.pageDown", Title: "Page Down", Category: "Scroll", Handler: pageBy(1)},
		Command{
			ID:       "scroll.home",
			Title:    "Scroll Home",
			Category: "Scroll",
			Handler: func(ctx Context) {
				if controller, ok := ctx.Focused.(scroll.Controller); ok {
					controller.ScrollToStart()
					if ctx.App != nil {
						ctx.App.Invalidate()
					}
				}
			},
		},
		Command{
			ID:       "scroll.end",
			Title:    "Scroll End",
			Category: "Scroll",
			Handler: func(ctx Context) {
				if controller, ok := ctx.Focused.(scroll.Controller); ok {
					controller.ScrollToEnd()
					if ctx.App != nil {
						ctx.App.Invalidate()
					}
				}
			},
		},
	)
}

// RegisterClipboardCommands registers clipboard commands for focused widgets.
func RegisterClipboardCommands(registry *CommandRegistry) {
	if registry == nil {
		return
	}
	registry.RegisterAll(
		Command{
			ID:       clipboard.CommandCopy,
			Title:    "Copy",
			Category: "Clipboard",
			Handler: func(ctx Context) {
				target, ok := ctx.Focused.(clipboard.Target)
				if !ok || ctx.App == nil {
					return
				}
				cb := ctx.App.Services().Clipboard()
				if cb == nil || !cb.Available() {
					return
				}
				if text, ok := target.ClipboardCopy(); ok {
					_ = cb.Write(text)
					if ctx.App != nil {
						ctx.App.Invalidate()
					}
				}
			},
		},
		Command{
			ID:       clipboard.CommandCut,
			Title:    "Cut",
			Category: "Clipboard",
			Handler: func(ctx Context) {
				target, ok := ctx.Focused.(clipboard.Target)
				if !ok || ctx.App == nil {
					return
				}
				cb := ctx.App.Services().Clipboard()
				if cb == nil || !cb.Available() {
					return
				}
				if text, ok := target.ClipboardCut(); ok {
					_ = cb.Write(text)
					if ctx.App != nil {
						ctx.App.Invalidate()
					}
				}
			},
		},
		Command{
			ID:       clipboard.CommandPaste,
			Title:    "Paste",
			Category: "Clipboard",
			Handler: func(ctx Context) {
				target, ok := ctx.Focused.(clipboard.Target)
				if !ok || ctx.App == nil {
					return
				}
				cb := ctx.App.Services().Clipboard()
				if cb == nil || !cb.Available() {
					return
				}
				if text, err := cb.Read(); err == nil {
					if target.ClipboardPaste(text) && ctx.App != nil {
						ctx.App.Invalidate()
					}
				}
			},
		},
	)
}
