// Package web provides a backend that renders FluffyUI applications to a web
// browser via WebSocket. This enables TUI apps to be accessed remotely without
// SSH, similar to Textual's textual-serve feature.
//
// The web backend uses xterm.js in the browser for terminal emulation,
// communicating with the Go server via WebSocket. Terminal data flows as
// ANSI escape sequences for maximum compatibility.
//
// Basic usage:
//
//	backend := web.New(web.Config{Addr: ":8080"})
//	app := runtime.NewApp(runtime.AppConfig{Backend: backend})
//	// ... set up app
//	app.Run(context.Background())
//
// The backend serves a terminal UI at http://localhost:8080 by default.
package web

import (
	"time"
)

// RendererType selects the output format for terminal data.
type RendererType int

const (
	// RendererANSI sends ANSI escape sequences (default).
	// Compatible with xterm.js and most terminal emulators.
	RendererANSI RendererType = iota

	// RendererCell sends binary cell data (future enhancement).
	// More compact but requires custom frontend support.
	RendererCell
)

// Config configures the web backend.
type Config struct {
	// Addr is the listen address (default: ":8080").
	// Use ":0" to let the system assign a port.
	Addr string

	// Renderer selects the output format (default: RendererANSI).
	Renderer RendererType

	// Path is the URL path for the terminal UI (default: "/").
	Path string

	// WSPath is the URL path for WebSocket connections (default: "/ws").
	WSPath string

	// Title is the page title (default: "FluffyUI Web Terminal").
	Title string

	// MaxConnections limits concurrent sessions (0 = unlimited).
	MaxConnections int

	// IdleTimeout closes inactive sessions after this duration (0 = never).
	IdleTimeout time.Duration

	// Auth enables basic authentication.
	Auth AuthConfig

	// TLS enables HTTPS/WSS.
	TLS TLSConfig

	// Chromeless hides the header, footer, and status indicators,
	// making the terminal fill the entire browser window.
	Chromeless bool

	// OnConnect is called when a client connects.
	OnConnect func(sessionID string)

	// OnDisconnect is called when a client disconnects.
	OnDisconnect func(sessionID string)
}

// AuthConfig configures authentication.
type AuthConfig struct {
	// Enabled turns on authentication.
	Enabled bool

	// Username for basic auth.
	Username string

	// Password for basic auth.
	Password string

	// Token enables token-based auth via ?token= URL parameter.
	Token string
}

// TLSConfig configures HTTPS/WSS.
type TLSConfig struct {
	// Enabled turns on TLS.
	Enabled bool

	// CertFile is the path to the TLS certificate.
	CertFile string

	// KeyFile is the path to the TLS key.
	KeyFile string
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Addr:     ":8080",
		Renderer: RendererANSI,
		Path:     "/",
		WSPath:   "/ws",
		Title:    "FluffyUI Web Terminal",
	}
}

// fillDefaults populates unset config values with defaults.
func fillDefaults(cfg *Config) {
	defaults := DefaultConfig()
	if cfg.Addr == "" {
		cfg.Addr = defaults.Addr
	}
	if cfg.Path == "" {
		cfg.Path = defaults.Path
	}
	if cfg.WSPath == "" {
		cfg.WSPath = defaults.WSPath
	}
	if cfg.Title == "" {
		cfg.Title = defaults.Title
	}
}
