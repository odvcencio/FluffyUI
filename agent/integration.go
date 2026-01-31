//go:build !js

// Package agent provides AI-friendly interaction with FluffyUI applications.
// This file contains integration helpers for using the agent server.
//
// The agent server now operates in real-time mode by default, providing:
//   - Live UI change notifications via event streaming
//   - Bidirectional WebSocket communication
//   - Async wait operations for UI conditions
//   - Event subscription system
//
// Basic usage:
//
//	server, err := agent.EnableFromEnv(app)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if server != nil {
//	    defer server.Stop()
//	}
//	app.Run(ctx)
package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
)

// EnableFromEnv enables the agent server from environment variables.
// This is the primary entry point for agent integration.
//
// The server operates in real-time mode by default, streaming UI events
// to connected clients.
//
// Environment variables:
//   - FLUFFYUI_AGENT: Server address (e.g., "unix:/tmp/agent.sock" or "tcp::8716")
//   - FLUFFYUI_AGENT_WS: WebSocket server address (e.g., ":8765")
//   - FLUFFYUI_AGENT_TOKEN: Optional authentication token
//   - FLUFFYUI_AGENT_ALLOW_TEXT: Set to "1" or "true" to allow text capture
//   - FLUFFYUI_AGENT_MAX_SESSIONS: Maximum concurrent sessions (default: 100)
//   - FLUFFYUI_AGENT_RATE_LIMIT: Requests per second limit (default: 1000)
//   - FLUFFYUI_AGENT_ENABLE_HEALTH: Set to "0" or "false" to disable health checks
//
// Returns nil if FLUFFYUI_AGENT is not set or is set to "0" or "false".
func EnableFromEnv(app *runtime.App) (*RealTimeServer, error) {
	if app == nil {
		return nil, errors.New("app is nil")
	}

	addr := strings.TrimSpace(os.Getenv("FLUFFYUI_AGENT"))
	if addr == "" || addr == "0" || strings.EqualFold(addr, "false") || strings.EqualFold(addr, "off") {
		return nil, nil
	}

	opts := DefaultEnhancedServerOptions()
	opts.Addr = addr
	opts.App = app
	opts.Token = strings.TrimSpace(os.Getenv("FLUFFYUI_AGENT_TOKEN"))
	opts.AllowText = envBool("FLUFFYUI_AGENT_ALLOW_TEXT")
	opts.EnableHealthCheck = !envBool("FLUFFYUI_AGENT_DISABLE_HEALTH")

	// Parse pool limits from env
	if maxSessions := envInt("FLUFFYUI_AGENT_MAX_SESSIONS"); maxSessions > 0 {
		opts.SessionPoolLimits.MaxSessions = maxSessions
	}
	if rateLimit := envInt("FLUFFYUI_AGENT_RATE_LIMIT"); rateLimit > 0 {
		opts.SessionPoolLimits.GlobalRateLimit = rateLimit
	}

	server, err := NewRealTimeServer(opts)
	if err != nil {
		return nil, err
	}

	if err := server.Start(); err != nil {
		return nil, err
	}

	// Start WebSocket server if configured
	if wsAddr := strings.TrimSpace(os.Getenv("FLUFFYUI_AGENT_WS")); wsAddr != "" {
		go func() {
			wsOpts := RealTimeWSOptions{
				EnhancedServerOptions: opts,
			}
			wsServer, err := NewRealTimeWebSocketServer(wsOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "agent: failed to create WebSocket server: %v\n", err)
				return
			}
			if err := wsServer.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "agent: failed to start WebSocket server: %v\n", err)
				return
			}

			http.Handle("/agent", wsServer)
			fmt.Printf("agent: WebSocket server listening on %s/agent\n", wsAddr)
			if err := http.ListenAndServe(wsAddr, nil); err != nil {
				fmt.Fprintf(os.Stderr, "agent: WebSocket server error: %v\n", err)
			}
		}()
	}

	return server, nil
}

// EnableEnhancedServerFromEnv is an alias for EnableFromEnv for backward compatibility.
// Deprecated: Use EnableFromEnv instead.
func EnableEnhancedServerFromEnv(app *runtime.App) (*RealTimeServer, error) {
	return EnableFromEnv(app)
}

// EnableServerFromEnv is an alias for EnableFromEnv for backward compatibility.
// Deprecated: Use EnableFromEnv instead.
func EnableServerFromEnv(app *runtime.App) (*RealTimeServer, error) {
	return EnableFromEnv(app)
}

// RunWithAgent runs the app with an agent server enabled from environment variables.
// The agent server is automatically cleaned up when the app exits.
//
// This uses real-time mode by default.
func RunWithAgent(app *runtime.App, ctx context.Context) error {
	if app == nil {
		return errors.New("app is nil")
	}

	server, err := EnableFromEnv(app)
	if err != nil {
		return fmt.Errorf("failed to start agent server: %w", err)
	}

	if server != nil {
		defer server.Stop()
	}

	return app.Run(ctx)
}

// RunWithRealTimeAgent is an alias for RunWithAgent.
// Deprecated: Use RunWithAgent instead.
func RunWithRealTimeAgent(app *runtime.App, ctx context.Context) error {
	return RunWithAgent(app, ctx)
}

// ServerConfig provides a fluent API for configuring the agent server
type ServerConfig struct {
	addr            string
	wsAddr          string
	token           string
	allowText       bool
	testMode        bool
	maxSessions     int
	maxConns        int
	requestTimeout  time.Duration
	enableHealth    bool
	backgroundMode  bool
	eventFilters    EventFilters
	allowedOrigins  []string
}

// NewConfig creates a new agent configuration
func NewConfig() *ServerConfig {
	return &ServerConfig{
		maxSessions:    100,
		maxConns:       0, // unlimited
		requestTimeout: 30 * time.Second,
		enableHealth:   true,
		eventFilters:   DefaultEventFilters(),
	}
}

// WithAddress sets the server address
func (c *ServerConfig) WithAddress(addr string) *ServerConfig {
	c.addr = addr
	return c
}

// WithWebSocketAddress sets the WebSocket server address
func (c *ServerConfig) WithWebSocketAddress(addr string) *ServerConfig {
	c.wsAddr = addr
	return c
}

// WithToken sets the authentication token
func (c *ServerConfig) WithToken(token string) *ServerConfig {
	c.token = token
	return c
}

// WithTextAccess enables text capture
func (c *ServerConfig) WithTextAccess() *ServerConfig {
	c.allowText = true
	return c
}

// WithTestMode enables test mode
func (c *ServerConfig) WithTestMode() *ServerConfig {
	c.testMode = true
	return c
}

// WithMaxSessions sets the maximum number of sessions
func (c *ServerConfig) WithMaxSessions(n int) *ServerConfig {
	c.maxSessions = n
	return c
}

// WithMaxConnections sets the maximum number of connections
func (c *ServerConfig) WithMaxConnections(n int) *ServerConfig {
	c.maxConns = n
	return c
}

// WithRequestTimeout sets the request timeout
func (c *ServerConfig) WithRequestTimeout(d time.Duration) *ServerConfig {
	c.requestTimeout = d
	return c
}

// WithHealthChecks enables health monitoring
func (c *ServerConfig) WithHealthChecks() *ServerConfig {
	c.enableHealth = true
	return c
}

// WithoutHealthChecks disables health monitoring
func (c *ServerConfig) WithoutHealthChecks() *ServerConfig {
	c.enableHealth = false
	return c
}

// WithBackgroundMode enables background processing mode
func (c *ServerConfig) WithBackgroundMode() *ServerConfig {
	c.backgroundMode = true
	return c
}

// WithEventFilters sets the event filters for real-time notifications
func (c *ServerConfig) WithEventFilters(filters EventFilters) *ServerConfig {
	c.eventFilters = filters
	return c
}

// WithAllowedOrigins sets allowed origins for WebSocket connections
func (c *ServerConfig) WithAllowedOrigins(origins ...string) *ServerConfig {
	c.allowedOrigins = origins
	return c
}

// Build creates a RealTimeServer from the configuration
func (c *ServerConfig) Build(app *runtime.App) (*RealTimeServer, error) {
	if app == nil {
		return nil, errors.New("app is nil")
	}
	if c.addr == "" {
		return nil, errors.New("address is required")
	}

	opts := DefaultEnhancedServerOptions()
	opts.Addr = c.addr
	opts.App = app
	opts.Token = c.token
	opts.AllowText = c.allowText
	opts.TestMode = c.testMode
	opts.MaxConnections = c.maxConns
	opts.RequestTimeout = c.requestTimeout
	opts.EnableHealthCheck = c.enableHealth
	opts.SessionPoolLimits.MaxSessions = c.maxSessions

	if c.backgroundMode {
		opts.SessionLimits = BackgroundSessionLimits()
	}

	return NewRealTimeServer(opts)
}

// BuildWebSocket creates a RealTimeWebSocketServer from the configuration
func (c *ServerConfig) BuildWebSocket(app *runtime.App) (*RealTimeWebSocketServer, error) {
	if app == nil {
		return nil, errors.New("app is nil")
	}
	if c.addr == "" {
		return nil, errors.New("address is required")
	}

	opts := RealTimeWSOptions{
		EnhancedServerOptions: EnhancedServerOptions{
			Addr:              c.addr,
			App:               app,
			Token:             c.token,
			AllowText:         c.allowText,
			TestMode:          c.testMode,
			MaxConnections:    c.maxConns,
			RequestTimeout:    c.requestTimeout,
			EnableHealthCheck: c.enableHealth,
			SessionPoolLimits: PoolLimits{
				MaxSessions: c.maxSessions,
			},
		},
		AllowedOrigins: c.allowedOrigins,
	}

	return NewRealTimeWebSocketServer(opts)
}

// AgentConfig is an alias for ServerConfig for backward compatibility.
// Deprecated: Use ServerConfig instead.
type AgentConfig = ServerConfig

// NewAgentConfig is an alias for NewConfig for backward compatibility.
// Deprecated: Use NewConfig instead.
func NewAgentConfig() *ServerConfig {
	return NewConfig()
}

// RealTimeConfig is an alias for ServerConfig.
// Deprecated: Use ServerConfig instead.
type RealTimeConfig = ServerConfig

// NewRealTimeConfig is an alias for NewConfig.
// Deprecated: Use NewConfig instead.
func NewRealTimeConfig() *ServerConfig {
	return NewConfig()
}

// envBool reads a boolean from environment
func envBool(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// envInt reads an integer from environment
func envInt(key string) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0
	}
	var n int
	fmt.Sscanf(value, "%d", &n)
	return n
}
