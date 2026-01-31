package runtime

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// MCPOptions configures the MCP server integration.
type MCPOptions struct {
	// Transport
	Transport string // "stdio", "unix", or "sse" (auto-detected if empty)
	Addr      string // Socket path or HTTP address

	// Security
	AllowText      bool   // Include raw screen text in snapshots
	AllowClipboard bool   // Enable clipboard tools
	Token          string // Auth token (optional)

	// Sessions
	SessionTimeout time.Duration // Inactivity timeout (default: 30m)
	MaxSessions    int           // Max concurrent (default: 10)

	// Rate limiting
	RateLimit        int    // Requests per second (0 = unlimited)
	BurstLimit       int    // Burst allowance (default: RateLimit * 2)
	MaxPendingEvents int    // Subscription backlog (default: 100)
	SlowClientPolicy string // "drop_oldest" | "drop_newest" | "disconnect"

	// Behavior
	StrictLabelMatching bool // Error on ambiguous labels

	// Testing only (panics in release builds)
	TestBypassTextGating      bool
	TestBypassClipboardGating bool
}

type mcpEnableFunc func(app *App, opts MCPOptions) (io.Closer, error)

var mcpEnabler mcpEnableFunc

// RegisterMCPEnabler registers the MCP implementation hook.
func RegisterMCPEnabler(fn mcpEnableFunc) {
	mcpEnabler = fn
}

// EnableMCP starts the MCP server for the app.
func (a *App) EnableMCP(opts ...MCPOptions) (io.Closer, error) {
	if a == nil {
		return nil, errors.New("app is nil")
	}
	if a.mcpCloser != nil {
		return a.mcpCloser, nil
	}
	if mcpEnabler == nil {
		return nil, errors.New("mcp support not linked (import agent/mcp)")
	}

	enabledOpts := MCPOptions{}
	if len(opts) > 0 {
		enabledOpts = opts[0]
	}
	normalized, err := normalizeMCPOptions(enabledOpts)
	if err != nil {
		return nil, err
	}
	closer, err := mcpEnabler(a, normalized)
	if err != nil {
		return nil, err
	}
	a.mcpCloser = closer
	return closer, nil
}

func (a *App) enableMCPFromEnv() error {
	if a == nil || a.mcpCloser != nil {
		return nil
	}
	opts, ok, err := mcpOptionsFromEnv()
	if err != nil || !ok {
		return err
	}
	_, err = a.EnableMCP(opts)
	return err
}

func mcpOptionsFromEnv() (MCPOptions, bool, error) {
	value := strings.TrimSpace(os.Getenv("FLUFFY_MCP"))
	if value == "" || value == "0" || strings.EqualFold(value, "false") || strings.EqualFold(value, "off") {
		return MCPOptions{}, false, nil
	}
	opts := MCPOptions{
		AllowText:           envBool("FLUFFY_MCP_ALLOW_TEXT"),
		AllowClipboard:      envBool("FLUFFY_MCP_ALLOW_CLIPBOARD"),
		Token:               strings.TrimSpace(os.Getenv("FLUFFY_MCP_TOKEN")),
		RateLimit:           envInt("FLUFFY_MCP_RATE_LIMIT"),
		BurstLimit:          envInt("FLUFFY_MCP_BURST_LIMIT"),
		MaxSessions:         envInt("FLUFFY_MCP_MAX_SESSIONS"),
		MaxPendingEvents:    envInt("FLUFFY_MCP_MAX_PENDING_EVENTS"),
		SlowClientPolicy:    strings.TrimSpace(os.Getenv("FLUFFY_MCP_SLOW_CLIENT_POLICY")),
		StrictLabelMatching: envBool("FLUFFY_MCP_STRICT_LABELS"),
	}
	if timeout, ok, err := envDuration("FLUFFY_MCP_SESSION_TIMEOUT"); err != nil {
		return MCPOptions{}, false, err
	} else if ok {
		opts.SessionTimeout = timeout
	}
	if envBool("FLUFFY_MCP_TEST_BYPASS") {
		opts.TestBypassTextGating = true
		opts.TestBypassClipboardGating = true
	}

	switch {
	case value == "1" || strings.EqualFold(value, "true") || strings.EqualFold(value, "on"):
		opts.Transport = ""
	case strings.EqualFold(value, "stdio"):
		opts.Transport = "stdio"
	case strings.EqualFold(value, "unix"):
		opts.Transport = "unix"
	case strings.HasPrefix(value, "unix://"):
		parsed, err := url.Parse(value)
		if err != nil {
			return MCPOptions{}, false, fmt.Errorf("invalid FLUFFY_MCP value: %w", err)
		}
		path := parsed.Path
		if path == "" {
			path = parsed.Host
		}
		if path == "" {
			return MCPOptions{}, false, fmt.Errorf("invalid FLUFFY_MCP value: missing unix socket path")
		}
		opts.Transport = "unix"
		opts.Addr = path
	case strings.HasPrefix(value, "sse://") || strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://"):
		parsed, err := url.Parse(value)
		if err != nil {
			return MCPOptions{}, false, fmt.Errorf("invalid FLUFFY_MCP value: %w", err)
		}
		if parsed.Host == "" {
			return MCPOptions{}, false, fmt.Errorf("invalid FLUFFY_MCP value: missing host")
		}
		opts.Transport = "sse"
		opts.Addr = parsed.Host
	default:
		opts.Transport = "sse"
		opts.Addr = value
	}

	return opts, true, nil
}

func envBool(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func envInt(key string) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return num
}

func envDuration(key string) (time.Duration, bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, false, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, false, fmt.Errorf("invalid %s duration: %w", key, err)
	}
	return parsed, true, nil
}

func normalizeMCPOptions(opts MCPOptions) (MCPOptions, error) {
	transport := strings.TrimSpace(opts.Transport)
	if transport == "" {
		transport = defaultTransport()
	}
	transport = strings.ToLower(transport)
	switch transport {
	case "stdio", "sse", "unix":
	default:
		return MCPOptions{}, fmt.Errorf("invalid MCP transport %q", transport)
	}
	opts.Transport = transport

	if transport == "sse" && strings.TrimSpace(opts.Addr) == "" {
		opts.Addr = ":8716"
	}
	if transport == "unix" && strings.TrimSpace(opts.Addr) == "" {
		opts.Addr = defaultUnixSocketPath()
	}
	if err := opts.validateTestFlags(); err != nil {
		return MCPOptions{}, err
	}
	return opts, nil
}

func defaultTransport() string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "unix"
	}
	return "stdio"
}

func defaultUnixSocketPath() string {
	dir := strings.TrimSpace(os.Getenv("XDG_RUNTIME_DIR"))
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, fmt.Sprintf("fluffy-mcp-%d.sock", os.Getpid()))
}
