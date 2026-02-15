//go:build !js

package web

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/terminal"
)

// Backend implements backend.Backend using WebSocket transport.
// It renders the TUI to web browsers via xterm.js.
type Backend struct {
	config Config

	// Server state
	server   *http.Server
	listener net.Listener
	upgrader websocket.Upgrader

	// Session management
	sessions     map[string]*Session
	sessionsMu   sync.RWMutex
	sessionCount atomic.Int32

	// Terminal state
	termMu   sync.RWMutex
	width    int
	height   int
	cells    []backend.Cell
	cursorX  int
	cursorY  int
	cursorOn bool

	// Input/output channels
	events   chan terminal.Event
	rendered chan []byte // ANSI data to send

	// Lifecycle
	ctx      context.Context
	cancel   context.CancelFunc
	initOnce sync.Once
	finiOnce sync.Once
}

// Session represents a connected web client.
type Session struct {
	ID       string
	Conn     *websocket.Conn
	Backend  *Backend
	LastSeen time.Time
	mu       sync.Mutex
}

// New creates a new web backend with the given configuration.
func New(config Config) *Backend {
	fillDefaults(&config)

	ctx, cancel := context.WithCancel(context.Background())

	b := &Backend{
		config:   config,
		sessions: make(map[string]*Session),
		events:   make(chan terminal.Event, 128),
		rendered: make(chan []byte, 16),
		ctx:      ctx,
		cancel:   cancel,
		width:    80,
		height:   24,
		cells:    make([]backend.Cell, 80*24),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin in development
				// Production should configure this more strictly
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	return b
}

// NewWithAddr creates a web backend listening on the specified address.
func NewWithAddr(addr string) *Backend {
	return New(Config{Addr: addr})
}

// Init initializes the backend and starts the HTTP server.
func (b *Backend) Init() error {
	var initErr error
	b.initOnce.Do(func() {
		initErr = b.init()
	})
	return initErr
}

func (b *Backend) init() error {
	// Create HTTP mux
	mux := http.NewServeMux()

	// Serve terminal UI
	mux.HandleFunc(b.config.Path, b.handleIndex)

	// Serve WebSocket endpoint
	mux.HandleFunc(b.config.WSPath, b.handleWebSocket)

	// Create listener
	listener, err := net.Listen("tcp", b.config.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", b.config.Addr, err)
	}
	b.listener = listener

	// Create HTTP server
	b.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in background
	go func() {
		if b.config.TLS.Enabled {
			_ = b.server.ServeTLS(listener, b.config.TLS.CertFile, b.config.TLS.KeyFile)
		} else {
			_ = b.server.Serve(listener)
		}
	}()

	// Wait for server to be ready
	time.Sleep(10 * time.Millisecond)

	return nil
}

// Fini cleans up the backend and stops the HTTP server.
func (b *Backend) Fini() {
	b.finiOnce.Do(func() {
		b.cancel()

		// Close all sessions
		b.sessionsMu.Lock()
		for _, session := range b.sessions {
			session.Close()
		}
		b.sessions = make(map[string]*Session)
		b.sessionsMu.Unlock()

		// Shutdown HTTP server
		if b.server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = b.server.Shutdown(ctx)
			cancel()
		}

		close(b.events)
	})
}

// Size returns the terminal dimensions.
func (b *Backend) Size() (width, height int) {
	b.termMu.RLock()
	defer b.termMu.RUnlock()
	return b.width, b.height
}

// SetContent sets a cell at position (x, y).
func (b *Backend) SetContent(x, y int, mainc rune, comb []rune, style backend.Style) {
	b.termMu.Lock()
	defer b.termMu.Unlock()
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}
	idx := y*b.width + x
	b.cells[idx] = backend.Cell{
		Rune:  mainc,
		Style: style,
	}
}

// SetRow updates an entire row of cells.
func (b *Backend) SetRow(y int, startX int, cells []backend.Cell) {
	b.termMu.Lock()
	defer b.termMu.Unlock()
	if y < 0 || y >= b.height || startX < 0 || startX >= b.width {
		return
	}
	rowStart := y*b.width + startX
	rowEnd := rowStart + len(cells)
	if rowEnd > (y+1)*b.width {
		rowEnd = (y + 1) * b.width
	}
	copy(b.cells[rowStart:rowEnd], cells)
}

// SetRect updates a rectangular region.
func (b *Backend) SetRect(x, y, width, height int, cells []backend.Cell) {
	b.termMu.Lock()
	defer b.termMu.Unlock()
	if width <= 0 || height <= 0 {
		return
	}
	expected := width * height
	if len(cells) < expected {
		return
	}
	for row := 0; row < height; row++ {
		srcStart := row * width
		dstRow := y + row
		if dstRow < 0 || dstRow >= b.height {
			continue
		}
		for col := 0; col < width; col++ {
			dstCol := x + col
			if dstCol < 0 || dstCol >= b.width {
				continue
			}
			idx := dstRow*b.width + dstCol
			b.cells[idx] = cells[srcStart+col]
		}
	}
}

// Show synchronizes the buffer to connected clients.
func (b *Backend) Show() {
	if b.config.Renderer == RendererANSI {
		data := b.renderANSI()
		b.broadcast(data)
	}
}

// Clear clears the screen.
func (b *Backend) Clear() {
	b.termMu.Lock()
	defer b.termMu.Unlock()
	for i := range b.cells {
		b.cells[i] = backend.Cell{Rune: ' ', Style: backend.DefaultStyle()}
	}
}

// HideCursor hides the cursor.
func (b *Backend) HideCursor() {
	b.termMu.Lock()
	b.cursorOn = false
	b.termMu.Unlock()
	if b.config.Renderer == RendererANSI {
		b.broadcast([]byte("\x1b[?25l")) // DECTCEM - hide cursor
	}
}

// ShowCursor shows the cursor.
func (b *Backend) ShowCursor() {
	b.termMu.Lock()
	b.cursorOn = true
	b.termMu.Unlock()
	if b.config.Renderer == RendererANSI {
		b.broadcast([]byte("\x1b[?25h")) // DECTCEM - show cursor
	}
}

// SetCursorPos sets the cursor position.
func (b *Backend) SetCursorPos(x, y int) {
	b.termMu.Lock()
	b.cursorX = x
	b.cursorY = y
	b.termMu.Unlock()
	if b.config.Renderer == RendererANSI {
		// ANSI cursor positioning: ESC [ row ; col H
		// ANSI is 1-indexed, our coords are 0-indexed
		data := fmt.Sprintf("\x1b[%d;%dH", y+1, x+1)
		b.broadcast([]byte(data))
	}
}

// PollEvent blocks until an event is available.
func (b *Backend) PollEvent() terminal.Event {
	select {
	case ev := <-b.events:
		return ev
	case <-b.ctx.Done():
		return nil
	}
}

// PostEvent injects an event into the event queue.
func (b *Backend) PostEvent(ev terminal.Event) error {
	select {
	case b.events <- ev:
		return nil
	case <-b.ctx.Done():
		return fmt.Errorf("backend closed")
	default:
		return fmt.Errorf("event queue full")
	}
}

// Beep emits an audible bell.
func (b *Backend) Beep() {
	b.broadcast([]byte("\x07")) // BEL character
}

// Sync forces a full redraw on next Show().
func (b *Backend) Sync() {
	// Full redraw happens automatically in Show()
}

// Addr returns the actual listen address.
// Useful when using :0 for dynamic port assignment.
func (b *Backend) Addr() net.Addr {
	if b.listener == nil {
		return nil
	}
	return b.listener.Addr()
}

// URL returns the HTTP URL to access the web terminal.
func (b *Backend) URL() string {
	if b.listener == nil {
		return ""
	}
	addr := b.listener.Addr()
	if addr == nil {
		return ""
	}

	scheme := "http"
	if b.config.TLS.Enabled {
		scheme = "https"
	}

	// Handle IPv6 addresses
	host, port, _ := net.SplitHostPort(addr.String())
	if host == "::" || host == "" {
		host = "localhost"
	}

	return fmt.Sprintf("%s://%s:%s%s", scheme, host, port, b.config.Path)
}

// broadcast sends data to all connected sessions.
func (b *Backend) broadcast(data []byte) {
	if len(data) == 0 {
		return
	}

	b.sessionsMu.RLock()
	sessions := make([]*Session, 0, len(b.sessions))
	for _, s := range b.sessions {
		sessions = append(sessions, s)
	}
	b.sessionsMu.RUnlock()

	for _, session := range sessions {
		session.Send(data)
	}
}

// addSession registers a new session.
func (b *Backend) addSession(s *Session) bool {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()

	if b.config.MaxConnections > 0 && len(b.sessions) >= b.config.MaxConnections {
		return false
	}

	b.sessions[s.ID] = s
	b.sessionCount.Add(1)

	if b.config.OnConnect != nil {
		go b.config.OnConnect(s.ID)
	}

	return true
}

// removeSession unregisters a session.
func (b *Backend) removeSession(id string) {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()

	if _, ok := b.sessions[id]; ok {
		delete(b.sessions, id)
		b.sessionCount.Add(-1)

		if b.config.OnDisconnect != nil {
			go b.config.OnDisconnect(id)
		}
	}
}

// renderANSI generates ANSI escape sequences from the current cell buffer.
func (b *Backend) renderANSI() []byte {
	// For efficiency, only render changed cells
	// For simplicity in v1, do full redraw with diff optimization
	return b.renderFullANSI()
}

// renderFullANSI generates a full ANSI representation of the screen.
func (b *Backend) renderFullANSI() []byte {
	b.termMu.RLock()
	defer b.termMu.RUnlock()
	return b.renderFullANSILocked()
}

func (b *Backend) renderFullANSILocked() []byte {
	var buf []byte

	// Clear screen and home cursor
	buf = append(buf, "\x1b[2J\x1b[H"...) // ED2 + CUP

	var lastFg, lastBg backend.Color
	var lastAttrs backend.AttrMask

	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			cell := b.cells[y*b.width+x]
			fg, bg, attrs := cell.Style.Decompose()

			// Output style changes
			if fg != lastFg || bg != lastBg || attrs != lastAttrs {
				buf = append(buf, b.styleToANSI(fg, bg, attrs)...)
				lastFg = fg
				lastBg = bg
				lastAttrs = attrs
			}

			// Output rune
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			buf = append(buf, string(r)...)
		}

		// Move to next line (except last row)
		if y < b.height-1 {
			buf = append(buf, '\r', '\n')
		}
	}

	// Restore cursor position if visible
	if b.cursorOn {
		buf = append(buf, fmt.Sprintf("\x1b[%d;%dH", b.cursorY+1, b.cursorX+1)...)
	}

	return buf
}

// styleToANSI converts a style to ANSI escape sequences.
func (b *Backend) styleToANSI(fg, bg backend.Color, attrs backend.AttrMask) []byte {
	var codes []int

	// Reset
	codes = append(codes, 0)

	// Attributes
	if attrs&backend.AttrBold != 0 {
		codes = append(codes, 1)
	}
	if attrs&backend.AttrDim != 0 {
		codes = append(codes, 2)
	}
	if attrs&backend.AttrItalic != 0 {
		codes = append(codes, 3)
	}
	if attrs&backend.AttrUnderline != 0 {
		codes = append(codes, 4)
	}
	if attrs&backend.AttrBlink != 0 {
		codes = append(codes, 5)
	}
	if attrs&backend.AttrReverse != 0 {
		codes = append(codes, 7)
	}
	if attrs&backend.AttrStrikeThrough != 0 {
		codes = append(codes, 9)
	}

	// Foreground color
	if fg == backend.ColorDefault {
		codes = append(codes, 39)
	} else if fg.IsRGB() {
		r, g, b := fg.RGB()
		codes = append(codes, 38, 2, int(r), int(g), int(b))
	} else {
		codes = append(codes, 38, 5, int(fg))
	}

	// Background color
	if bg == backend.ColorDefault {
		codes = append(codes, 49)
	} else if bg.IsRGB() {
		r, g, bCol := bg.RGB()
		codes = append(codes, 48, 2, int(r), int(g), int(bCol))
	} else {
		codes = append(codes, 48, 5, int(bg))
	}

	// Build escape sequence
	var buf []byte
	buf = append(buf, "\x1b["...)
	for i, c := range codes {
		if i > 0 {
			buf = append(buf, ';')
		}
		buf = append(buf, fmt.Sprintf("%d", c)...)
	}
	buf = append(buf, 'm')
	return buf
}

// Ensure Backend implements backend.Backend
var _ backend.Backend = (*Backend)(nil)

// Ensure Backend implements optional interfaces
var _ backend.RowWriter = (*Backend)(nil)
var _ backend.RectWriter = (*Backend)(nil)
