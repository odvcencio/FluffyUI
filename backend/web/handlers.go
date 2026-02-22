//go:build !js

package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/odvcencio/fluffyui/terminal"
)

// handleIndex serves the terminal HTML page.
func (b *Backend) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if !b.checkAuth(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="FluffyUI"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML(b.config.Title, b.config.WSPath, b.config.Chromeless))
}

// handleWebSocket handles WebSocket connections.
func (b *Backend) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if !b.checkAuth(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check token auth via URL parameter
	if b.config.Auth.Enabled && b.config.Auth.Token != "" {
		token := r.URL.Query().Get("token")
		if token != b.config.Auth.Token {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
	}

	// Upgrade to WebSocket
	conn, err := b.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Create session
	session := &Session{
		ID:       generateSessionID(),
		Conn:     conn,
		Backend:  b,
		LastSeen: time.Now(),
	}

	// Check connection limit
	if !b.addSession(session) {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","message":"Server at capacity"}`))
		conn.Close()
		return
	}

	// Handle the connection
	go session.handle()
}

// checkAuth verifies authentication for the request.
func (b *Backend) checkAuth(r *http.Request) bool {
	if !b.config.Auth.Enabled {
		return true
	}

	// Check basic auth
	if b.config.Auth.Username != "" || b.config.Auth.Password != "" {
		user, pass, ok := r.BasicAuth()
		if !ok {
			return false
		}
		if user != b.config.Auth.Username || pass != b.config.Auth.Password {
			return false
		}
	}

	return true
}

// handle manages the WebSocket connection lifecycle.
func (s *Session) handle() {
	defer func() {
		s.Close()
		s.Backend.removeSession(s.ID)
	}()

	// Set up ping/pong
	s.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	s.Conn.SetPongHandler(func(string) error {
		s.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		s.mu.Lock()
		s.LastSeen = time.Now()
		s.mu.Unlock()
		return nil
	})

	// Read messages from client
	// Initial content is sent after the browser's first resize message
	// to avoid rendering at the wrong dimensions (80x24 default).
	for {
		msgType, data, err := s.Conn.ReadMessage()
		if err != nil {
			return
		}

		s.mu.Lock()
		s.LastSeen = time.Now()
		s.mu.Unlock()

		switch msgType {
		case websocket.BinaryMessage:
			// Binary data is treated as terminal input (ANSI keystrokes)
			s.handleInput(data)
		case websocket.TextMessage:
			// Text data is JSON control messages
			s.handleControlMessage(data)
		}
	}
}

// handleInput processes terminal input from the client.
func (s *Session) handleInput(data []byte) {
	// Parse ANSI input sequences
	// For now, handle basic characters and common escape sequences
	for i := 0; i < len(data); i++ {
		b := data[i]

		switch b {
		case 0x1b: // ESC
			if i+1 < len(data) {
				next := data[i+1]
				switch next {
				case '[': // CSI sequence
					i += 2
					ev, consumed := s.parseCSI(data[i:])
					if consumed > 0 {
						i += consumed - 1
						if ev != nil {
							s.Backend.PostEvent(ev)
						}
					}
				case 'O': // SS3 sequence
					i += 2
					ev, consumed := s.parseSS3(data[i:])
					if consumed > 0 {
						i += consumed - 1
						if ev != nil {
							s.Backend.PostEvent(ev)
						}
					}
				default:
					// Alt+key
					if i+1 < len(data) {
						s.Backend.PostEvent(terminal.KeyEvent{
							Key:  terminal.KeyRune,
							Rune: rune(data[i+1]),
							Alt:  true,
						})
						i++
					}
				}
			}
		case 0x7f: // DEL (backspace)
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyBackspace})
		case 0x0d: // CR (Enter)
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyEnter})
		case 0x09: // Tab
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyTab})
		case 0x01: // Ctrl+A
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlA})
		case 0x02: // Ctrl+B
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlB})
		case 0x03: // Ctrl+C
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlC})
		case 0x04: // Ctrl+D
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlD})
		case 0x06: // Ctrl+F
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlF})
		case 0x10: // Ctrl+P
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlP})
		case 0x16: // Ctrl+V
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlV})
		case 0x18: // Ctrl+X
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlX})
		case 0x19: // Ctrl+Y
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlY})
		case 0x1a: // Ctrl+Z
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlZ})
		case 0x05: // Ctrl+E
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlE})
		case 0x0B: // Ctrl+K
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlK})
		case 0x0C: // Ctrl+L
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlL})
		case 0x0E: // Ctrl+N
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlN})
		case 0x0F: // Ctrl+O
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlO})
		case 0x11: // Ctrl+Q
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlQ})
		case 0x12: // Ctrl+R
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlR})
		case 0x13: // Ctrl+S
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlS})
		case 0x14: // Ctrl+T
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlT})
		case 0x15: // Ctrl+U
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlU})
		case 0x17: // Ctrl+W
			s.Backend.PostEvent(terminal.KeyEvent{Key: terminal.KeyCtrlW})
		default:
			if b >= 32 && b < 127 {
				s.Backend.PostEvent(terminal.KeyEvent{
					Key:  terminal.KeyRune,
					Rune: rune(b),
				})
			}
		}
	}
}

// parseCSI parses a CSI escape sequence and returns the corresponding key event.
// Also returns a bool indicating if this was a mouse event.
func (s *Session) parseCSI(data []byte) (terminal.Event, int) {
	if len(data) == 0 {
		return nil, 0
	}

	// Find end of CSI sequence (letter or ~)
	end := 0
	for end < len(data) && ((data[end] >= '0' && data[end] <= '9') || data[end] == ';' || data[end] == '?' || data[end] == '<') {
		end++
	}
	if end >= len(data) {
		return nil, 0
	}
	end++ // Include the final character

	seq := string(data[:end])
	final := seq[len(seq)-1]

	// Parse parameters
	var params []int
	var current int
	for _, c := range seq[:len(seq)-1] {
		if c >= '0' && c <= '9' {
			current = current*10 + int(c-'0')
		} else if c == ';' {
			params = append(params, current)
			current = 0
		}
	}
	params = append(params, current)

	// Handle mouse sequences: ESC [ < Cb ; Cx ; Cy M/m
	if len(seq) > 1 && seq[0] == '<' && (final == 'M' || final == 'm') {
		return s.parseMouseCSI(params, final == 'M', end)
	}

	// Handle regular CSI sequences
	switch final {
	case 'A':
		return terminal.KeyEvent{Key: terminal.KeyUp}, end
	case 'B':
		return terminal.KeyEvent{Key: terminal.KeyDown}, end
	case 'C':
		return terminal.KeyEvent{Key: terminal.KeyRight}, end
	case 'D':
		return terminal.KeyEvent{Key: terminal.KeyLeft}, end
	case 'H':
		return terminal.KeyEvent{Key: terminal.KeyHome}, end
	case 'F':
		return terminal.KeyEvent{Key: terminal.KeyEnd}, end
	case '~':
		if len(params) > 0 {
			switch params[0] {
			case 1:
				return terminal.KeyEvent{Key: terminal.KeyHome}, end
			case 2:
				return terminal.KeyEvent{Key: terminal.KeyInsert}, end
			case 3:
				return terminal.KeyEvent{Key: terminal.KeyDelete}, end
			case 4:
				return terminal.KeyEvent{Key: terminal.KeyEnd}, end
			case 5:
				return terminal.KeyEvent{Key: terminal.KeyPageUp}, end
			case 6:
				return terminal.KeyEvent{Key: terminal.KeyPageDown}, end
			}
		}
	}

	return nil, end
}

// parseSS3 parses an SS3 escape sequence.
func (s *Session) parseSS3(data []byte) (terminal.Event, int) {
	if len(data) == 0 {
		return nil, 0
	}

	switch data[0] {
	case 'P':
		return terminal.KeyEvent{Key: terminal.KeyF1}, 1
	case 'Q':
		return terminal.KeyEvent{Key: terminal.KeyF2}, 1
	case 'R':
		return terminal.KeyEvent{Key: terminal.KeyF3}, 1
	case 'S':
		return terminal.KeyEvent{Key: terminal.KeyF4}, 1
	}

	return nil, 1
}

// parseMouseCSI parses a mouse CSI sequence (SGR 1006 mode).
// Returns a MouseEvent as terminal.Event and the number of consumed bytes.
func (s *Session) parseMouseCSI(params []int, pressed bool, consumed int) (terminal.Event, int) {
	if len(params) < 3 {
		return nil, consumed
	}

	cb := params[0]
	x := params[1] - 1 // Convert to 0-indexed
	y := params[2] - 1 // Convert to 0-indexed

	// cb encoding: button (0-2), shift (4), meta/alt (8), ctrl (16), motion (32), button 4-11
	btn := cb & 3
	if cb&64 != 0 {
		btn += 4 // Buttons 4-7 (wheel)
	}

	var button terminal.MouseButton
	switch btn {
	case 0:
		button = terminal.MouseLeft
	case 1:
		button = terminal.MouseMiddle
	case 2:
		button = terminal.MouseRight
	case 4:
		button = terminal.MouseWheelUp
	case 5:
		button = terminal.MouseWheelDown
	}

	action := terminal.MouseRelease
	if pressed {
		action = terminal.MousePress
		if cb&32 != 0 {
			action = terminal.MouseMove
		}
	}

	// Extract modifiers
	shift := cb&4 != 0
	alt := cb&8 != 0
	ctrl := cb&16 != 0

	return terminal.MouseEvent{
		X:      x,
		Y:      y,
		Button: button,
		Action: action,
		Shift:  shift,
		Alt:    alt,
		Ctrl:   ctrl,
	}, consumed
}

// controlMessage is a JSON message from the client.
type controlMessage struct {
	Type   string `json:"type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// handleControlMessage processes JSON control messages.
func (s *Session) handleControlMessage(data []byte) {
	var msg controlMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "resize":
		if msg.Width > 0 && msg.Height > 0 {
			s.Backend.Resize(msg.Width, msg.Height)
		}
	}
}

// Send transmits data to the client.
func (s *Session) Send(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Conn == nil {
		return fmt.Errorf("connection closed")
	}

	return s.Conn.WriteMessage(websocket.BinaryMessage, data)
}

// Close closes the session connection.
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Conn != nil {
		return s.Conn.Close()
	}
	return nil
}

// generateSessionID creates a unique session identifier.
func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
