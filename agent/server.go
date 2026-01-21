package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/odvcencio/fluffy-ui/keybind"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// ServerOptions configures the agent interaction server.
// TestMode should only be enabled in tests; it bypasses text gating.
type ServerOptions struct {
	Addr            string
	App             *runtime.App
	Agent           *Agent
	AllowText       bool
	TestMode        bool
	Token           string
	SnapshotTimeout time.Duration
}

// Capabilities describes server features exposed to clients.
type Capabilities struct {
	AllowText bool `json:"allow_text"`
	TestMode  bool `json:"test_mode"`
}

// Server exposes an out-of-process JSONL API for agent interaction.
type Server struct {
	opts     ServerOptions
	agent    *Agent
	listener net.Listener
	unixPath string
	mu       sync.Mutex
}

// NewServer validates options and constructs a server.
func NewServer(opts ServerOptions) (*Server, error) {
	if strings.TrimSpace(opts.Addr) == "" {
		return nil, errors.New("agent server address is required")
	}
	if opts.Agent == nil {
		if opts.App == nil {
			return nil, errors.New("agent server requires App or Agent")
		}
		opts.Agent = New(Config{App: opts.App})
	}
	if opts.SnapshotTimeout <= 0 {
		opts.SnapshotTimeout = 2 * time.Second
	}

	return &Server{
		opts:  opts,
		agent: opts.Agent,
	}, nil
}

// Serve starts listening and blocks until the context is done or the listener closes.
func (s *Server) Serve(ctx context.Context) error {
	if s == nil {
		return errors.New("agent server is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ln, unixPath, err := listenAgentAddr(s.opts.Addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = ln
	s.unixPath = unixPath
	s.mu.Unlock()

	defer s.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return ctx.Err()
			}
			continue
		}
		go s.handleConn(ctx, conn)
	}
}

// Close stops the listener and cleans up any unix socket file.
func (s *Server) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	ln := s.listener
	s.listener = nil
	unixPath := s.unixPath
	s.unixPath = ""
	s.mu.Unlock()

	var err error
	if ln != nil {
		err = ln.Close()
	}
	if unixPath != "" {
		_ = os.Remove(unixPath)
	}
	return err
}

type request struct {
	ID          int    `json:"id,omitempty"`
	Type        string `json:"type"`
	Token       string `json:"token,omitempty"`
	Key         string `json:"key,omitempty"`
	Text        string `json:"text,omitempty"`
	X           int    `json:"x,omitempty"`
	Y           int    `json:"y,omitempty"`
	Button      string `json:"button,omitempty"`
	Action      string `json:"action,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Alt         bool   `json:"alt,omitempty"`
	Ctrl        bool   `json:"ctrl,omitempty"`
	Shift       bool   `json:"shift,omitempty"`
	IncludeText bool   `json:"include_text,omitempty"`
}

type response struct {
	ID           int           `json:"id,omitempty"`
	OK           bool          `json:"ok,omitempty"`
	Error        string        `json:"error,omitempty"`
	Message      string        `json:"message,omitempty"`
	Snapshot     *Snapshot     `json:"snapshot,omitempty"`
	Capabilities *Capabilities `json:"capabilities,omitempty"`
}

type session struct {
	authed bool
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	enc := json.NewEncoder(conn)
	enc.SetEscapeHTML(false)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	sess := &session{authed: s.opts.Token == ""}

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(response{
				OK:      false,
				Error:   "bad_json",
				Message: err.Error(),
			})
			continue
		}

		resp := s.handleRequest(ctx, sess, req)
		_ = enc.Encode(resp)
	}
}

func (s *Server) handleRequest(ctx context.Context, sess *session, req request) response {
	if strings.TrimSpace(req.Type) == "" {
		return response{ID: req.ID, OK: false, Error: "missing_type"}
	}

	if !sess.authed && req.Type != "hello" {
		return response{ID: req.ID, OK: false, Error: "unauthorized", Message: "authentication required"}
	}

	switch req.Type {
	case "hello":
		if s.opts.Token != "" && req.Token != s.opts.Token {
			return response{ID: req.ID, OK: false, Error: "unauthorized", Message: "invalid token"}
		}
		sess.authed = true
		return response{
			ID: req.ID,
			OK: true,
			Capabilities: &Capabilities{
				AllowText: s.opts.AllowText,
				TestMode:  s.opts.TestMode,
			},
		}
	case "ping":
		return response{ID: req.ID, OK: true}
	case "snapshot":
		includeText := req.IncludeText
		if includeText && !s.opts.AllowText && !s.opts.TestMode {
			return response{ID: req.ID, OK: false, Error: "text_disabled"}
		}
		timeout := s.opts.SnapshotTimeout
		if timeout <= 0 {
			timeout = 2 * time.Second
		}
		ctxSnap, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		snap, err := s.agent.SnapshotWithContext(ctxSnap, SnapshotOptions{
			IncludeText: includeText,
		})
		if err != nil {
			return response{ID: req.ID, OK: false, Error: "snapshot_failed", Message: err.Error()}
		}
		return response{ID: req.ID, OK: true, Snapshot: &snap}
	case "key":
		press, err := parseKeyPress(req.Key)
		if err != nil {
			return response{ID: req.ID, OK: false, Error: "invalid_key", Message: err.Error()}
		}
		err = s.agent.SendKeyMsg(runtime.KeyMsg{
			Key:   press.Key,
			Rune:  press.Rune,
			Alt:   press.Alt,
			Ctrl:  press.Ctrl,
			Shift: press.Shift,
		})
		if err != nil {
			return response{ID: req.ID, OK: false, Error: "send_key_failed", Message: err.Error()}
		}
		return response{ID: req.ID, OK: true}
	case "text":
		if strings.TrimSpace(req.Text) == "" {
			return response{ID: req.ID, OK: false, Error: "missing_text"}
		}
		if err := s.agent.SendKeyString(req.Text); err != nil {
			return response{ID: req.ID, OK: false, Error: "send_text_failed", Message: err.Error()}
		}
		return response{ID: req.ID, OK: true}
	case "mouse":
		button, err := parseMouseButton(req.Button)
		if err != nil {
			return response{ID: req.ID, OK: false, Error: "invalid_mouse_button", Message: err.Error()}
		}
		action, err := parseMouseAction(req.Action)
		if err != nil {
			return response{ID: req.ID, OK: false, Error: "invalid_mouse_action", Message: err.Error()}
		}
		err = s.agent.SendMouse(runtime.MouseMsg{
			X:      req.X,
			Y:      req.Y,
			Button: button,
			Action: action,
			Alt:    req.Alt,
			Ctrl:   req.Ctrl,
			Shift:  req.Shift,
		})
		if err != nil {
			return response{ID: req.ID, OK: false, Error: "send_mouse_failed", Message: err.Error()}
		}
		return response{ID: req.ID, OK: true}
	case "paste":
		if err := s.agent.SendPaste(req.Text); err != nil {
			return response{ID: req.ID, OK: false, Error: "send_paste_failed", Message: err.Error()}
		}
		return response{ID: req.ID, OK: true}
	case "resize":
		if req.Width <= 0 || req.Height <= 0 {
			return response{ID: req.ID, OK: false, Error: "invalid_resize"}
		}
		if err := s.agent.SendResize(req.Width, req.Height); err != nil {
			return response{ID: req.ID, OK: false, Error: "send_resize_failed", Message: err.Error()}
		}
		return response{ID: req.ID, OK: true}
	default:
		return response{ID: req.ID, OK: false, Error: "unknown_type"}
	}
}

func listenAgentAddr(addr string) (net.Listener, string, error) {
	switch {
	case strings.HasPrefix(addr, "unix:"):
		path := strings.TrimPrefix(addr, "unix:")
		if strings.TrimSpace(path) == "" {
			return nil, "", errors.New("unix socket path is required")
		}
		_ = os.Remove(path)
		ln, err := net.Listen("unix", path)
		return ln, path, err
	case strings.HasPrefix(addr, "tcp:"):
		host := strings.TrimPrefix(addr, "tcp:")
		if strings.TrimSpace(host) == "" {
			return nil, "", errors.New("tcp address is required")
		}
		ln, err := net.Listen("tcp", host)
		return ln, "", err
	default:
		return nil, "", fmt.Errorf("unsupported address %q (use unix: or tcp:)", addr)
	}
}

func parseKeyPress(key string) (keybind.KeyPress, error) {
	seq, err := keybind.ParseKeySequence(key)
	if err != nil {
		return keybind.KeyPress{}, err
	}
	if len(seq.Sequence) != 1 {
		return keybind.KeyPress{}, errors.New("key must be a single press")
	}
	return seq.Sequence[0], nil
}

func parseMouseButton(button string) (runtime.MouseButton, error) {
	switch strings.ToLower(strings.TrimSpace(button)) {
	case "none", "":
		return runtime.MouseNone, nil
	case "left":
		return runtime.MouseLeft, nil
	case "middle":
		return runtime.MouseMiddle, nil
	case "right":
		return runtime.MouseRight, nil
	case "wheel_up", "wheelup", "up":
		return runtime.MouseWheelUp, nil
	case "wheel_down", "wheeldown", "down":
		return runtime.MouseWheelDown, nil
	default:
		return runtime.MouseNone, fmt.Errorf("unknown button %q", button)
	}
}

func parseMouseAction(action string) (runtime.MouseAction, error) {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "press", "":
		return runtime.MousePress, nil
	case "release":
		return runtime.MouseRelease, nil
	case "move":
		return runtime.MouseMove, nil
	default:
		return runtime.MouseMove, fmt.Errorf("unknown action %q", action)
	}
}
