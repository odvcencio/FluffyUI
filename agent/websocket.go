package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type wsConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *wsConnection) write(messageType int, data []byte) error {
	if c == nil || c.conn == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(messageType, data)
}

// WebSocketServer exposes the agent JSON API over WebSockets.
type WebSocketServer struct {
	server         *Server
	upgrader       websocket.Upgrader
	connections    sync.Map
	allowedOrigins []string // Empty = allow all (dev mode only!)
}

// WebSocketOptions configures the WebSocket server.
type WebSocketOptions struct {
	ServerOptions
	// AllowedOrigins restricts which origins can connect.
	// Empty slice allows all origins (insecure, dev only).
	// Use []string{"*"} explicitly to allow all in production.
	AllowedOrigins []string
}

// NewWebSocketServer builds a WebSocket server with the provided options.
func NewWebSocketServer(opts ServerOptions) (*WebSocketServer, error) {
	return NewWebSocketServerWithOptions(WebSocketOptions{ServerOptions: opts})
}

// NewWebSocketServerWithOptions builds a WebSocket server with full options.
func NewWebSocketServerWithOptions(opts WebSocketOptions) (*WebSocketServer, error) {
	normalized, agent, err := normalizeServerOptions(opts.ServerOptions)
	if err != nil {
		return nil, err
	}
	srv := &Server{
		opts:  normalized,
		agent: agent,
	}
	return &WebSocketServer{
		server:         srv,
		allowedOrigins: opts.AllowedOrigins,
	}, nil
}

// ServeHTTP upgrades the connection and processes JSON messages.
func (s *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s == nil || s.server == nil {
		http.Error(w, "agent server not configured", http.StatusServiceUnavailable)
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
	defer conn.Close()

	wrapped := &wsConnection{conn: conn}
	s.connections.Store(conn, wrapped)
	defer s.connections.Delete(conn)

	sess := &session{authed: s.server.opts.Token == ""}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		resp := s.handleMessage(r.Context(), sess, message)
		if len(resp) > 0 {
			_ = wrapped.write(websocket.TextMessage, resp)
		}
	}
}

// Broadcast sends a JSON message to all active connections.
func (s *WebSocketServer) Broadcast(msg any) {
	if s == nil {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.connections.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)
		wrapped := value.(*wsConnection)
		if err := wrapped.write(websocket.TextMessage, data); err != nil {
			_ = conn.Close()
			s.connections.Delete(conn)
		}
		return true
	})
}

// NotifyChange broadcasts a structured change message.
func (s *WebSocketServer) NotifyChange(changeType string, data any) {
	s.Broadcast(map[string]any{
		"type":       "change",
		"changeType": changeType,
		"data":       data,
	})
}

// checkOrigin validates the request origin against allowed origins.
func (s *WebSocketServer) checkOrigin(r *http.Request) bool {
	if s == nil {
		return false
	}
	// No restrictions configured = allow all (dev mode)
	if len(s.allowedOrigins) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		// No origin header (same-origin request or non-browser client)
		return true
	}
	for _, allowed := range s.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (s *WebSocketServer) handleMessage(ctx context.Context, sess *session, message []byte) []byte {
	if s == nil || s.server == nil {
		return nil
	}
	var req request
	if err := json.Unmarshal(message, &req); err != nil {
		resp := response{
			OK:      false,
			Error:   "bad_json",
			Message: err.Error(),
		}
		out, _ := json.Marshal(resp)
		return out
	}
	resp := s.server.handleRequest(ctx, sess, req)
	out, err := json.Marshal(resp)
	if err != nil {
		return nil
	}
	return out
}
