package widgets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// FilePickerOption configures a FilePicker widget.
type FilePickerOption = Option[FilePicker]

// FilePicker is a focusable file/directory browser for selecting files.
type FilePicker struct {
	FocusableBase

	currentDir string
	entries    []os.DirEntry
	selected   int
	filter     string // file extension filter, e.g. "*.go"
	onSelect   func(path string)
	showHidden bool
	services   runtime.Services

	style    backend.Style
	styleSet bool
}

// NewFilePicker creates a file picker for the given directory.
func NewFilePicker(dir string, opts ...FilePickerOption) *FilePicker {
	if dir == "" {
		dir, _ = os.Getwd()
		if dir == "" {
			dir = "/"
		}
	}
	fp := &FilePicker{
		currentDir: dir,
		style:      backend.DefaultStyle(),
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(fp)
	}
	fp.Base.Role = accessibility.RoleTree
	fp.loadEntries()
	fp.syncA11y()
	return fp
}

// WithFileFilter sets the file extension filter (e.g. "*.go").
func WithFileFilter(pattern string) FilePickerOption {
	return func(fp *FilePicker) {
		if fp == nil {
			return
		}
		fp.filter = pattern
	}
}

// WithFilePickerShowHidden controls whether hidden files are shown.
func WithFilePickerShowHidden(show bool) FilePickerOption {
	return func(fp *FilePicker) {
		if fp == nil {
			return
		}
		fp.showHidden = show
	}
}

// WithOnFileSelect sets the file selection callback.
func WithOnFileSelect(fn func(string)) FilePickerOption {
	return func(fp *FilePicker) {
		if fp == nil {
			return
		}
		fp.onSelect = fn
	}
}

// WithStartDir sets the starting directory.
func WithStartDir(dir string) FilePickerOption {
	return func(fp *FilePicker) {
		if fp == nil {
			return
		}
		fp.currentDir = dir
	}
}

// CurrentDir returns the current directory path.
func (fp *FilePicker) CurrentDir() string {
	if fp == nil {
		return ""
	}
	return fp.currentDir
}

// SelectedIndex returns the currently selected entry index.
func (fp *FilePicker) SelectedIndex() int {
	if fp == nil {
		return 0
	}
	return fp.selected
}

// EntryCount returns the number of visible entries (including "..").
func (fp *FilePicker) EntryCount() int {
	if fp == nil {
		return 0
	}
	// +1 for the ".." entry
	return len(fp.entries) + 1
}

// SetStyle sets the widget style.
func (fp *FilePicker) SetStyle(style backend.Style) {
	if fp == nil {
		return
	}
	fp.style = style
	fp.styleSet = true
}

// StyleType returns the selector type name.
func (fp *FilePicker) StyleType() string {
	return "FilePicker"
}

// Bind attaches app services.
func (fp *FilePicker) Bind(services runtime.Services) {
	if fp == nil {
		return
	}
	fp.services = services
}

// Unbind releases app services.
func (fp *FilePicker) Unbind() {
	if fp == nil {
		return
	}
	fp.services = runtime.Services{}
}

// Measure returns the desired size.
func (fp *FilePicker) Measure(constraints runtime.Constraints) runtime.Size {
	return fp.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := 40
		if contentConstraints.MaxWidth < width {
			width = contentConstraints.MaxWidth
		}
		// header line + separator + entries (max 15) + 1 for ".."
		visibleEntries := len(fp.entries) + 1
		if visibleEntries > 15 {
			visibleEntries = 15
		}
		height := 2 + visibleEntries // path header + separator + entries
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the file picker.
func (fp *FilePicker) Render(ctx runtime.RenderContext) {
	if fp == nil {
		return
	}
	fp.syncA11y()
	bounds := fp.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, fp, fp.style, fp.styleSet)
	dirStyle := baseStyle.Foreground(backend.ColorCyan).Bold(true)
	selectedStyle := baseStyle.Reverse(true)
	dimStyle := baseStyle.Dim(true)

	row := 0

	// Directory path header
	if row < bounds.Height {
		dirPath := truncateString(fp.currentDir, bounds.Width)
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, dirPath, dirStyle)
		row++
	}

	// Separator
	if row < bounds.Height {
		sep := strings.Repeat("\u2504", bounds.Width)
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(sep, bounds.Width), dimStyle)
		row++
	}

	// ".." parent directory entry
	totalEntries := len(fp.entries) + 1
	entryIdx := 0

	// Scroll offset for large directories
	scrollOffset := 0
	maxVisible := bounds.Height - row
	if maxVisible < 1 {
		maxVisible = 1
	}
	if fp.selected >= scrollOffset+maxVisible {
		scrollOffset = fp.selected - maxVisible + 1
	}
	if scrollOffset > fp.selected {
		scrollOffset = fp.selected
	}

	for i := scrollOffset; i < totalEntries && row < bounds.Height; i++ {
		lineStyle := baseStyle
		indicator := "  "
		if i == fp.selected {
			lineStyle = selectedStyle
			indicator = "> "
		}

		var line string
		if i == 0 {
			// Parent directory
			line = indicator + "\U0001F4C1 .."
		} else {
			entry := fp.entries[i-1]
			icon := "\U0001F4C4" // file icon
			suffix := ""
			if entry.IsDir() {
				icon = "\U0001F4C1"
				suffix = "/"
			}
			name := entry.Name() + suffix
			sizeStr := ""
			if !entry.IsDir() {
				info, err := entry.Info()
				if err == nil {
					sizeStr = formatFileSize(info.Size())
				}
			}
			nameWidth := bounds.Width - textWidth(indicator) - textWidth(icon) - 1 - textWidth(sizeStr)
			if sizeStr != "" {
				nameWidth-- // space before size
			}
			if nameWidth < 1 {
				nameWidth = 1
			}
			line = indicator + icon + " " + padRight(truncateString(name, nameWidth), nameWidth)
			if sizeStr != "" {
				line += " " + sizeStr
			}
		}

		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, truncateString(line, bounds.Width), lineStyle)
		row++
		entryIdx++
	}

	// Fill remaining space
	for row < bounds.Height {
		writePadded(ctx.Buffer, bounds.X, bounds.Y+row, bounds.Width, "", baseStyle)
		row++
	}

	_ = entryIdx
}

// HandleMessage processes keyboard input.
func (fp *FilePicker) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if fp == nil || !fp.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	totalEntries := len(fp.entries) + 1

	switch key.Key {
	case terminal.KeyUp:
		if fp.selected > 0 {
			fp.selected--
			fp.announceSelected()
			fp.syncA11y()
		}
		return runtime.Handled()

	case terminal.KeyDown:
		if fp.selected < totalEntries-1 {
			fp.selected++
			fp.announceSelected()
			fp.syncA11y()
		}
		return runtime.Handled()

	case terminal.KeyEnter:
		if fp.selected == 0 {
			// Go to parent
			fp.navigateUp()
		} else {
			entry := fp.entries[fp.selected-1]
			if entry.IsDir() {
				fp.navigateInto(entry.Name())
			} else {
				path := filepath.Join(fp.currentDir, entry.Name())
				if fp.onSelect != nil {
					fp.onSelect(path)
				}
				if announcer := fp.services.Announcer(); announcer != nil {
					announcer.Announce(fmt.Sprintf("Selected %s", entry.Name()), accessibility.PriorityPolite)
				}
			}
		}
		return runtime.Handled()

	case terminal.KeyBackspace, terminal.KeyLeft:
		fp.navigateUp()
		return runtime.Handled()

	case terminal.KeyHome:
		fp.selected = 0
		fp.announceSelected()
		fp.syncA11y()
		return runtime.Handled()

	case terminal.KeyEnd:
		fp.selected = totalEntries - 1
		if fp.selected < 0 {
			fp.selected = 0
		}
		fp.announceSelected()
		fp.syncA11y()
		return runtime.Handled()
	}

	return runtime.Unhandled()
}

func (fp *FilePicker) navigateUp() {
	parent := filepath.Dir(fp.currentDir)
	if parent == fp.currentDir {
		return // already at root
	}
	fp.currentDir = parent
	fp.selected = 0
	fp.loadEntries()
	fp.syncA11y()
	if announcer := fp.services.Announcer(); announcer != nil {
		announcer.Announce(fmt.Sprintf("Navigated to %s", fp.currentDir), accessibility.PriorityPolite)
	}
}

func (fp *FilePicker) navigateInto(name string) {
	newDir := filepath.Join(fp.currentDir, name)
	fp.currentDir = newDir
	fp.selected = 0
	fp.loadEntries()
	fp.syncA11y()
	if announcer := fp.services.Announcer(); announcer != nil {
		announcer.Announce(fmt.Sprintf("Navigated to %s", fp.currentDir), accessibility.PriorityPolite)
	}
}

func (fp *FilePicker) loadEntries() {
	entries, err := os.ReadDir(fp.currentDir)
	if err != nil {
		fp.entries = nil
		return
	}

	filtered := entries[:0]
	for _, e := range entries {
		// Filter hidden files
		if !fp.showHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		// Apply file filter (directories always pass)
		if fp.filter != "" && !e.IsDir() {
			matched, _ := filepath.Match(fp.filter, e.Name())
			if !matched {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	fp.entries = filtered
}

func (fp *FilePicker) announceSelected() {
	if fp == nil {
		return
	}
	announcer := fp.services.Announcer()
	if announcer == nil {
		return
	}
	totalEntries := len(fp.entries) + 1
	if fp.selected == 0 {
		announcer.Announce(
			fmt.Sprintf("Parent directory, 1 of %d", totalEntries),
			accessibility.PriorityPolite,
		)
	} else if fp.selected-1 < len(fp.entries) {
		entry := fp.entries[fp.selected-1]
		kind := "file"
		if entry.IsDir() {
			kind = "directory"
		}
		announcer.Announce(
			fmt.Sprintf("%s %s, %d of %d", entry.Name(), kind, fp.selected+1, totalEntries),
			accessibility.PriorityPolite,
		)
	}
}

func (fp *FilePicker) syncA11y() {
	if fp == nil {
		return
	}
	if fp.Base.Role == "" {
		fp.Base.Role = accessibility.RoleTree
	}
	fp.Base.Label = "File Picker"
	fp.Base.Description = fp.currentDir
	fp.Base.Orientation = "vertical"
	totalEntries := len(fp.entries) + 1
	if fp.selected < totalEntries {
		fp.Base.PosInSet = fp.selected + 1
		fp.Base.SetSize = totalEntries
	}
}

func formatFileSize(size int64) string {
	switch {
	case size >= 1024*1024*1024:
		return fmt.Sprintf("%.1fG", float64(size)/(1024*1024*1024))
	case size >= 1024*1024:
		return fmt.Sprintf("%.1fM", float64(size)/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%.1fK", float64(size)/1024)
	default:
		return fmt.Sprintf("%dB", size)
	}
}

var _ runtime.Widget = (*FilePicker)(nil)
var _ runtime.Focusable = (*FilePicker)(nil)
var _ runtime.Bindable = (*FilePicker)(nil)
var _ runtime.Unbindable = (*FilePicker)(nil)
