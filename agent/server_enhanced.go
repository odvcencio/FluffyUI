//go:build !js

package agent

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
)

// EnhancedServerOptions configures the enhanced agent interaction server.
type EnhancedServerOptions struct {
	Addr            string
	App             *runtime.App
	Agent           *Agent
	AllowText       bool
	TestMode        bool
	Token           string
	SnapshotTimeout time.Duration

	// Session management
	SessionPoolLimits PoolLimits
	SessionLimits     SessionLimits

	// Request queue
	QueueOptions QueueOptions

	// Background tasks
	MaxBackgroundTasks    int
	MaxTasksPerSession    int

	// Connection handling
	MaxConnections        int           // Max concurrent connections (0 = unlimited)
	ConnectionIdleTimeout time.Duration // Timeout for idle connections
	RequestTimeout        time.Duration // Max time to process a request
	
	// Health and monitoring
	EnableHealthCheck bool
	HealthInterval    time.Duration
}

// DefaultEnhancedServerOptions returns reasonable default options
func DefaultEnhancedServerOptions() EnhancedServerOptions {
	return EnhancedServerOptions{
		SnapshotTimeout:       2 * time.Second,
		SessionPoolLimits:     DefaultPoolLimits(),
		SessionLimits:         DefaultSessionLimits(),
		QueueOptions:          DefaultQueueOptions(),
		MaxBackgroundTasks:    50,
		MaxTasksPerSession:    5,
		ConnectionIdleTimeout: 5 * time.Minute,
		RequestTimeout:        30 * time.Second,
		EnableHealthCheck:     true,
		HealthInterval:        30 * time.Second,
	}
}

// EnhancedServer exposes an out-of-process JSONL API with session management,
// request queuing, and background task support.
type EnhancedServer struct {
	opts EnhancedServerOptions
	agent *Agent

	// Connection management
	listener   net.Listener
	unixPath   string
	connCount  atomic.Int64
	maxConns   int
	connMu     sync.Mutex
	conns      map[net.Conn]context.CancelFunc

	// Core components
	sessionPool *SessionPool
	queue       *RequestQueue
	taskManager *BackgroundTaskManager

	// State
	ctx       context.Context
	cancel    context.CancelFunc
	running   atomic.Bool
	closeOnce sync.Once
	wg        sync.WaitGroup

	// Health
	healthMu   sync.RWMutex
	healthStatus HealthStatus
	lastHealth time.Time
}

// HealthStatus represents the current health of the server
type HealthStatus struct {
	Healthy       bool      `json:"healthy"`
	Message       string    `json:"message,omitempty"`
	ActiveConns   int64     `json:"active_connections"`
	ActiveSessions int      `json:"active_sessions"`
	QueueSize     int       `json:"queue_size"`
	ActiveTasks   int       `json:"active_tasks"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewEnhancedServer validates options and constructs an enhanced server.
func NewEnhancedServer(opts EnhancedServerOptions) (*EnhancedServer, error) {
	if strings.TrimSpace(opts.Addr) == "" {
		return nil, errors.New("agent server address is required")
	}

	// Normalize options
	if opts.SnapshotTimeout <= 0 {
		opts.SnapshotTimeout = 2 * time.Second
	}
	if opts.ConnectionIdleTimeout <= 0 {
		opts.ConnectionIdleTimeout = 5 * time.Minute
	}
	if opts.RequestTimeout <= 0 {
		opts.RequestTimeout = 30 * time.Second
	}
	if opts.HealthInterval <= 0 {
		opts.HealthInterval = 30 * time.Second
	}

	// Create agent if not provided
	agent := opts.Agent
	if agent == nil {
		if opts.App == nil {
			return nil, errors.New("agent server requires App or Agent")
		}
		agent = New(Config{App: opts.App})
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &EnhancedServer{
		opts:        opts,
		agent:       agent,
		maxConns:    opts.MaxConnections,
		conns:       make(map[net.Conn]context.CancelFunc),
		sessionPool: NewSessionPool(opts.SessionPoolLimits),
		queue:       NewRequestQueue(opts.QueueOptions),
		taskManager: NewBackgroundTaskManager(opts.MaxBackgroundTasks, opts.MaxTasksPerSession),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Set up callbacks
	s.queue.SetQueueFullCallback(s.onQueueFull)
	s.queue.SetRequestStartCallback(s.onRequestStart)
	s.queue.SetRequestDoneCallback(s.onRequestDone)

	return s, nil
}

// Start begins accepting connections.
func (s *EnhancedServer) Start() error {
	if s == nil {
		return errors.New("server is nil")
	}

	if !s.running.CompareAndSwap(false, true) {
		return errors.New("server already running")
	}

	ln, unixPath, err := listenAgentAddr(s.opts.Addr)
	if err != nil {
		s.running.Store(false)
		return err
	}

	s.listener = ln
	s.unixPath = unixPath

	// Start session pool housekeeping
	s.sessionPool.Start()

	// Start health checks
	if s.opts.EnableHealthCheck {
		s.wg.Add(1)
		go s.healthCheckLoop()
	}

	// Accept connections
	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop gracefully shuts down the server.
func (s *EnhancedServer) Stop() error {
	if s == nil {
		return nil
	}

	s.closeOnce.Do(func() {
		s.running.Store(false)
		s.cancel()

		// Stop accepting new connections
		if s.listener != nil {
			s.listener.Close()
		}

		// Close all existing connections
		s.connMu.Lock()
		for conn, cancel := range s.conns {
			cancel()
			conn.Close()
		}
		s.conns = make(map[net.Conn]context.CancelFunc)
		s.connMu.Unlock()

		// Clean up unix socket
		if s.unixPath != "" {
			os.Remove(s.unixPath)
		}

		// Stop queue (processes remaining requests)
		s.queue.Stop()

		// Stop session pool
		s.sessionPool.Stop()

		// Wait for all goroutines
		s.wg.Wait()
	})

	return nil
}

// Health returns the current health status
func (s *EnhancedServer) Health() HealthStatus {
	if s == nil {
		return HealthStatus{Healthy: false, Message: "server is nil"}
	}
	s.healthMu.RLock()
	defer s.healthMu.RUnlock()
	return s.healthStatus
}

// Stats returns comprehensive server statistics
func (s *EnhancedServer) Stats() ServerStats {
	if s == nil {
		return ServerStats{}
	}

	poolStats := s.sessionPool.Stats()
	queueStats := s.queue.Stats()

	return ServerStats{
		Running:        s.running.Load(),
		ActiveConns:    s.connCount.Load(),
		SessionStats:   poolStats,
		QueueStats:     queueStats,
		ActiveTasks:    s.taskManager.Count(),
		Health:         s.Health(),
	}
}

// ServerStats contains comprehensive server statistics
type ServerStats struct {
	Running     bool         `json:"running"`
	ActiveConns int64        `json:"active_connections"`
	SessionStats PoolStats   `json:"sessions"`
	QueueStats   QueueStats  `json:"queue"`
	ActiveTasks  int         `json:"active_tasks"`
	Health       HealthStatus `json:"health"`
}

// SubmitBackgroundTask submits a background task
func (s *EnhancedServer) SubmitBackgroundTask(name, description, sessionID string, fn BackgroundTaskFunc) (*BackgroundTask, error) {
	if s == nil {
		return nil, errors.New("server is nil")
	}
	id := generateSessionID()
	return s.taskManager.Submit(id, name, description, sessionID, fn)
}

// acceptLoop accepts incoming connections
func (s *EnhancedServer) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}

		// Check connection limit
		if s.maxConns > 0 && s.connCount.Load() >= int64(s.maxConns) {
			conn.Close()
			continue
		}

		s.connCount.Add(1)
		connCtx, connCancel := context.WithCancel(s.ctx)

		s.connMu.Lock()
		s.conns[conn] = connCancel
		s.connMu.Unlock()

		s.wg.Add(1)
		go s.handleConn(connCtx, conn, connCancel)
	}
}

// handleConn handles a single client connection
func (s *EnhancedServer) handleConn(ctx context.Context, conn net.Conn, cancel context.CancelFunc) {
	defer func() {
		s.connCount.Add(-1)
		s.connMu.Lock()
		delete(s.conns, conn)
		s.connMu.Unlock()
		cancel()
		conn.Close()
		s.wg.Done()
	}()

	// Set idle timeout
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(3 * time.Minute)
	}

	enc := json.NewEncoder(conn)
	enc.SetEscapeHTML(false)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	// Create session
	sessionID := generateSessionID()
	mode := ModeNormal
	limits := s.opts.SessionLimits

	session, err := s.sessionPool.CreateSession(sessionID, mode, limits)
	if err != nil {
		// Send error and close
		_ = enc.Encode(response{
			OK:      false,
			Error:   "session_rejected",
			Message: err.Error(),
		})
		return
	}
	defer s.sessionPool.RemoveSession(sessionID)

	// Cancel tasks for this session when done
	defer s.taskManager.CancelSession(sessionID)

	sess := &serverSession{
		id:      sessionID,
		session: session,
		authed:  s.opts.Token == "",
	}

	// Process requests
	idleTimer := time.NewTimer(s.opts.ConnectionIdleTimeout)
	defer idleTimer.Stop()

	for {
		// Reset idle timer on activity
		if !idleTimer.Stop() {
			select {
			case <-idleTimer.C:
			default:
			}
		}
		idleTimer.Reset(s.opts.ConnectionIdleTimeout)

		// Use a goroutine to handle the scanner to allow for timeout
		done := make(chan bool, 1)
		var line []byte

		go func() {
			if scanner.Scan() {
				line = scanner.Bytes()
				done <- true
			} else {
				done <- false
			}
		}()

		select {
		case <-ctx.Done():
			return
		case <-idleTimer.C:
			return
		case success := <-done:
			if !success {
				return
			}
		}

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

		// Process request with timeout
		resp := s.processRequest(ctx, sess, req)
		if err := enc.Encode(resp); err != nil {
			return
		}
	}
}

// processRequest processes a single request
func (s *EnhancedServer) processRequest(ctx context.Context, sess *serverSession, req request) response {
	// Check session
	session := sess.session
	if session.IsExpired(time.Now()) {
		return response{ID: req.ID, OK: false, Error: "session_expired"}
	}
	if session.IsRejected() {
		return response{ID: req.ID, OK: false, Error: "session_rejected"}
	}

	// Check global rate limit
	if err := s.sessionPool.CheckGlobalRate(); err != nil {
		var rateErr *RateLimitError
		if errors.As(err, &rateErr) {
			return response{
				ID:      req.ID,
				OK:      false,
				Error:   "rate_limited",
				Message: fmt.Sprintf("retry after %v", rateErr.RetryAfter),
			}
		}
		return response{ID: req.ID, OK: false, Error: "rate_limited"}
	}

	// Start request tracking
	if err := session.StartRequest(); err != nil {
		var rateErr *RateLimitError
		if errors.As(err, &rateErr) {
			return response{
				ID:      req.ID,
				OK:      false,
				Error:   "rate_limited",
				Message: fmt.Sprintf("retry after %v", rateErr.RetryAfter),
			}
		}
		return response{ID: req.ID, OK: false, Error: "too_many_requests"}
	}
	defer session.EndRequest(true)

	// Handle request with timeout
	reqCtx, cancel := context.WithTimeout(ctx, s.opts.RequestTimeout)
	defer cancel()

	done := make(chan response, 1)
	go func() {
		done <- s.handleRequest(reqCtx, sess, req)
	}()

	select {
	case <-reqCtx.Done():
		if errors.Is(reqCtx.Err(), context.DeadlineExceeded) {
			return response{ID: req.ID, OK: false, Error: "timeout", Message: "request timed out"}
		}
		return response{ID: req.ID, OK: false, Error: "cancelled"}
	case resp := <-done:
		return resp
	}
}

// handleRequest handles the actual request logic
func (s *EnhancedServer) handleRequest(ctx context.Context, sess *serverSession, req request) response {
	if strings.TrimSpace(req.Type) == "" {
		return response{ID: req.ID, OK: false, Error: "missing_type"}
	}

	if !sess.authed && req.Type != "hello" {
		return response{ID: req.ID, OK: false, Error: "unauthorized", Message: "authentication required"}
	}

	switch req.Type {
	case "hello":
		return s.handleHello(sess, req)
	case "ping":
		return response{ID: req.ID, OK: true}
	case "health":
		return s.handleHealth(req)
	case "stats":
		return s.handleStats(req)
	case "snapshot":
		return s.handleSnapshot(ctx, req)
	case "key":
		return s.handleKey(req)
	case "text":
		return s.handleText(req)
	case "mouse":
		return s.handleMouse(req)
	case "paste":
		return s.handlePaste(req)
	case "resize":
		return s.handleResize(req)
	case "background_task":
		return s.handleBackgroundTask(sess, req)
	case "task_status":
		return s.handleTaskStatus(req)
	case "task_cancel":
		return s.handleTaskCancel(req)
	case "close":
		return response{ID: req.ID, OK: true, Message: "closing"}
	default:
		return response{ID: req.ID, OK: false, Error: "unknown_type"}
	}
}

// handleHello handles authentication
func (s *EnhancedServer) handleHello(sess *serverSession, req request) response {
	if s.opts.Token != "" && req.Token != s.opts.Token {
		return response{ID: req.ID, OK: false, Error: "unauthorized", Message: "invalid token"}
	}
	sess.authed = true
	sess.session.Auth()
	return response{
		ID: req.ID,
		OK: true,
		Capabilities: &Capabilities{
			AllowText: s.opts.AllowText,
			TestMode:  s.opts.TestMode,
		},
	}
}

// handleHealth returns health status
func (s *EnhancedServer) handleHealth(req request) response {
	health := s.Health()
	return response{ID: req.ID, OK: true, Message: encodeJSON(health)}
}

// handleStats returns server statistics
func (s *EnhancedServer) handleStats(req request) response {
	stats := s.Stats()
	return response{ID: req.ID, OK: true, Message: encodeJSON(stats)}
}

// handleSnapshot captures a UI snapshot
func (s *EnhancedServer) handleSnapshot(ctx context.Context, req request) response {
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
}

// handleKey handles key input
func (s *EnhancedServer) handleKey(req request) response {
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
}

// handleText handles text input
func (s *EnhancedServer) handleText(req request) response {
	if strings.TrimSpace(req.Text) == "" {
		return response{ID: req.ID, OK: false, Error: "missing_text"}
	}
	if err := s.agent.SendKeyString(req.Text); err != nil {
		return response{ID: req.ID, OK: false, Error: "send_text_failed", Message: err.Error()}
	}
	return response{ID: req.ID, OK: true}
}

// handleMouse handles mouse input
func (s *EnhancedServer) handleMouse(req request) response {
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
}

// handlePaste handles paste operations
func (s *EnhancedServer) handlePaste(req request) response {
	if err := s.agent.SendPaste(req.Text); err != nil {
		return response{ID: req.ID, OK: false, Error: "send_paste_failed", Message: err.Error()}
	}
	return response{ID: req.ID, OK: true}
}

// handleResize handles resize operations
func (s *EnhancedServer) handleResize(req request) response {
	if req.Width <= 0 || req.Height <= 0 {
		return response{ID: req.ID, OK: false, Error: "invalid_resize"}
	}
	if err := s.agent.SendResize(req.Width, req.Height); err != nil {
		return response{ID: req.ID, OK: false, Error: "send_resize_failed", Message: err.Error()}
	}
	return response{ID: req.ID, OK: true}
}

// handleBackgroundTask submits a background task
func (s *EnhancedServer) handleBackgroundTask(sess *serverSession, req request) response {
	// Task submission via request would need custom parsing
	// For now, just return capability info
	return response{
		ID:      req.ID,
		OK:      true,
		Message: fmt.Sprintf("background tasks available, session %s", sess.id),
	}
}

// handleTaskStatus returns task status
func (s *EnhancedServer) handleTaskStatus(req request) response {
	taskID := req.Text // Use Text field for task ID
	if taskID == "" {
		// Return all tasks
		stats := s.taskManager.Stats()
		return response{ID: req.ID, OK: true, Message: encodeJSON(stats)}
	}

	task := s.taskManager.Get(taskID)
	if task == nil {
		return response{ID: req.ID, OK: false, Error: "task_not_found"}
	}

	return response{ID: req.ID, OK: true, Message: encodeJSON(task.Stats())}
}

// handleTaskCancel cancels a task
func (s *EnhancedServer) handleTaskCancel(req request) response {
	taskID := req.Text
	if taskID == "" {
		return response{ID: req.ID, OK: false, Error: "missing_task_id"}
	}

	if !s.taskManager.Cancel(taskID) {
		return response{ID: req.ID, OK: false, Error: "task_not_found"}
	}

	return response{ID: req.ID, OK: true}
}

// healthCheckLoop performs periodic health checks
func (s *EnhancedServer) healthCheckLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.opts.HealthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateHealth()
		}
	}
}

// updateHealth updates the health status
func (s *EnhancedServer) updateHealth() {
	stats := s.Stats()

	healthy := true
	message := "healthy"

	// Check thresholds
	if stats.QueueStats.CriticalSize > 100 {
		healthy = false
		message = "high critical queue backlog"
	} else if stats.QueueStats.TotalQueued > 500 {
		healthy = false
		message = "high total queue backlog"
	} else if stats.SessionStats.TotalPendingRequests > 200 {
		healthy = false
		message = "high pending request count"
	}

	s.healthMu.Lock()
	s.healthStatus = HealthStatus{
		Healthy:         healthy,
		Message:         message,
		ActiveConns:     stats.ActiveConns,
		ActiveSessions:  stats.SessionStats.TotalSessions,
		QueueSize:       stats.QueueStats.CriticalSize + stats.QueueStats.HighSize + stats.QueueStats.NormalSize + stats.QueueStats.LowSize + stats.QueueStats.BackgroundSize,
		ActiveTasks:     stats.ActiveTasks,
		Timestamp:       time.Now(),
	}
	s.lastHealth = time.Now()
	s.healthMu.Unlock()
}

// Callbacks
func (s *EnhancedServer) onQueueFull(req *Request) {
	// Log or handle queue full condition
}

func (s *EnhancedServer) onRequestStart(req *Request) {
	// Track request start
}

func (s *EnhancedServer) onRequestDone(req *Request, duration time.Duration, err error) {
	// Track request completion
}

// serverSession wraps a Session for the server
type serverSession struct {
	id      string
	session *Session
	authed  bool
}

// encodeJSON helper
func encodeJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b) + "-" + time.Now().Format("20060102150405")
}

// Ensure EnhancedServer implements io.Closer
var _ io.Closer = (*EnhancedServer)(nil)

// Close implements io.Closer
func (s *EnhancedServer) Close() error {
	return s.Stop()
}
