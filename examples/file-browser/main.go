package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/terminal"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	view, err := NewFileBrowserView()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init failed: %v\n", err)
		os.Exit(1)
	}
	bundle, err := demo.NewApp(view, demo.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type FileEntry struct {
	Name     string
	Path     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	IsParent bool
}

type FileBrowserView struct {
	widgets.Component
	currentDir string
	entries    *state.Signal[[]FileEntry]
	selected   *FileEntry

	pathLabel   *widgets.Label
	statusLabel *widgets.Label
	list        *widgets.List[FileEntry]
	details     *widgets.Text
	leftPanel   *widgets.Panel
	rightPanel  *widgets.Panel
	splitter    *widgets.Splitter
}

func NewFileBrowserView() (*FileBrowserView, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	view := &FileBrowserView{
		currentDir: cwd,
		entries:    state.NewSignal([]FileEntry{}),
	}
	view.pathLabel = widgets.NewLabel("").WithStyle(backend.DefaultStyle().Bold(true))
	view.statusLabel = widgets.NewLabel("")
	view.details = widgets.NewText("")

	adapter := widgets.NewSignalAdapter(view.entries, func(item FileEntry, index int, selected bool, ctx runtime.RenderContext) {
		style := backend.DefaultStyle()
		if selected {
			style = style.Reverse(true)
		}
		marker := "[F]"
		name := item.Name
		if item.IsParent {
			marker = "[..]"
			name = "Parent directory"
		} else if item.IsDir {
			marker = "[D]"
			name = name + "/"
		}
		line := marker + " " + name
		line = truncateAndPad(line, ctx.Bounds.Width)
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, line, style)
	})
	view.list = widgets.NewList(adapter)
	view.list.OnSelect(func(index int, item FileEntry) {
		view.selected = &item
		view.updateDetails()
	})

	view.leftPanel = widgets.NewPanel(view.list).WithBorder(backend.DefaultStyle())
	view.leftPanel.SetTitle("Files")
	view.rightPanel = widgets.NewPanel(view.details).WithBorder(backend.DefaultStyle())
	view.rightPanel.SetTitle("Details")
	view.splitter = widgets.NewSplitter(view.leftPanel, view.rightPanel)
	view.splitter.Ratio = 0.55

	if err := view.loadDir(view.currentDir); err != nil {
		view.statusLabel.SetText(err.Error())
	}
	return view, nil
}

func (f *FileBrowserView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (f *FileBrowserView) Layout(bounds runtime.Rect) {
	f.Component.Layout(bounds)
	y := bounds.Y
	if f.pathLabel != nil {
		f.pathLabel.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	statusHeight := 1
	mainHeight := bounds.Height - (y - bounds.Y) - statusHeight
	if mainHeight < 0 {
		mainHeight = 0
	}
	if f.splitter != nil {
		f.splitter.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: mainHeight})
	}
	if f.statusLabel != nil {
		f.statusLabel.Layout(runtime.Rect{X: bounds.X, Y: y + mainHeight, Width: bounds.Width, Height: statusHeight})
	}
}

func (f *FileBrowserView) Render(ctx runtime.RenderContext) {
	if f.pathLabel != nil {
		f.pathLabel.Render(ctx)
	}
	if f.splitter != nil {
		f.splitter.Render(ctx)
	}
	if f.statusLabel != nil {
		f.statusLabel.Render(ctx)
	}
}

func (f *FileBrowserView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if key, ok := msg.(runtime.KeyMsg); ok {
		switch key.Key {
		case terminal.KeyEnter:
			f.openSelected()
			return runtime.Handled()
		case terminal.KeyBackspace, terminal.KeyLeft:
			f.goUp()
			return runtime.Handled()
		case terminal.KeyRune:
			switch key.Rune {
			case 'q', 'Q':
				return runtime.WithCommand(runtime.Quit{})
			case 'r', 'R':
				_ = f.loadDir(f.currentDir)
				return runtime.Handled()
			}
		}
	}
	if f.splitter != nil {
		return f.splitter.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (f *FileBrowserView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if f.pathLabel != nil {
		children = append(children, f.pathLabel)
	}
	if f.splitter != nil {
		children = append(children, f.splitter)
	}
	if f.statusLabel != nil {
		children = append(children, f.statusLabel)
	}
	return children
}

func (f *FileBrowserView) loadDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		f.statusLabel.SetText("Error: " + err.Error())
		return err
	}
	list := make([]FileEntry, 0, len(entries)+1)
	parent := filepath.Dir(path)
	if parent != path {
		list = append(list, FileEntry{Path: parent, IsDir: true, IsParent: true})
	}
	for _, entry := range entries {
		info, infoErr := entry.Info()
		item := FileEntry{
			Name:  entry.Name(),
			Path:  filepath.Join(path, entry.Name()),
			IsDir: entry.IsDir(),
		}
		if infoErr == nil {
			item.Size = info.Size()
			item.ModTime = info.ModTime()
		}
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].IsParent {
			return true
		}
		if list[j].IsParent {
			return false
		}
		if list[i].IsDir != list[j].IsDir {
			return list[i].IsDir
		}
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})

	f.currentDir = path
	f.entries.Set(list)
	f.pathLabel.SetText("Path: " + path)
	f.statusLabel.SetText("Enter to open, Backspace to go up, R to refresh, Q to quit")
	f.list.SetSelected(0)
	f.selected = nil
	f.updateDetails()
	f.Invalidate()
	return nil
}

func (f *FileBrowserView) openSelected() {
	if f.list == nil {
		return
	}
	item, ok := f.list.SelectedItem()
	if !ok {
		return
	}
	if item.IsDir {
		_ = f.loadDir(item.Path)
	}
}

func (f *FileBrowserView) goUp() {
	parent := filepath.Dir(f.currentDir)
	if parent == f.currentDir {
		return
	}
	_ = f.loadDir(parent)
}

func (f *FileBrowserView) updateDetails() {
	if f.details == nil {
		return
	}
	item, ok := f.list.SelectedItem()
	if !ok {
		f.details.SetText("Select a file to view details.")
		return
	}
	if item.IsParent {
		f.details.SetText("Parent directory")
		return
	}
	kind := "File"
	if item.IsDir {
		kind = "Directory"
	}
	lines := []string{
		"Name: " + item.Name,
		"Type: " + kind,
		"Size: " + formatSize(item.Size),
		"Modified: " + formatTime(item.ModTime),
		"",
		"Path:",
		item.Path,
	}
	f.details.SetText(strings.Join(lines, "\n"))
}

func formatSize(size int64) string {
	if size <= 0 {
		return "-"
	}
	units := []string{"B", "KB", "MB", "GB"}
	value := float64(size)
	idx := 0
	for value >= 1024 && idx < len(units)-1 {
		value /= 1024
		idx++
	}
	return fmt.Sprintf("%.1f %s", value, units[idx])
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

func truncateAndPad(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(text) > width {
		if width <= 3 {
			return text[:width]
		}
		text = text[:width-3] + "..."
	}
	if len(text) < width {
		pad := make([]byte, width-len(text))
		for i := range pad {
			pad[i] = ' '
		}
		text += string(pad)
	}
	return text
}
