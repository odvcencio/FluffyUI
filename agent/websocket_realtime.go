//go:build !js

package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

// RealTimeWebSocketServer provides bidirectional real-time communication
type RealTimeWebSocketServer struct {
	server         *RealTimeServer
	upgrader       websocket.Upgrader
	connections    sync.Map // map[*websocket.Conn]*realTimeConnection
	allowedOrigins []string
}

// RealTimeWSOptions configures the real-time WebSocket server
type RealTimeWSOptions struct {
	EnhancedServerOptions
	AllowedOrigins []string
}

// realTimeConnection wraps a WebSocket connection with session info
type realTimeConnection struct {
	conn       *websocket.Conn
	sessionID  string
	subscriber *RealTimeSubscriber
	writeMu    sync.Mutex
	done       chan struct{}
}

// wsMessage represents an incoming WebSocket message
type wsMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// wsResponse represents an outgoing WebSocket message
type wsResponse struct {
	Type  string `json:"type"`
	ID    string `json:"id,omitempty"`
	OK    bool   `json:"ok,omitempty"`
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

// NewRealTimeWebSocketServer creates a new real-time WebSocket server
func NewRealTimeWebSocketServer(opts RealTimeWSOptions) (*RealTimeWebSocketServer, error) {
	server, err := NewRealTimeServer(opts.EnhancedServerOptions)
	if err != nil {
		return nil, err
	}

	return &RealTimeWebSocketServer{
		server:         server,
		allowedOrigins: opts.AllowedOrigins,
	}, nil
}

// Start begins the WebSocket server
func (s *RealTimeWebSocketServer) Start() error {
	return s.server.Start()
}

// Stop stops the WebSocket server
func (s *RealTimeWebSocketServer) Stop() error {
	// Close all connections
	s.connections.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)
		conn.Close()
		return true
	})

	return s.server.Stop()
}

// ServeHTTP implements http.Handler
func (s *RealTimeWebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s == nil || s.server == nil {
		http.Error(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	upgrader := s.upgrader
	if upgrader.CheckOrigin == nil {
		upgrader.CheckOrigin = s.checkOrigin
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.handleConnection(conn)
}

// handleConnection manages a single WebSocket connection
func (s *RealTimeWebSocketServer) handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	sessionID := generateSessionID()

	// Create subscription for real-time events
	subscriber := s.server.Subscribe(sessionID, DefaultEventFilters())
	if subscriber == nil {
		return
	}
	defer s.server.Unsubscribe(subscriber.ID)

	connInfo := &realTimeConnection{
		conn:       conn,
		sessionID:  sessionID,
		subscriber: subscriber,
		done:       make(chan struct{}),
	}

	s.connections.Store(conn, connInfo)
	defer s.connections.Delete(conn)

	// Start goroutines for reading and writing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// Write loop - sends real-time events to client
	go s.writeLoop(ctx, connInfo, &wg)

	// Read loop - receives commands from client
	go s.readLoop(ctx, connInfo, &wg)

	wg.Wait()
}

// writeLoop sends events to the client
func (s *RealTimeWebSocketServer) writeLoop(ctx context.Context, conn *realTimeConnection, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.done:
			return
		case event := <-conn.subscriber.Events:
			resp := wsResponse{
				Type: "event",
				Data: event,
			}
			if err := s.writeResponse(conn, resp); err != nil {
				return
			}
		}
	}
}

// readLoop receives and processes commands from the client
func (s *RealTimeWebSocketServer) readLoop(ctx context.Context, conn *realTimeConnection, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var msg wsMessage
		if err := conn.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error
			}
			close(conn.done)
			return
		}

		go s.handleMessage(ctx, conn, msg)
	}
}

// handleMessage processes a single message
func (s *RealTimeWebSocketServer) handleMessage(ctx context.Context, conn *realTimeConnection, msg wsMessage) {
	var resp wsResponse
	resp.ID = msg.ID

	switch msg.Type {
	case "ping":
		resp.Type = "pong"
		resp.OK = true

	case "snapshot":
		resp = s.handleSnapshot(conn, msg)

	case "find":
		resp = s.handleFind(conn, msg)

	case "action":
		resp = s.handleAction(ctx, conn, msg)

	case "wait":
		resp = s.handleWait(ctx, conn, msg)

	case "subscribe":
		resp = s.handleSubscribe(conn, msg)

	case "unsubscribe":
		resp = s.handleUnsubscribe(conn, msg)

	case "key":
		resp = s.handleKey(conn, msg)

	case "text":
		resp = s.handleText(conn, msg)

	case "mouse":
		resp = s.handleMouse(conn, msg)

	case "task":
		resp = s.handleTask(conn, msg)

	default:
		resp.Type = "error"
		resp.Error = "unknown message type: " + msg.Type
	}

	_ = s.writeResponse(conn, resp)
}

// handleSnapshot captures a UI snapshot
func (s *RealTimeWebSocketServer) handleSnapshot(conn *realTimeConnection, msg wsMessage) wsResponse {
	snap := s.server.agent.Snapshot()
	return wsResponse{
		Type: "snapshot",
		ID:   msg.ID,
		OK:   true,
		Data: snap,
	}
}

// handleFind searches for widgets
func (s *RealTimeWebSocketServer) handleFind(conn *realTimeConnection, msg wsMessage) wsResponse {
	var req struct {
		By    string `json:"by"`    // "label", "id", "role"
		Value string `json:"value"` // search value
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	snap := s.server.agent.Snapshot()
	var result any

	switch req.By {
	case "label":
		result = findByLabelIn(snap.Widgets, req.Value)
	case "id":
		result = findByIDIn(snap.Widgets, req.Value)
	case "role":
		var results []WidgetInfo
		role := accessibility.Role(req.Value)
		findByRoleIn(snap.Widgets, role, &results)
		result = results
	default:
		return wsResponse{Type: "error", ID: msg.ID, Error: "unknown find type: " + req.By}
	}

	return wsResponse{
		Type: "find_result",
		ID:   msg.ID,
		OK:   true,
		Data: result,
	}
}

// handleAction performs an action on a widget
func (s *RealTimeWebSocketServer) handleAction(ctx context.Context, conn *realTimeConnection, msg wsMessage) wsResponse {
	var req struct {
		WidgetID string `json:"widget_id,omitempty"`
		Label    string `json:"label,omitempty"`
		Action   string `json:"action"`          // "focus", "activate", "type", "clear"
		Value    string `json:"value,omitempty"` // for "type" action
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	var err error
	switch req.Action {
	case "focus":
		if req.Label != "" {
			err = s.server.agent.Focus(req.Label)
		}
	case "activate":
		if req.Label != "" {
			err = s.server.agent.Activate(req.Label)
		}
	case "type":
		if req.Label != "" && req.Value != "" {
			err = s.server.agent.Type(req.Label, req.Value)
		}
	default:
		return wsResponse{Type: "error", ID: msg.ID, Error: "unknown action: " + req.Action}
	}

	if err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	return wsResponse{Type: "action_complete", ID: msg.ID, OK: true}
}

// handleWait waits for a condition
func (s *RealTimeWebSocketServer) handleWait(ctx context.Context, conn *realTimeConnection, msg wsMessage) wsResponse {
	var req struct {
		For     string        `json:"for"`     // "widget", "text", "focus", "value"
		Value   string        `json:"value"`   // widget label, text to find, etc.
		Timeout time.Duration `json:"timeout"` // milliseconds
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	if req.Timeout == 0 {
		req.Timeout = 5000 // Default 5 seconds
	}
	req.Timeout *= time.Millisecond

	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	var result any
	var err error

	switch req.For {
	case "widget":
		result, err = s.server.WaitForWidget(ctx, req.Value, req.Timeout)
	case "text":
		err = s.server.WaitForText(ctx, req.Value, req.Timeout)
	case "focus":
		err = s.server.WaitForFocus(ctx, req.Value, req.Timeout)
	default:
		return wsResponse{Type: "error", ID: msg.ID, Error: "unknown wait type: " + req.For}
	}

	if err != nil {
		return wsResponse{Type: "wait_timeout", ID: msg.ID, Error: err.Error()}
	}

	return wsResponse{
		Type: "wait_complete",
		ID:   msg.ID,
		OK:   true,
		Data: result,
	}
}

// handleSubscribe updates subscription filters
func (s *RealTimeWebSocketServer) handleSubscribe(conn *realTimeConnection, msg wsMessage) wsResponse {
	var filters EventFilters
	if err := json.Unmarshal(msg.Payload, &filters); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	// Update subscription
	conn.subscriber.Filters = filters

	return wsResponse{Type: "subscribed", ID: msg.ID, OK: true}
}

// handleUnsubscribe removes subscription
func (s *RealTimeWebSocketServer) handleUnsubscribe(conn *realTimeConnection, msg wsMessage) wsResponse {
	s.server.Unsubscribe(conn.subscriber.ID)
	return wsResponse{Type: "unsubscribed", ID: msg.ID, OK: true}
}

// handleKey sends a key press
func (s *RealTimeWebSocketServer) handleKey(conn *realTimeConnection, msg wsMessage) wsResponse {
	var req struct {
		Key   string `json:"key"`
		Rune  string `json:"rune,omitempty"`
		Alt   bool   `json:"alt,omitempty"`
		Ctrl  bool   `json:"ctrl,omitempty"`
		Shift bool   `json:"shift,omitempty"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	press, err := parseKeyPress(req.Key)
	if err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	if req.Rune != "" && len(req.Rune) > 0 {
		press.Rune = rune(req.Rune[0])
	}
	press.Alt = req.Alt
	press.Ctrl = req.Ctrl
	press.Shift = req.Shift

	if err := s.server.agent.SendKeyMsg(runtime.KeyMsg{
		Key:   press.Key,
		Rune:  press.Rune,
		Alt:   press.Alt,
		Ctrl:  press.Ctrl,
		Shift: press.Shift,
	}); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	return wsResponse{Type: "key_sent", ID: msg.ID, OK: true}
}

// handleText sends text input
func (s *RealTimeWebSocketServer) handleText(conn *realTimeConnection, msg wsMessage) wsResponse {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	if err := s.server.agent.SendKeyString(req.Text); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	return wsResponse{Type: "text_sent", ID: msg.ID, OK: true}
}

// handleMouse sends a mouse event
func (s *RealTimeWebSocketServer) handleMouse(conn *realTimeConnection, msg wsMessage) wsResponse {
	var req struct {
		X      int    `json:"x"`
		Y      int    `json:"y"`
		Button string `json:"button,omitempty"`
		Action string `json:"action,omitempty"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	button, _ := parseMouseButton(req.Button)
	action, _ := parseMouseAction(req.Action)

	if err := s.server.agent.SendMouse(runtime.MouseMsg{
		X:      req.X,
		Y:      req.Y,
		Button: button,
		Action: action,
	}); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	return wsResponse{Type: "mouse_sent", ID: msg.ID, OK: true}
}

// handleTask manages background tasks
func (s *RealTimeWebSocketServer) handleTask(conn *realTimeConnection, msg wsMessage) wsResponse {
	var req struct {
		Action      string `json:"action"` // "submit", "status", "cancel"
		TaskID      string `json:"task_id,omitempty"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return wsResponse{Type: "error", ID: msg.ID, Error: err.Error()}
	}

	switch req.Action {
	case "submit":
		// Note: Actual task function would need to be provided
		return wsResponse{Type: "task_submitted", ID: msg.ID, OK: true, Data: map[string]string{"task_id": "demo-task"}}

	case "status":
		task := s.server.taskManager.Get(req.TaskID)
		if task == nil {
			return wsResponse{Type: "error", ID: msg.ID, Error: "task not found"}
		}
		return wsResponse{Type: "task_status", ID: msg.ID, OK: true, Data: task.Stats()}

	case "cancel":
		if !s.server.taskManager.Cancel(req.TaskID) {
			return wsResponse{Type: "error", ID: msg.ID, Error: "task not found"}
		}
		return wsResponse{Type: "task_cancelled", ID: msg.ID, OK: true}

	default:
		return wsResponse{Type: "error", ID: msg.ID, Error: "unknown task action: " + req.Action}
	}
}

// writeResponse sends a response to the client
func (s *RealTimeWebSocketServer) writeResponse(conn *realTimeConnection, resp wsResponse) error {
	conn.writeMu.Lock()
	defer conn.writeMu.Unlock()

	return conn.conn.WriteJSON(resp)
}

// checkOrigin validates request origin
func (s *RealTimeWebSocketServer) checkOrigin(r *http.Request) bool {
	if s == nil {
		return false
	}
	if len(s.allowedOrigins) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	for _, allowed := range s.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// Broadcast sends a message to all connected clients
func (s *RealTimeWebSocketServer) Broadcast(msg any) {
	if s == nil {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	s.connections.Range(func(key, value any) bool {
		conn := value.(*realTimeConnection)
		conn.writeMu.Lock()
		err := conn.conn.WriteMessage(websocket.TextMessage, data)
		conn.writeMu.Unlock()
		if err != nil {
			conn.conn.Close()
			s.connections.Delete(key)
		}
		return true
	})
}

// ConnectionCount returns the number of active connections
func (s *RealTimeWebSocketServer) ConnectionCount() int {
	count := 0
	s.connections.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// Ensure RealTimeWebSocketServer implements http.Handler
var _ http.Handler = (*RealTimeWebSocketServer)(nil)

// RealTimeHandler returns an http.Handler for the real-time server
func RealTimeHandler(opts RealTimeWSOptions) (http.Handler, error) {
	return NewRealTimeWebSocketServer(opts)
}
