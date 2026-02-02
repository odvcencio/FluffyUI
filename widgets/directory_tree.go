package widgets

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/scroll"
	"github.com/odvcencio/fluffyui/terminal"
)

// DirectoryEntry represents a file or directory in the tree.
type DirectoryEntry struct {
	Path     string
	Name     string
	IsDir    bool
	Expanded bool
	Depth    int
	Parent   *DirectoryEntry
	Children []*DirectoryEntry
	Loaded   bool // true if children have been loaded (for lazy loading)
}

// DirectoryTree is a lazy-loading file browser with virtual scrolling.
type DirectoryTree struct {
	FocusableBase

	root       string
	showHidden bool
	filter     func(os.DirEntry) bool
	onSelect   func(path string)
	lazyLoad   bool

	rootEntry     *DirectoryEntry
	flatCache     []*DirectoryEntry
	flatDirty     bool
	selectedIndex int
	offset        int

	label         string
	style         backend.Style
	selectedStyle backend.Style
	dirIcon       string
	fileIcon      string
	expandedIcon  string
	collapsedIcon string
	services      runtime.Services
}

// DirectoryTreeOption is a functional option for DirectoryTree.
type DirectoryTreeOption func(*DirectoryTree)

// WithShowHidden enables showing hidden files (starting with .).
func WithShowHidden(show bool) DirectoryTreeOption {
	return func(d *DirectoryTree) {
		d.showHidden = show
	}
}

// WithDirectoryFilter sets a filter function for entries.
func WithDirectoryFilter(filter func(os.DirEntry) bool) DirectoryTreeOption {
	return func(d *DirectoryTree) {
		d.filter = filter
	}
}

// WithOnSelect sets the selection callback.
func WithOnSelect(fn func(path string)) DirectoryTreeOption {
	return func(d *DirectoryTree) {
		d.onSelect = fn
	}
}

// WithLazyLoad enables lazy loading of children.
func WithLazyLoad(lazy bool) DirectoryTreeOption {
	return func(d *DirectoryTree) {
		d.lazyLoad = lazy
	}
}

// WithDirectoryStyle sets the base style.
func WithDirectoryStyle(s backend.Style) DirectoryTreeOption {
	return func(d *DirectoryTree) {
		d.style = s
	}
}

// WithDirectorySelectedStyle sets the selected row style.
func WithDirectorySelectedStyle(s backend.Style) DirectoryTreeOption {
	return func(d *DirectoryTree) {
		d.selectedStyle = s
	}
}

// WithDirectoryIcons sets custom icons.
func WithDirectoryIcons(dir, file, expanded, collapsed string) DirectoryTreeOption {
	return func(d *DirectoryTree) {
		d.dirIcon = dir
		d.fileIcon = file
		d.expandedIcon = expanded
		d.collapsedIcon = collapsed
	}
}

// NewDirectoryTree creates a new directory tree widget.
func NewDirectoryTree(root string, opts ...DirectoryTreeOption) *DirectoryTree {
	d := &DirectoryTree{
		root:          root,
		lazyLoad:      true,
		label:         "Directory Tree",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		dirIcon:       "[D]",
		fileIcon:      "   ",
		expandedIcon:  "[-]",
		collapsedIcon: "[+]",
		flatDirty:     true,
	}
	for _, opt := range opts {
		opt(d)
	}
	d.Base.Role = accessibility.RoleTree
	d.initRoot()
	d.syncA11y()
	return d
}

// Bind attaches app services.
func (d *DirectoryTree) Bind(services runtime.Services) {
	if d == nil {
		return
	}
	d.services = services
}

// Unbind releases app services.
func (d *DirectoryTree) Unbind() {
	if d == nil {
		return
	}
	d.services = runtime.Services{}
}

// SetRoot changes the root directory.
func (d *DirectoryTree) SetRoot(root string) {
	if d == nil {
		return
	}
	d.root = root
	d.initRoot()
	d.flatDirty = true
	d.selectedIndex = 0
	d.offset = 0
	d.syncA11y()
}

// Root returns the current root directory.
func (d *DirectoryTree) Root() string {
	if d == nil {
		return ""
	}
	return d.root
}

// SetShowHidden enables or disables showing hidden files.
func (d *DirectoryTree) SetShowHidden(show bool) {
	if d == nil {
		return
	}
	if d.showHidden != show {
		d.showHidden = show
		d.Refresh()
	}
}

// ShowHidden returns whether hidden files are shown.
func (d *DirectoryTree) ShowHidden() bool {
	if d == nil {
		return false
	}
	return d.showHidden
}

// SetFilter sets the filter function.
func (d *DirectoryTree) SetFilter(filter func(os.DirEntry) bool) {
	if d == nil {
		return
	}
	d.filter = filter
	d.Refresh()
}

// SetOnSelect sets the selection callback.
func (d *DirectoryTree) SetOnSelect(fn func(path string)) {
	if d == nil {
		return
	}
	d.onSelect = fn
}

// SetLazyLoad enables or disables lazy loading.
func (d *DirectoryTree) SetLazyLoad(lazy bool) {
	if d == nil {
		return
	}
	d.lazyLoad = lazy
}

// LazyLoad returns whether lazy loading is enabled.
func (d *DirectoryTree) LazyLoad() bool {
	if d == nil {
		return false
	}
	return d.lazyLoad
}

// SetStyle sets the base style.
func (d *DirectoryTree) SetStyle(style backend.Style) {
	if d == nil {
		return
	}
	d.style = style
}

// SetSelectedStyle sets the selected row style.
func (d *DirectoryTree) SetSelectedStyle(style backend.Style) {
	if d == nil {
		return
	}
	d.selectedStyle = style
}

// SetLabel sets the accessibility label.
func (d *DirectoryTree) SetLabel(label string) {
	if d == nil {
		return
	}
	d.label = label
	d.syncA11y()
}

// StyleType returns the selector type name.
func (d *DirectoryTree) StyleType() string {
	return "DirectoryTree"
}

// SelectedPath returns the currently selected path.
func (d *DirectoryTree) SelectedPath() string {
	if d == nil {
		return ""
	}
	entries := d.flatten()
	if d.selectedIndex >= 0 && d.selectedIndex < len(entries) {
		return entries[d.selectedIndex].Path
	}
	return ""
}

// SelectedEntry returns the currently selected entry.
func (d *DirectoryTree) SelectedEntry() *DirectoryEntry {
	if d == nil {
		return nil
	}
	entries := d.flatten()
	if d.selectedIndex >= 0 && d.selectedIndex < len(entries) {
		return entries[d.selectedIndex]
	}
	return nil
}

// Refresh reloads the directory tree.
func (d *DirectoryTree) Refresh() {
	if d == nil {
		return
	}
	d.initRoot()
	d.flatDirty = true
	d.syncA11y()
	d.Invalidate()
}

// ExpandSelected expands the selected directory.
func (d *DirectoryTree) ExpandSelected() {
	if d == nil {
		return
	}
	entry := d.SelectedEntry()
	if entry != nil && entry.IsDir && !entry.Expanded {
		d.expandEntry(entry)
		d.flatDirty = true
		d.syncA11y()
		d.Invalidate()
	}
}

// CollapseSelected collapses the selected directory.
func (d *DirectoryTree) CollapseSelected() {
	if d == nil {
		return
	}
	entry := d.SelectedEntry()
	if entry != nil && entry.IsDir && entry.Expanded {
		entry.Expanded = false
		d.flatDirty = true
		d.syncA11y()
		d.Invalidate()
	}
}

// ToggleSelected toggles expand/collapse on the selected directory.
func (d *DirectoryTree) ToggleSelected() {
	if d == nil {
		return
	}
	entry := d.SelectedEntry()
	if entry != nil && entry.IsDir {
		if entry.Expanded {
			entry.Expanded = false
		} else {
			d.expandEntry(entry)
		}
		d.flatDirty = true
		d.syncA11y()
		d.Invalidate()
	}
}

// Measure returns the desired size.
func (d *DirectoryTree) Measure(constraints runtime.Constraints) runtime.Size {
	return d.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		count := len(d.flatten())
		height := min(count, contentConstraints.MaxHeight)
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: height})
	})
}

// Layout stores the assigned bounds.
func (d *DirectoryTree) Layout(bounds runtime.Rect) {
	if d == nil {
		return
	}
	d.FocusableBase.Layout(bounds)
}

// Render draws the directory tree.
func (d *DirectoryTree) Render(ctx runtime.RenderContext) {
	if d == nil {
		return
	}
	d.syncA11y()
	outer := d.bounds
	content := d.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}

	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, d, backend.DefaultStyle(), false), d.style)
	ctx.Buffer.Fill(outer, ' ', baseStyle)

	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	entries := d.flatten()
	if len(entries) == 0 {
		return
	}

	// Clamp selection
	if d.selectedIndex < 0 {
		d.selectedIndex = 0
	}
	if d.selectedIndex >= len(entries) {
		d.selectedIndex = len(entries) - 1
	}

	// Adjust scroll offset
	if d.selectedIndex < d.offset {
		d.offset = d.selectedIndex
	}
	if d.selectedIndex >= d.offset+content.Height {
		d.offset = d.selectedIndex - content.Height + 1
	}

	for i := 0; i < content.Height; i++ {
		rowIndex := d.offset + i
		if rowIndex < 0 || rowIndex >= len(entries) {
			break
		}

		entry := entries[rowIndex]
		style := baseStyle
		if rowIndex == d.selectedIndex {
			style = mergeBackendStyles(baseStyle, d.selectedStyle)
		}

		// Build the line
		indent := strings.Repeat("  ", entry.Depth)
		var icon string
		if entry.IsDir {
			if entry.Expanded {
				icon = d.expandedIcon
			} else {
				icon = d.collapsedIcon
			}
		} else {
			icon = d.fileIcon
		}

		line := indent + icon + " " + entry.Name
		line = truncateString(line, content.Width)
		writePadded(ctx.Buffer, content.X, content.Y+i, content.Width, line, style)
	}
}

// HandleMessage handles navigation and expansion input.
func (d *DirectoryTree) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d == nil || !d.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	entries := d.flatten()
	count := len(entries)

	switch key.Key {
	case terminal.KeyUp:
		d.setSelected(d.selectedIndex-1, count)
		return runtime.Handled()

	case terminal.KeyDown:
		d.setSelected(d.selectedIndex+1, count)
		return runtime.Handled()

	case terminal.KeyPageUp:
		pageSize := d.ContentBounds().Height
		if pageSize < 1 {
			pageSize = 1
		}
		d.setSelected(d.selectedIndex-pageSize, count)
		return runtime.Handled()

	case terminal.KeyPageDown:
		pageSize := d.ContentBounds().Height
		if pageSize < 1 {
			pageSize = 1
		}
		d.setSelected(d.selectedIndex+pageSize, count)
		return runtime.Handled()

	case terminal.KeyHome:
		d.setSelected(0, count)
		return runtime.Handled()

	case terminal.KeyEnd:
		d.setSelected(count-1, count)
		return runtime.Handled()

	case terminal.KeyLeft:
		d.CollapseSelected()
		return runtime.Handled()

	case terminal.KeyRight:
		d.ExpandSelected()
		return runtime.Handled()

	case terminal.KeyEnter:
		entry := d.SelectedEntry()
		if entry != nil {
			if entry.IsDir {
				d.ToggleSelected()
			}
			if d.onSelect != nil {
				d.onSelect(entry.Path)
			}
		}
		return runtime.Handled()

	case terminal.KeyTab:
		if key.Shift {
			return runtime.WithCommand(runtime.FocusPrev{})
		}
		return runtime.WithCommand(runtime.FocusNext{})

	case terminal.KeyEscape:
		return runtime.WithCommand(runtime.Cancel{})
	}

	return runtime.Unhandled()
}

// ScrollBy scrolls selection by delta rows.
func (d *DirectoryTree) ScrollBy(dx, dy int) {
	if d == nil || dy == 0 {
		return
	}
	entries := d.flatten()
	d.setSelected(d.selectedIndex+dy, len(entries))
	d.Invalidate()
}

// ScrollTo scrolls to an absolute row index.
func (d *DirectoryTree) ScrollTo(x, y int) {
	if d == nil {
		return
	}
	entries := d.flatten()
	d.setSelected(y, len(entries))
	d.Invalidate()
}

// PageBy scrolls by a number of pages.
func (d *DirectoryTree) PageBy(pages int) {
	if d == nil {
		return
	}
	entries := d.flatten()
	pageSize := d.bounds.Height
	if pageSize < 1 {
		pageSize = 1
	}
	d.setSelected(d.selectedIndex+pages*pageSize, len(entries))
	d.Invalidate()
}

// ScrollToStart scrolls to the first row.
func (d *DirectoryTree) ScrollToStart() {
	if d == nil {
		return
	}
	entries := d.flatten()
	d.setSelected(0, len(entries))
	d.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (d *DirectoryTree) ScrollToEnd() {
	if d == nil {
		return
	}
	entries := d.flatten()
	d.setSelected(len(entries)-1, len(entries))
	d.Invalidate()
}

func (d *DirectoryTree) initRoot() {
	if d == nil {
		return
	}

	info, err := os.Stat(d.root)
	if err != nil {
		d.rootEntry = nil
		return
	}

	d.rootEntry = &DirectoryEntry{
		Path:     d.root,
		Name:     filepath.Base(d.root),
		IsDir:    info.IsDir(),
		Expanded: true,
		Depth:    0,
		Loaded:   false,
	}

	if d.rootEntry.IsDir {
		d.loadChildren(d.rootEntry)
	}
}

func (d *DirectoryTree) loadChildren(entry *DirectoryEntry) {
	if entry == nil || !entry.IsDir || entry.Loaded {
		return
	}

	entries, err := os.ReadDir(entry.Path)
	if err != nil {
		entry.Loaded = true
		return
	}

	// Sort: directories first, then alphabetically
	sort.Slice(entries, func(i, j int) bool {
		iDir := entries[i].IsDir()
		jDir := entries[j].IsDir()
		if iDir != jDir {
			return iDir
		}
		return entries[i].Name() < entries[j].Name()
	})

	entry.Children = nil
	for _, e := range entries {
		name := e.Name()

		// Skip hidden files if not showing them
		if !d.showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		// Apply filter
		if d.filter != nil && !d.filter(e) {
			continue
		}

		child := &DirectoryEntry{
			Path:   filepath.Join(entry.Path, name),
			Name:   name,
			IsDir:  e.IsDir(),
			Depth:  entry.Depth + 1,
			Parent: entry,
			Loaded: !d.lazyLoad || !e.IsDir(), // Files are always "loaded"
		}
		entry.Children = append(entry.Children, child)
	}

	entry.Loaded = true
}

func (d *DirectoryTree) expandEntry(entry *DirectoryEntry) {
	if entry == nil || !entry.IsDir {
		return
	}

	if !entry.Loaded {
		d.loadChildren(entry)
	}
	entry.Expanded = true
}

func (d *DirectoryTree) flatten() []*DirectoryEntry {
	if d == nil || d.rootEntry == nil {
		return nil
	}

	if !d.flatDirty && d.flatCache != nil {
		return d.flatCache
	}

	d.flatCache = d.flatCache[:0]
	d.flattenRecursive(d.rootEntry)
	d.flatDirty = false

	return d.flatCache
}

func (d *DirectoryTree) flattenRecursive(entry *DirectoryEntry) {
	if entry == nil {
		return
	}

	d.flatCache = append(d.flatCache, entry)

	if entry.IsDir && entry.Expanded {
		for _, child := range entry.Children {
			d.flattenRecursive(child)
		}
	}
}

func (d *DirectoryTree) setSelected(index int, count int) {
	if count == 0 {
		d.selectedIndex = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	d.selectedIndex = index
	d.syncA11y()
}

func (d *DirectoryTree) syncA11y() {
	if d == nil {
		return
	}
	if d.Base.Role == "" {
		d.Base.Role = accessibility.RoleTree
	}
	label := strings.TrimSpace(d.label)
	if label == "" {
		label = "Directory Tree"
	}
	d.Base.Label = label

	entries := d.flatten()
	d.Base.Description = fmt.Sprintf("%d items", len(entries))

	if entry := d.SelectedEntry(); entry != nil {
		d.Base.Value = &accessibility.ValueInfo{Text: entry.Path}
		if entry.IsDir {
			d.Base.State.Expanded = accessibility.BoolPtr(entry.Expanded)
		} else {
			d.Base.State.Expanded = nil
		}
	} else {
		d.Base.Value = nil
		d.Base.State.Expanded = nil
	}
}

var _ scroll.Controller = (*DirectoryTree)(nil)
var _ runtime.Widget = (*DirectoryTree)(nil)
var _ runtime.Focusable = (*DirectoryTree)(nil)
