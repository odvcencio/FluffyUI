package widgets

import (
	"fmt"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// ColorPickerOption configures a ColorPicker widget.
type ColorPickerOption = Option[ColorPicker]

// colorEntry pairs a backend color with a display name.
type colorEntry struct {
	Color backend.Color
	Name  string
}

const colorPickerColumns = 6

// ColorPicker is a focusable color selection widget.
// It displays a grid of preset color swatches and allows keyboard navigation.
type ColorPicker struct {
	FocusableBase

	palette  []colorEntry
	selected int
	onChange func(backend.Color)
	services runtime.Services

	style    backend.Style
	styleSet bool
}

// NewColorPicker creates a color picker with the default palette.
func NewColorPicker(opts ...ColorPickerOption) *ColorPicker {
	cp := &ColorPicker{
		palette:  defaultColorPalette(),
		selected: 0,
		style:    backend.DefaultStyle(),
	}
	for _, opt := range opts {
		opt(cp)
	}
	cp.syncA11y()
	return cp
}

// WithColorPickerOnChange sets the change handler.
func WithColorPickerOnChange(fn func(backend.Color)) ColorPickerOption {
	return func(cp *ColorPicker) {
		cp.onChange = fn
	}
}

// WithColorPickerPalette sets custom palette colors.
func WithColorPickerPalette(colors []backend.Color, names []string) ColorPickerOption {
	return func(cp *ColorPicker) {
		entries := make([]colorEntry, len(colors))
		for i, c := range colors {
			name := fmt.Sprintf("Color %d", i)
			if i < len(names) && names[i] != "" {
				name = names[i]
			}
			entries[i] = colorEntry{Color: c, Name: name}
		}
		cp.palette = entries
	}
}

// Selected returns the currently selected color.
func (cp *ColorPicker) Selected() backend.Color {
	if cp == nil || len(cp.palette) == 0 {
		return backend.ColorDefault
	}
	return cp.palette[cp.selected].Color
}

// SelectedIndex returns the index of the selected color.
func (cp *ColorPicker) SelectedIndex() int {
	if cp == nil {
		return 0
	}
	return cp.selected
}

// SetSelected sets the selected index.
func (cp *ColorPicker) SetSelected(index int) {
	if cp == nil || len(cp.palette) == 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(cp.palette) {
		index = len(cp.palette) - 1
	}
	old := cp.selected
	cp.selected = index
	cp.syncA11y()
	if old != index {
		if cp.onChange != nil {
			cp.onChange(cp.palette[index].Color)
		}
		if announcer := cp.services.Announcer(); announcer != nil {
			announcer.Announce(
				fmt.Sprintf("Selected %s, %d of %d", cp.palette[index].Name, index+1, len(cp.palette)),
				accessibility.PriorityPolite,
			)
		}
	}
}

// SetStyle sets the widget style.
func (cp *ColorPicker) SetStyle(style backend.Style) {
	if cp == nil {
		return
	}
	cp.style = style
	cp.styleSet = true
}

// StyleType returns the selector type name.
func (cp *ColorPicker) StyleType() string {
	return "ColorPicker"
}

// Bind attaches app services.
func (cp *ColorPicker) Bind(services runtime.Services) {
	if cp == nil {
		return
	}
	cp.services = services
}

// Unbind releases app services.
func (cp *ColorPicker) Unbind() {
	if cp == nil {
		return
	}
	cp.services = runtime.Services{}
}

// Measure returns the desired size.
func (cp *ColorPicker) Measure(constraints runtime.Constraints) runtime.Size {
	return cp.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		cols := colorPickerColumns
		rows := (len(cp.palette) + cols - 1) / cols
		// Each swatch is 4 chars wide: [##]
		width := cols * 4
		// rows for swatches + 1 for selected display
		height := rows + 1
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the color picker.
func (cp *ColorPicker) Render(ctx runtime.RenderContext) {
	if cp == nil {
		return
	}
	cp.syncA11y()
	bounds := cp.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, cp, cp.style, cp.styleSet)

	cols := colorPickerColumns
	row := 0
	for i, entry := range cp.palette {
		col := i % cols
		if col == 0 && i > 0 {
			row++
		}
		x := bounds.X + col*4
		y := bounds.Y + row
		if y >= bounds.Y+bounds.Height-1 {
			break
		}
		swatchStyle := baseStyle.Background(entry.Color).Foreground(backend.ColorWhite)
		label := "\u2588\u2588"
		if i == cp.selected {
			if cp.focused {
				label = "><"
			} else {
				label = "**"
			}
		}
		swatch := "[" + label + "]"
		if x+4 <= bounds.X+bounds.Width {
			writePadded(ctx.Buffer, x, y, 4, swatch, swatchStyle)
		}
	}

	// Bottom row: show selected color name
	infoY := bounds.Y + bounds.Height - 1
	if infoY > bounds.Y+row && infoY < bounds.Y+bounds.Height {
		info := ""
		if cp.selected < len(cp.palette) {
			info = cp.palette[cp.selected].Name
		}
		writePadded(ctx.Buffer, bounds.X, infoY, bounds.Width, truncateString(info, bounds.Width), baseStyle)
	}
}

// HandleMessage processes keyboard input.
func (cp *ColorPicker) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if cp == nil || !cp.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if len(cp.palette) == 0 {
		return runtime.Unhandled()
	}
	cols := colorPickerColumns
	switch key.Key {
	case terminal.KeyRight:
		if cp.selected < len(cp.palette)-1 {
			cp.SetSelected(cp.selected + 1)
			return runtime.Handled()
		}
	case terminal.KeyLeft:
		if cp.selected > 0 {
			cp.SetSelected(cp.selected - 1)
			return runtime.Handled()
		}
	case terminal.KeyDown:
		next := cp.selected + cols
		if next < len(cp.palette) {
			cp.SetSelected(next)
			return runtime.Handled()
		}
	case terminal.KeyUp:
		prev := cp.selected - cols
		if prev >= 0 {
			cp.SetSelected(prev)
			return runtime.Handled()
		}
	case terminal.KeyEnter:
		if cp.onChange != nil {
			cp.onChange(cp.palette[cp.selected].Color)
		}
		if announcer := cp.services.Announcer(); announcer != nil {
			announcer.Announce(
				fmt.Sprintf("Confirmed %s", cp.palette[cp.selected].Name),
				accessibility.PriorityPolite,
			)
		}
		return runtime.Handled()
	case terminal.KeyHome:
		cp.SetSelected(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		cp.SetSelected(len(cp.palette) - 1)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (cp *ColorPicker) syncA11y() {
	if cp == nil {
		return
	}
	if cp.Base.Role == "" {
		cp.Base.Role = accessibility.RoleListbox
	}
	cp.Base.Label = "Color Picker"
	cp.Base.RoleDescription = "color picker"
	cp.Base.Orientation = "horizontal"
	if len(cp.palette) > 0 && cp.selected < len(cp.palette) {
		cp.Base.Description = cp.palette[cp.selected].Name
		cp.Base.PosInSet = cp.selected + 1
		cp.Base.SetSize = len(cp.palette)
	}
}

func defaultColorPalette() []colorEntry {
	return []colorEntry{
		{backend.ColorBlack, "Black"},
		{backend.ColorRed, "Red"},
		{backend.ColorGreen, "Green"},
		{backend.ColorYellow, "Yellow"},
		{backend.ColorBlue, "Blue"},
		{backend.ColorMagenta, "Magenta"},
		{backend.ColorCyan, "Cyan"},
		{backend.ColorWhite, "White"},
		{backend.ColorBrightBlack, "Bright Black"},
		{backend.ColorBrightRed, "Bright Red"},
		{backend.ColorBrightGreen, "Bright Green"},
		{backend.ColorBrightYellow, "Bright Yellow"},
		{backend.ColorBrightBlue, "Bright Blue"},
		{backend.ColorBrightMagenta, "Bright Magenta"},
		{backend.ColorBrightCyan, "Bright Cyan"},
		{backend.ColorBrightWhite, "Bright White"},
		{backend.ColorRGB(255, 128, 0), "Orange"},
		{backend.ColorRGB(128, 0, 128), "Purple"},
		{backend.ColorRGB(255, 192, 203), "Pink"},
		{backend.ColorRGB(165, 42, 42), "Brown"},
		{backend.ColorRGB(128, 128, 128), "Gray"},
		{backend.ColorRGB(0, 128, 128), "Teal"},
		{backend.ColorRGB(75, 0, 130), "Indigo"},
		{backend.ColorRGB(255, 215, 0), "Gold"},
	}
}

// Ensure interface compliance to catch build errors early.
// Use _ to suppress "unused" errors.
var _ runtime.Widget = (*ColorPicker)(nil)
var _ runtime.Focusable = (*ColorPicker)(nil)
var _ runtime.Bindable = (*ColorPicker)(nil)
var _ runtime.Unbindable = (*ColorPicker)(nil)

