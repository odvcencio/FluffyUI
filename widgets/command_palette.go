package widgets

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// CommandPaletteOption configures a CommandPalette widget.
type CommandPaletteOption = Option[CommandPalette]

// Command describes an executable action in the command palette.
type Command struct {
	Name        string
	Description string
	Action      func()
	Shortcut    string // display only, e.g. "Ctrl+S"
}

// CommandPalette is a focusable fuzzy-search command launcher.
// It displays a filterable list of commands with keyboard navigation.
type CommandPalette struct {
	FocusableBase

	commands []Command
	filtered []int // indices into commands
	query    string
	selected int
	open     bool
	services runtime.Services

	style    backend.Style
	styleSet bool
}

// NewCommandPalette creates a command palette with the given commands.
func NewCommandPalette(commands []Command, opts ...CommandPaletteOption) *CommandPalette {
	cp := &CommandPalette{
		commands: commands,
		style:    backend.DefaultStyle(),
	}
	cp.filterCommands()
	for _, opt := range opts {
		opt(cp)
	}
	cp.syncA11y()
	return cp
}

// WithCommandPaletteOpen sets the initial open state.
func WithCommandPaletteOpen(open bool) CommandPaletteOption {
	return func(cp *CommandPalette) {
		cp.open = open
	}
}

// Open returns whether the palette is visible.
func (cp *CommandPalette) Open() bool {
	if cp == nil {
		return false
	}
	return cp.open
}

// SetOpen sets the visibility of the palette.
func (cp *CommandPalette) SetOpen(open bool) {
	if cp == nil {
		return
	}
	cp.open = open
	if open {
		cp.query = ""
		cp.selected = 0
		cp.filterCommands()
	}
	cp.syncA11y()
}

// Query returns the current search query.
func (cp *CommandPalette) Query() string {
	if cp == nil {
		return ""
	}
	return cp.query
}

// SelectedIndex returns the index within the filtered results.
func (cp *CommandPalette) SelectedIndex() int {
	if cp == nil {
		return 0
	}
	return cp.selected
}

// FilteredCount returns the number of matching commands.
func (cp *CommandPalette) FilteredCount() int {
	if cp == nil {
		return 0
	}
	return len(cp.filtered)
}

// SetCommands replaces the command list.
func (cp *CommandPalette) SetCommands(commands []Command) {
	if cp == nil {
		return
	}
	cp.commands = commands
	cp.filterCommands()
	cp.syncA11y()
}

// SetStyle sets the widget style.
func (cp *CommandPalette) SetStyle(style backend.Style) {
	if cp == nil {
		return
	}
	cp.style = style
	cp.styleSet = true
}

// StyleType returns the selector type name.
func (cp *CommandPalette) StyleType() string {
	return "CommandPalette"
}

// Bind attaches app services.
func (cp *CommandPalette) Bind(services runtime.Services) {
	if cp == nil {
		return
	}
	cp.services = services
}

// Unbind releases app services.
func (cp *CommandPalette) Unbind() {
	if cp == nil {
		return
	}
	cp.services = runtime.Services{}
}

// Measure returns the desired size.
func (cp *CommandPalette) Measure(constraints runtime.Constraints) runtime.Size {
	return cp.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if !cp.open {
			return runtime.Size{}
		}
		width := 30
		if contentConstraints.MaxWidth < width {
			width = contentConstraints.MaxWidth
		}
		// header + search + separator + visible commands (max 10) + footer
		visibleItems := len(cp.filtered)
		if visibleItems > 10 {
			visibleItems = 10
		}
		height := 3 + visibleItems + 1 // top border + query + separator + items + bottom border
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the command palette.
func (cp *CommandPalette) Render(ctx runtime.RenderContext) {
	if cp == nil || !cp.open {
		return
	}
	cp.syncA11y()
	bounds := cp.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, cp, cp.style, cp.styleSet)
	borderStyle := baseStyle.Foreground(backend.ColorCyan)
	highlightStyle := baseStyle.Reverse(true)

	row := 0

	// Top border
	if row < bounds.Height {
		top := "\u256d\u2500 Commands " + strings.Repeat("\u2500", max(0, bounds.Width-13)) + "\u256e"
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(top, bounds.Width), borderStyle)
		row++
	}

	// Search input
	if row < bounds.Height {
		prompt := "\u2502 > " + cp.query
		prompt = padRight(prompt, bounds.Width-1) + "\u2502"
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(prompt, bounds.Width), baseStyle)
		row++
	}

	// Separator
	if row < bounds.Height {
		sep := "\u2502 " + strings.Repeat("\u2504", max(0, bounds.Width-4)) + " \u2502"
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(sep, bounds.Width), borderStyle)
		row++
	}

	// Command list
	for i, cmdIdx := range cp.filtered {
		if row >= bounds.Height-1 {
			break
		}
		if i >= 10 {
			break
		}
		cmd := cp.commands[cmdIdx]
		indicator := "  "
		lineStyle := baseStyle
		if i == cp.selected {
			indicator = "> "
			lineStyle = highlightStyle
		}
		name := cmd.Name
		shortcut := cmd.Shortcut
		inner := bounds.Width - 4 // subtract border chars and padding
		if inner < 1 {
			inner = 1
		}
		nameWidth := inner
		if shortcut != "" {
			nameWidth = inner - textWidth(shortcut) - 1
			if nameWidth < 1 {
				nameWidth = 1
			}
		}
		line := indicator + padRight(truncateString(name, nameWidth), nameWidth)
		if shortcut != "" {
			line += " " + shortcut
		}
		full := "\u2502 " + padRight(line, max(0, bounds.Width-4)) + " \u2502"
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(full, bounds.Width), lineStyle)
		row++
	}

	// Fill empty space
	for row < bounds.Height-1 {
		empty := "\u2502" + strings.Repeat(" ", max(0, bounds.Width-2)) + "\u2502"
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(empty, bounds.Width), borderStyle)
		row++
	}

	// Bottom border
	if row < bounds.Height {
		bottom := "\u2570" + strings.Repeat("\u2500", max(0, bounds.Width-2)) + "\u256f"
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(bottom, bounds.Width), borderStyle)
	}
}

// HandleMessage processes keyboard input.
func (cp *CommandPalette) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if cp == nil || !cp.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	if !cp.open {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyEscape:
		cp.SetOpen(false)
		if announcer := cp.services.Announcer(); announcer != nil {
			announcer.Announce("Command palette closed", accessibility.PriorityPolite)
		}
		return runtime.Handled()

	case terminal.KeyUp:
		if cp.selected > 0 {
			cp.selected--
			cp.announceSelected()
			cp.syncA11y()
		}
		return runtime.Handled()

	case terminal.KeyDown:
		if cp.selected < len(cp.filtered)-1 {
			cp.selected++
			cp.announceSelected()
			cp.syncA11y()
		}
		return runtime.Handled()

	case terminal.KeyEnter:
		if cp.selected >= 0 && cp.selected < len(cp.filtered) {
			cmdIdx := cp.filtered[cp.selected]
			cmd := cp.commands[cmdIdx]
			cp.SetOpen(false)
			if announcer := cp.services.Announcer(); announcer != nil {
				announcer.Announce(fmt.Sprintf("Executed %s", cmd.Name), accessibility.PriorityPolite)
			}
			if cmd.Action != nil {
				cmd.Action()
			}
		}
		return runtime.Handled()

	case terminal.KeyBackspace:
		if len(cp.query) > 0 {
			_, size := utf8.DecodeLastRuneInString(cp.query)
			cp.query = cp.query[:len(cp.query)-size]
			cp.filterCommands()
			cp.syncA11y()
		}
		return runtime.Handled()

	case terminal.KeyRune:
		cp.query += string(key.Rune)
		cp.filterCommands()
		cp.syncA11y()
		return runtime.Handled()
	}

	return runtime.Unhandled()
}

func (cp *CommandPalette) filterCommands() {
	cp.filtered = cp.filtered[:0]
	query := strings.ToLower(cp.query)
	for i, cmd := range cp.commands {
		if query == "" || strings.Contains(strings.ToLower(cmd.Name), query) {
			cp.filtered = append(cp.filtered, i)
		}
	}
	if cp.selected >= len(cp.filtered) {
		cp.selected = max(0, len(cp.filtered)-1)
	}
}

func (cp *CommandPalette) announceSelected() {
	if cp.selected < 0 || cp.selected >= len(cp.filtered) {
		return
	}
	cmdIdx := cp.filtered[cp.selected]
	cmd := cp.commands[cmdIdx]
	if announcer := cp.services.Announcer(); announcer != nil {
		msg := cmd.Name
		if cmd.Shortcut != "" {
			msg += ", " + cmd.Shortcut
		}
		announcer.Announce(
			fmt.Sprintf("%s, %d of %d", msg, cp.selected+1, len(cp.filtered)),
			accessibility.PriorityPolite,
		)
	}
}

func (cp *CommandPalette) syncA11y() {
	if cp == nil {
		return
	}
	if cp.Base.Role == "" {
		cp.Base.Role = accessibility.RoleCombobox
	}
	cp.Base.Label = "Command Palette"
	cp.Base.HasPopup = "listbox"
	cp.Base.Autocomplete = "list"
	cp.Base.Placeholder = "Type a command..."
	if cp.open {
		expanded := true
		cp.Base.State.Expanded = &expanded
	} else {
		expanded := false
		cp.Base.State.Expanded = &expanded
	}
	if len(cp.filtered) > 0 && cp.selected < len(cp.filtered) {
		cmdIdx := cp.filtered[cp.selected]
		cp.Base.Description = cp.commands[cmdIdx].Name
		cp.Base.PosInSet = cp.selected + 1
		cp.Base.SetSize = len(cp.filtered)
	} else {
		cp.Base.Description = ""
		cp.Base.PosInSet = 0
		cp.Base.SetSize = 0
	}
}

var _ runtime.Widget = (*CommandPalette)(nil)
var _ runtime.Focusable = (*CommandPalette)(nil)
var _ runtime.Bindable = (*CommandPalette)(nil)
var _ runtime.Unbindable = (*CommandPalette)(nil)
