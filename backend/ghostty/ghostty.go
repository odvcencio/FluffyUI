//go:build linux || darwin || windows

package ghostty

import (
	"errors"
	"sync"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/terminal"
)

const (
	defaultEventBuffer = 128
	pollInterval       = 8 * time.Millisecond
)

var (
	ErrBackendClosed  = errors.New("ghostty backend closed")
	ErrEventQueueFull = errors.New("ghostty event queue full")
)

// Backend implements backend.Backend using libghostty via purego.
type Backend struct {
	lib     *ghosttyLib
	config  *ghosttyConfig
	app     *ghosttyApp
	surface *ghosttySurface

	width  int
	height int
	cells  []backend.Cell
	mouseX int
	mouseY int

	mousePosValid bool

	events   chan terminal.Event
	closed   chan struct{}
	once     sync.Once
	pollOnce sync.Once
	pollWG   sync.WaitGroup

	mu          sync.Mutex
	forceRedraw bool
}

// New creates a new ghostty backend.
func New() (*Backend, error) {
	lib, err := loadGhosttyLib()
	if err != nil {
		return nil, err
	}
	return &Backend{
		lib:    lib,
		events: make(chan terminal.Event, defaultEventBuffer),
		closed: make(chan struct{}),
	}, nil
}

// Init initializes the backend.
func (b *Backend) Init() error {
	if b == nil {
		return errors.New("ghostty backend is nil")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ensureSurfaceLocked()
}

// Fini cleans up the backend.
func (b *Backend) Fini() {
	if b == nil {
		return
	}
	b.once.Do(func() {
		close(b.closed)
	})
	b.pollWG.Wait()
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface != nil && b.lib != nil && b.lib.surfaceFree != nil {
		b.lib.surfaceFree(b.surface)
	}
	b.surface = nil
	if b.app != nil && b.lib != nil && b.lib.appFree != nil {
		b.lib.appFree(b.app)
	}
	b.app = nil
	if b.config != nil && b.lib != nil && b.lib.configFree != nil {
		b.lib.configFree(b.config)
	}
	b.config = nil
}

// Size returns the terminal dimensions.
func (b *Backend) Size() (width, height int) {
	if b == nil {
		return 0, 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil {
		return b.width, b.height
	}
	b.refreshSizeLocked()
	return b.width, b.height
}

// SetContent sets a cell at position (x, y).
func (b *Backend) SetContent(x, y int, mainc rune, comb []rune, style backend.Style) {
	if b == nil {
		return
	}
	if len(comb) > 0 {
		// Combining characters are ignored until libghostty exposes multi-codepoint cells.
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil {
		return
	}
	if x < 0 || y < 0 || x >= b.width || y >= b.height {
		return
	}
	b.setCellLocked(x, y, mainc, style)
}

// SetRow updates a row using a slice of cells.
func (b *Backend) SetRow(y int, startX int, cells []backend.Cell) {
	if b == nil || startX < 0 || len(cells) == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil || y < 0 || y >= b.height || startX >= b.width {
		return
	}
	max := b.width - startX
	if len(cells) > max {
		cells = cells[:max]
	}
	rowStart := y*b.width + startX
	for i, cell := range cells {
		idx := rowStart + i
		if idx >= 0 && idx < len(b.cells) {
			b.cells[idx] = cell
		}
		b.emitCellLocked(startX+i, y, cell)
	}
}

// SetRect updates a rectangle using row-major cells.
func (b *Backend) SetRect(x, y, width, height int, cells []backend.Cell) {
	if b == nil || width <= 0 || height <= 0 {
		return
	}
	expected := width * height
	if len(cells) < expected {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil {
		return
	}
	for row := 0; row < height; row++ {
		rowY := y + row
		if rowY < 0 || rowY >= b.height {
			continue
		}
		rowStart := row * width
		for col := 0; col < width; col++ {
			colX := x + col
			if colX < 0 || colX >= b.width {
				continue
			}
			cell := cells[rowStart+col]
			idx := rowY*b.width + colX
			if idx >= 0 && idx < len(b.cells) {
				b.cells[idx] = cell
			}
			b.emitCellLocked(colX, rowY, cell)
		}
	}
}

// Show synchronizes the buffer to the terminal.
func (b *Backend) Show() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil {
		return
	}
	if b.forceRedraw {
		b.redrawLocked()
		b.forceRedraw = false
	}
	b.lib.surfaceShow(b.surface)
}

// Clear clears the screen.
func (b *Backend) Clear() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface != nil {
		b.lib.surfaceClear(b.surface)
	}
	b.resetCellsLocked()
}

// HideCursor hides the terminal cursor.
func (b *Backend) HideCursor() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil || b.lib == nil || b.lib.surfaceHideCursor == nil {
		return
	}
	b.lib.surfaceHideCursor(b.surface)
}

// ShowCursor shows the terminal cursor.
func (b *Backend) ShowCursor() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil || b.lib == nil || b.lib.surfaceShowCursor == nil {
		return
	}
	b.lib.surfaceShowCursor(b.surface)
}

// SetCursorPos sets the cursor position.
func (b *Backend) SetCursorPos(x, y int) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil || b.lib == nil || b.lib.surfaceSetCursorPos == nil {
		return
	}
	b.lib.surfaceSetCursorPos(b.surface, int32(x), int32(y))
}

// PollEvent blocks until an event is available.
func (b *Backend) PollEvent() terminal.Event {
	if b == nil {
		return nil
	}
	select {
	case <-b.closed:
		return nil
	case ev := <-b.events:
		return ev
	}
}

// PostEvent injects an event into the queue.
func (b *Backend) PostEvent(ev terminal.Event) error {
	if b == nil || ev == nil {
		return nil
	}
	select {
	case <-b.closed:
		return ErrBackendClosed
	case b.events <- ev:
		return nil
	default:
		return ErrEventQueueFull
	}
}

// Beep emits an audible bell.
func (b *Backend) Beep() {}

// Sync forces a full redraw.
func (b *Backend) Sync() {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.forceRedraw = true
	b.mu.Unlock()
}

func (b *Backend) startPollerLocked() {
	if b == nil || b.lib == nil || b.surface == nil {
		return
	}
	if b.lib.surfacePoll == nil && b.lib.appTick == nil {
		return
	}
	b.pollOnce.Do(func() {
		b.pollWG.Add(1)
		go b.pollLoop()
	})
}

func (b *Backend) pollLoop() {
	defer b.pollWG.Done()
	for {
		select {
		case <-b.closed:
			return
		default:
		}

		if b.lib != nil && b.lib.surfacePoll != nil {
			if b.pollSurfaceEvents() {
				continue
			}
		} else {
			b.tickApp()
		}
		time.Sleep(pollInterval)
	}
}

func (b *Backend) pollSurfaceEvents() bool {
	if b == nil || b.lib == nil || b.lib.surfacePoll == nil {
		return false
	}
	handled := false
	for {
		var ev ghosttyEvent
		if !b.pollSurfaceEvent(&ev) {
			break
		}
		handled = true
		b.handleSurfaceEvent(ev)
	}
	return handled
}

func (b *Backend) pollSurfaceEvent(out *ghosttyEvent) bool {
	if b == nil || out == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.surface == nil || b.lib == nil || b.lib.surfacePoll == nil {
		return false
	}
	return b.lib.surfacePoll(b.surface, out, 0) == 1
}

func (b *Backend) tickApp() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.app == nil || b.lib == nil || b.lib.appTick == nil {
		return
	}
	b.lib.appTick(b.app)
}

func (b *Backend) handleSurfaceEvent(ev ghosttyEvent) {
	switch ev.Tag {
	case ghosttyEventRender:
		b.mu.Lock()
		if b.surface != nil && b.lib != nil && b.lib.surfaceShow != nil {
			if b.forceRedraw {
				b.redrawLocked()
				b.forceRedraw = false
			}
			b.lib.surfaceShow(b.surface)
		}
		b.mu.Unlock()
	case ghosttyEventResize:
		columns := int(ev.Resize.Columns)
		rows := int(ev.Resize.Rows)
		changed := false
		b.mu.Lock()
		if b.surface != nil && (columns != b.width || rows != b.height) {
			b.width = columns
			b.height = rows
			b.resetCellsLocked()
			b.forceRedraw = true
			changed = true
		}
		b.mu.Unlock()
		if changed {
			_ = b.PostEvent(terminal.ResizeEvent{Width: columns, Height: rows})
		}
	case ghosttyEventKey:
		if keyEvent, ok := ghosttyKeyEventToTerminal(ev.Key); ok {
			_ = b.PostEvent(keyEvent)
		}
	case ghosttyEventMouseButton, ghosttyEventMouseMove, ghosttyEventMouseScroll:
		if mouseEvent, ok := ghosttyMouseEventToTerminal(ev.Tag, ev.Mouse); ok {
			_ = b.PostEvent(mouseEvent)
		}
	}
}

func (b *Backend) ensureSurfaceLocked() error {
	if b.surface != nil {
		return nil
	}
	if b.lib == nil {
		return errors.New("ghostty library unavailable")
	}
	if b.lib.configNew == nil || b.lib.appNewHeadless == nil || b.lib.surfaceNewHeadless == nil {
		return errors.New("ghostty library missing required symbols")
	}
	cfg := b.lib.configNew()
	if cfg == nil {
		return errors.New("ghostty_config_new returned nil")
	}
	if b.lib.configFinalize != nil {
		b.lib.configFinalize(cfg)
	}
	app := b.lib.appNewHeadless(cfg)
	if app == nil {
		if b.lib.configFree != nil {
			b.lib.configFree(cfg)
		}
		return errors.New("ghostty_app_new_headless returned nil")
	}
	surface := b.lib.surfaceNewHeadless(app)
	if surface == nil {
		if b.lib.appFree != nil {
			b.lib.appFree(app)
		}
		if b.lib.configFree != nil {
			b.lib.configFree(cfg)
		}
		return errors.New("ghostty_surface_new_headless returned nil")
	}
	b.config = cfg
	b.app = app
	b.surface = surface
	b.refreshSizeLocked()
	b.startPollerLocked()
	return nil
}

func (b *Backend) refreshSizeLocked() {
	if b.surface == nil || b.lib.surfaceGetSize == nil {
		return
	}
	var w, h int32
	b.lib.surfaceGetSize(b.surface, &w, &h)
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	if int(w) == b.width && int(h) == b.height {
		return
	}
	b.width = int(w)
	b.height = int(h)
	b.resetCellsLocked()
	b.forceRedraw = true
}

func (b *Backend) resetCellsLocked() {
	if b.width <= 0 || b.height <= 0 {
		b.cells = nil
		return
	}
	b.cells = make([]backend.Cell, b.width*b.height)
	defaultStyle := backend.DefaultStyle()
	for i := range b.cells {
		b.cells[i] = backend.Cell{Rune: ' ', Style: defaultStyle}
	}
}

func (b *Backend) setCellLocked(x, y int, mainc rune, style backend.Style) {
	if mainc == 0 {
		mainc = ' '
	}
	idx := y*b.width + x
	if idx >= 0 && idx < len(b.cells) {
		b.cells[idx] = backend.Cell{Rune: mainc, Style: style}
	}
	b.emitCellLocked(x, y, backend.Cell{Rune: mainc, Style: style})
}

func (b *Backend) emitCellLocked(x, y int, cell backend.Cell) {
	if b.surface == nil || b.lib.surfaceSetCell == nil {
		return
	}
	r := cell.Rune
	if r == 0 {
		r = ' '
	}
	fg, bg, attrs := cell.Style.Decompose()
	b.lib.surfaceSetCell(
		b.surface,
		uint32(x),
		uint32(y),
		uint32(r),
		colorToGhostty(fg),
		colorToGhostty(bg),
		attrsToGhostty(attrs),
	)
}

func (b *Backend) redrawLocked() {
	if b.surface == nil || b.lib.surfaceSetCell == nil {
		return
	}
	for y := 0; y < b.height; y++ {
		rowStart := y * b.width
		for x := 0; x < b.width; x++ {
			cell := b.cells[rowStart+x]
			b.emitCellLocked(x, y, cell)
		}
	}
}

func colorToGhostty(c backend.Color) uint32 {
	return uint32(c)
}

func attrsToGhostty(attrs backend.AttrMask) uint8 {
	return uint8(attrs)
}

// Ensure Backend implements backend.Backend.
var _ backend.Backend = (*Backend)(nil)

// Ensure Backend implements optional fast paths.
var _ backend.RowWriter = (*Backend)(nil)
var _ backend.RectWriter = (*Backend)(nil)
