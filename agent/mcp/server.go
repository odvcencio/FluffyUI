package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/runtime"
	"golang.org/x/time/rate"
)

const (
	authErrorCode     = -32001
	limitErrorCode    = -32002
	rateErrorCode     = -32003
	accessErrorCode   = -32004
	defaultServerName = "fluffyui"
	defaultVersion    = "dev"
)

type Server struct {
	app       *runtime.App
	agent     *agent.Agent
	opts      runtime.MCPOptions
	mcpServer *mcpserver.MCPServer
	startedAt time.Time

	ctx    context.Context
	cancel context.CancelFunc

	sseServer   *mcpserver.SSEServer
	stdioServer *mcpserver.StdioServer

	unixListener io.Closer
	unixPath     string

	sessionsMu sync.Mutex
	sessions   map[string]*sessionState

	watcher *resourceWatcher

	closeOnce sync.Once
}

type sessionState struct {
	createdAt time.Time
	lastSeen  time.Time
	authed    bool
	rejected  bool
	tokenHash string
	limiter   *rate.Limiter
}

type mcpError struct {
	code    int
	message string
	data    map[string]any
}

func (e *mcpError) Error() string {
	return e.message
}

func (e *mcpError) MCPCode() int {
	return e.code
}

func (e *mcpError) MCPData() any {
	return e.data
}

type authTokenKey struct{}

func init() {
	runtime.RegisterMCPEnabler(enableMCP)
}

func enableMCP(app *runtime.App, opts runtime.MCPOptions) (io.Closer, error) {
	if app == nil {
		return nil, errors.New("mcp server requires app")
	}
	server, err := NewServer(app, opts)
	if err != nil {
		return nil, err
	}
	if err := server.Start(); err != nil {
		return nil, err
	}
	return server, nil
}

func NewServer(app *runtime.App, opts runtime.MCPOptions) (*Server, error) {
	if app == nil {
		return nil, errors.New("mcp server requires app")
	}
	opts = applyMCPDefaults(opts)

	srv := &Server{
		app:      app,
		agent:    agent.New(agent.Config{App: app}),
		opts:     opts,
		sessions: make(map[string]*sessionState),
	}

	hooks := &mcpserver.Hooks{}
	hooks.AddOnRegisterSession(srv.onRegisterSession)
	hooks.AddOnUnregisterSession(srv.onUnregisterSession)
	hooks.AddOnRequestInitialization(srv.onRequestInitialization)

	srv.mcpServer = mcpserver.NewMCPServer(
		defaultServerName,
		defaultVersion,
		mcpserver.WithResourceCapabilities(true, false),
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithNotificationQueuePolicy(opts.SlowClientPolicy),
		mcpserver.WithHooks(hooks),
		mcpserver.WithInstructions("Fluffy UI MCP server for observing and controlling terminal widgets."),
	)

	registerTools(srv)
	registerResources(srv)

	return srv, nil
}

func (s *Server) Start() error {
	if s == nil {
		return errors.New("mcp server is nil")
	}
	if s.startedAt.IsZero() {
		s.startedAt = time.Now()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel

	if err := s.startTransport(); err != nil {
		cancel()
		return err
	}

	s.watcher = newResourceWatcher(s)
	go s.watcher.Run(ctx)
	return nil
}

func (s *Server) Close() error {
	if s == nil {
		return nil
	}
	var err error
	s.closeOnce.Do(func() {
		if s.mcpServer != nil {
			s.mcpServer.SendNotificationToAllClients("notifications/shutdown", map[string]any{
				"reason": "app_exit",
				"code":   0,
			})
		}
		if s.cancel != nil {
			s.cancel()
		}
		if s.sseServer != nil {
			_ = s.sseServer.Shutdown(context.Background())
		}
		if s.unixListener != nil {
			_ = s.unixListener.Close()
		}
		if s.unixPath != "" {
			_ = os.Remove(s.unixPath)
		}
	})
	return err
}

func (s *Server) onRegisterSession(ctx context.Context, session mcpserver.ClientSession) {
	if session == nil {
		return
	}
	now := time.Now()
	state := &sessionState{
		createdAt: now,
		lastSeen:  now,
		authed:    s.opts.Token == "",
		tokenHash: hashToken(s.opts.Token),
	}
	if s.opts.RateLimit > 0 {
		burst := s.opts.BurstLimit
		if burst <= 0 {
			burst = s.opts.RateLimit * 2
		}
		state.limiter = rate.NewLimiter(rate.Limit(s.opts.RateLimit), burst)
	}

	s.sessionsMu.Lock()
	if s.opts.MaxSessions > 0 && len(s.sessions) >= s.opts.MaxSessions {
		state.rejected = true
	}
	s.sessions[session.SessionID()] = state
	s.sessionsMu.Unlock()
}

func (s *Server) onUnregisterSession(ctx context.Context, session mcpserver.ClientSession) {
	if session == nil {
		return
	}
	s.sessionsMu.Lock()
	delete(s.sessions, session.SessionID())
	s.sessionsMu.Unlock()
}

func (s *Server) onRequestInitialization(ctx context.Context, id any, message any) error {
	state, sessionID := s.sessionState(ctx)
	if state == nil {
		return newMCPError(authErrorCode, "authentication required", nil)
	}
	if state.rejected {
		return newMCPError(limitErrorCode, "too many sessions", nil)
	}
	if s.opts.SessionTimeout > 0 && time.Since(state.lastSeen) > s.opts.SessionTimeout {
		s.mcpServer.UnregisterSession(ctx, sessionID)
		return newMCPError(authErrorCode, "session expired", nil)
	}

	method, params := parseRawRequest(message)
	if s.opts.Token != "" && !state.authed {
		if method != string(mcp.MethodInitialize) {
			return newMCPError(authErrorCode, "authentication required", nil)
		}
		token := authTokenFromContext(ctx)
		if token == "" {
			token = authTokenFromParams(params)
		}
		if token == "" {
			return newMCPError(authErrorCode, "authentication required", nil)
		}
		if token != s.opts.Token {
			return newMCPError(authErrorCode, "authentication failed", nil)
		}
		state.authed = true
	}

	if method == string(mcp.MethodResourcesSubscribe) {
		if uri := paramString(params, "uri"); uri != "" {
			switch uri {
			case "fluffy://screen":
				if !s.textAllowed() {
					return textDeniedError("resources/subscribe")
				}
			case "fluffy://clipboard":
				if !s.clipboardAllowed() {
					return clipboardDeniedError("resources/subscribe")
				}
			}
		}
	}

	if state.limiter != nil && !state.limiter.Allow() {
		retry := rateRetryDelay(state.limiter)
		data := map[string]any{
			"retry_after_ms": retry.Milliseconds(),
			"limit":          s.opts.RateLimit,
			"window_ms":      1000,
		}
		return newMCPError(rateErrorCode, "rate limit exceeded", data)
	}

	state.lastSeen = time.Now()
	return nil
}

func (s *Server) sessionState(ctx context.Context) (*sessionState, string) {
	session := mcpserver.ClientSessionFromContext(ctx)
	if session == nil {
		return nil, ""
	}
	sessionID := session.SessionID()
	s.sessionsMu.Lock()
	state := s.sessions[sessionID]
	s.sessionsMu.Unlock()
	return state, sessionID
}

func (s *Server) textAllowed() bool {
	if s.opts.TestBypassTextGating {
		return true
	}
	return s.opts.AllowText
}

func (s *Server) clipboardAllowed() bool {
	if s.opts.TestBypassClipboardGating {
		return true
	}
	return s.opts.AllowClipboard
}

func (s *Server) sseContext(ctx context.Context, r *http.Request) context.Context {
	if r == nil {
		return ctx
	}
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, authTokenKey{}, token)
}

func authTokenFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if token, ok := ctx.Value(authTokenKey{}).(string); ok {
		return token
	}
	return ""
}

func parseRawRequest(message any) (string, map[string]any) {
	raw, ok := message.(json.RawMessage)
	if !ok {
		return "", nil
	}
	var req struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return "", nil
	}
	params := map[string]any{}
	if len(req.Params) > 0 {
		_ = json.Unmarshal(req.Params, &params)
	}
	return req.Method, params
}

func authTokenFromParams(params map[string]any) string {
	if len(params) == 0 {
		return ""
	}
	authRaw, ok := params["auth"]
	if !ok {
		return ""
	}
	authMap, ok := authRaw.(map[string]any)
	if !ok {
		return ""
	}
	token, _ := authMap["token"].(string)
	return strings.TrimSpace(token)
}

func paramString(params map[string]any, key string) string {
	if len(params) == 0 {
		return ""
	}
	value, ok := params[key]
	if !ok {
		return ""
	}
	if str, ok := value.(string); ok {
		return strings.TrimSpace(str)
	}
	return ""
}

func rateRetryDelay(limiter *rate.Limiter) time.Duration {
	if limiter == nil {
		return time.Second
	}
	res := limiter.Reserve()
	if !res.OK() {
		return time.Second
	}
	delay := res.Delay()
	res.CancelAt(time.Now())
	if delay < 0 {
		return 0
	}
	return delay
}

func textDeniedError(tool string) error {
	data := map[string]any{
		"tool":   tool,
		"reason": "AllowText not enabled",
		"hint":   "Set AllowText: true in MCPOptions or FLUFFY_MCP_ALLOW_TEXT=1",
	}
	return newMCPError(accessErrorCode, "text access denied", data)
}

func clipboardDeniedError(tool string) error {
	data := map[string]any{
		"tool":   tool,
		"reason": "AllowClipboard not enabled",
		"hint":   "Set AllowClipboard: true in MCPOptions or FLUFFY_MCP_ALLOW_CLIPBOARD=1",
	}
	return newMCPError(accessErrorCode, "clipboard access denied", data)
}

func newMCPError(code int, message string, data map[string]any) error {
	return &mcpError{
		code:    code,
		message: message,
		data:    data,
	}
}

func bearerToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func hashToken(token string) string {
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

func applyMCPDefaults(opts runtime.MCPOptions) runtime.MCPOptions {
	if opts.SessionTimeout <= 0 {
		opts.SessionTimeout = 30 * time.Minute
	}
	if opts.MaxSessions <= 0 {
		opts.MaxSessions = 10
	}
	if opts.RateLimit < 0 {
		opts.RateLimit = 0
	}
	if opts.RateLimit == 0 {
		opts.BurstLimit = 0
	} else if opts.BurstLimit <= 0 {
		opts.BurstLimit = opts.RateLimit * 2
	}
	if opts.MaxPendingEvents <= 0 {
		opts.MaxPendingEvents = 100
	}
	if strings.TrimSpace(opts.SlowClientPolicy) == "" {
		opts.SlowClientPolicy = "drop_oldest"
	}
	switch opts.SlowClientPolicy {
	case "drop_oldest", "drop_newest", "disconnect":
	default:
		log.Printf("mcp: invalid slow client policy %q, using drop_oldest", opts.SlowClientPolicy)
		opts.SlowClientPolicy = "drop_oldest"
	}
	return opts
}
