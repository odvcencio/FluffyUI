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

// PaletteCommand describes an executable action in the command palette.
type PaletteCommand struct {
	ID          string
	Label       string
	Description string
	Shortcut    string // display only, e.g. "Ctrl+S"
	Category    string // for grouping
	OnExecute   func()
}

// CommandPalette is a focusable fuzzy-search command launcher (VS Code-style Ctrl+K/Ctrl+P).
// It displays a filterable list of commands with keyboard navigation.
type CommandPalette struct {
	FocusableBase

	commands   []PaletteCommand
	filtered   []PaletteCommand
	query      string
	selected   int
	visible    bool
	maxResults int // default 10
	services   runtime.Services
	onExecute  func(cmd PaletteCommand) // called after command executes

	style    backend.Style
	styleSet bool
}

// NewCommandPalette creates a command palette with the given commands.
func NewCommandPalette(commands ...PaletteCommand) *CommandPalette {
	cp := &CommandPalette{
		commands:   commands,
		maxResults: 10,
		style:      backend.DefaultStyle(),
	}
	cp.filterCommands()
	cp.Base.Role = accessibility.RoleCombobox
	cp.syncA11y()
	return cp
}

// WithCommandPaletteOpen sets the initial open/visible state.
func WithCommandPaletteOpen(open bool) CommandPaletteOption {
	return func(cp *CommandPalette) {
		if cp == nil {
			return
		}
		cp.visible = open
	}
}

// Show makes the palette visible, resetting the query and selection.
func (cp *CommandPalette) Show() {
	if cp == nil {
		return
	}
	cp.visible = true
	cp.query = ""
	cp.selected = 0
	cp.filterCommands()
	cp.syncA11y()
}

// Hide closes the palette.
func (cp *CommandPalette) Hide() {
	if cp == nil {
		return
	}
	cp.visible = false
	cp.syncA11y()
}

// Toggle switches the palette between visible and hidden.
func (cp *CommandPalette) Toggle() {
	if cp == nil {
		return
	}
	if cp.visible {
		cp.Hide()
	} else {
		cp.Show()
	}
}

// Open returns whether the palette is visible.
func (cp *CommandPalette) Open() bool {
	if cp == nil {
		return false
	}
	return cp.visible
}

// SetOpen sets the visibility of the palette.
func (cp *CommandPalette) SetOpen(open bool) {
	if cp == nil {
		return
	}
	if open {
		cp.Show()
	} else {
		cp.Hide()
	}
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

// SetCommands replaces the command list and refilters.
func (cp *CommandPalette) SetCommands(cmds []PaletteCommand) {
	if cp == nil {
		return
	}
	cp.commands = cmds
	cp.filterCommands()
	cp.syncA11y()
}

// AddCommand appends a single command and refilters.
func (cp *CommandPalette) AddCommand(cmd PaletteCommand) {
	if cp == nil {
		return
	}
	cp.commands = append(cp.commands, cmd)
	cp.filterCommands()
	cp.syncA11y()
}

// SetMaxResults sets the maximum number of results to display.
func (cp *CommandPalette) SetMaxResults(n int) {
	if cp == nil {
		return
	}
	if n < 1 {
		n = 1
	}
	cp.maxResults = n
}

// SetOnExecute sets a callback invoked after a command is executed.
func (cp *CommandPalette) SetOnExecute(fn func(cmd PaletteCommand)) {
	if cp == nil {
		return
	}
	cp.onExecute = fn
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
		if !cp.visible {
			return runtime.Size{}
		}
		width := 30
		if contentConstraints.MaxWidth < width {
			width = contentConstraints.MaxWidth
		}
		// header + search + separator + visible commands (max maxResults) + footer
		visibleItems := len(cp.filtered)
		if visibleItems > cp.maxResults {
			visibleItems = cp.maxResults
		}
		height := 3 + visibleItems + 1 // top border + query + separator + items + bottom border
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the command palette.
func (cp *CommandPalette) Render(ctx runtime.RenderContext) {
	if cp == nil || !cp.visible {
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
	dimStyle := baseStyle.Foreground(backend.ColorDefault)

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
	for i, cmd := range cp.filtered {
		if row >= bounds.Height-1 {
			break
		}
		if i >= cp.maxResults {
			break
		}
		indicator := "  "
		lineStyle := baseStyle
		if i == cp.selected {
			indicator = "> "
			lineStyle = highlightStyle
		}
		label := cmd.Label
		shortcut := cmd.Shortcut
		inner := bounds.Width - 4 // subtract border chars and padding
		if inner < 1 {
			inner = 1
		}
		labelWidth := inner
		if shortcut != "" {
			labelWidth = inner - textWidth(shortcut) - 1
			if labelWidth < 1 {
				labelWidth = 1
			}
		}

		// Build the label portion, optionally with dimmed description
		labelText := truncateString(label, labelWidth)
		if cmd.Description != "" {
			descAvail := labelWidth - textWidth(labelText) - 1
			if descAvail > 3 {
				desc := truncateString(cmd.Description, descAvail)
				// Write label and description separately for styling
				line := indicator + labelText
				full := "\u2502 " + padRight(line, bounds.Width-4) + " \u2502"
				writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(full, bounds.Width), lineStyle)
				// Overlay the description in dim style (if not selected)
				descX := bounds.X + 2 + textWidth(indicator) + textWidth(labelText) + 1
				descStyle := dimStyle
				if i == cp.selected {
					descStyle = highlightStyle
				}
				if descX+textWidth(desc) < bounds.X+bounds.Width-2 {
					ctx.Buffer.SetString(descX, bounds.Y+row, desc, descStyle)
				}
				// Overlay shortcut right-aligned
				if shortcut != "" {
					shortcutX := bounds.X + bounds.Width - 2 - textWidth(shortcut)
					ctx.Buffer.SetString(shortcutX, bounds.Y+row, shortcut, lineStyle)
				}
				row++
				continue
			}
		}

		line := indicator + padRight(labelText, labelWidth)
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

	if !cp.visible {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyEscape:
		cp.Hide()
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
			cmd := cp.filtered[cp.selected]
			cp.Hide()
			if announcer := cp.services.Announcer(); announcer != nil {
				announcer.Announce(fmt.Sprintf("Executed %s", cmd.Label), accessibility.PriorityPolite)
			}
			if cmd.OnExecute != nil {
				cmd.OnExecute()
			}
			if cp.onExecute != nil {
				cp.onExecute(cmd)
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

// filterCommands updates the filtered list based on the current query.
// When the query is empty, all commands are shown (up to maxResults in rendering).
// When a query is provided, commands are matched against Label and Description
// using case-insensitive substring matching, then sorted by relevance:
// exact prefix matches on Label first, then contains matches.
func (cp *CommandPalette) filterCommands() {
	cp.filtered = cp.filtered[:0]
	query := strings.ToLower(cp.query)

	if query == "" {
		// Show all commands when query is empty
		cp.filtered = append(cp.filtered, cp.commands...)
	} else {
		type scored struct {
			cmd   PaletteCommand
			score int
		}
		var matches []scored
		for _, cmd := range cp.commands {
			labelLower := strings.ToLower(cmd.Label)
			descLower := strings.ToLower(cmd.Description)

			if strings.Contains(labelLower, query) || strings.Contains(descLower, query) {
				s := 0
				// Exact prefix match on label scores highest
				if strings.HasPrefix(labelLower, query) {
					s = 3
				} else if strings.Contains(labelLower, query) {
					// Contains in label
					s = 2
				} else {
					// Contains in description only
					s = 1
				}
				matches = append(matches, scored{cmd: cmd, score: s})
			}
		}
		// Sort: higher score first, preserve original order for ties
		for i := 1; i < len(matches); i++ {
			for j := i; j > 0 && matches[j].score > matches[j-1].score; j-- {
				matches[j], matches[j-1] = matches[j-1], matches[j]
			}
		}
		for _, m := range matches {
			cp.filtered = append(cp.filtered, m.cmd)
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
	cmd := cp.filtered[cp.selected]
	if announcer := cp.services.Announcer(); announcer != nil {
		msg := cmd.Label
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
	if cp.visible {
		expanded := true
		cp.Base.State.Expanded = &expanded
	} else {
		expanded := false
		cp.Base.State.Expanded = &expanded
	}
	if len(cp.filtered) > 0 && cp.selected < len(cp.filtered) {
		cmd := cp.filtered[cp.selected]
		cp.Base.Description = cmd.Label
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
