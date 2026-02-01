package agent

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
)

func TestEnvHelpers(t *testing.T) {
	t.Setenv("FLUFFYUI_AGENT_ALLOW_TEXT", "1")
	if !envBool("FLUFFYUI_AGENT_ALLOW_TEXT") {
		t.Fatal("expected envBool true")
	}
	t.Setenv("FLUFFYUI_AGENT_ALLOW_TEXT", "off")
	if envBool("FLUFFYUI_AGENT_ALLOW_TEXT") {
		t.Fatal("expected envBool false")
	}

	t.Setenv("FLUFFYUI_AGENT_MAX_SESSIONS", "42")
	if envInt("FLUFFYUI_AGENT_MAX_SESSIONS") != 42 {
		t.Fatal("expected envInt to parse")
	}
	t.Setenv("FLUFFYUI_AGENT_MAX_SESSIONS", "")
	if envInt("FLUFFYUI_AGENT_MAX_SESSIONS") != 0 {
		t.Fatal("expected envInt empty to return 0")
	}
}

func TestEnableFromEnvDisabled(t *testing.T) {
	app := runtime.NewApp(runtime.AppConfig{Backend: sim.New(5, 3)})
	runAppForTest(t, app)

	t.Setenv("FLUFFYUI_AGENT", "0")
	server, err := EnableFromEnv(app)
	if err != nil {
		t.Fatalf("enable from env: %v", err)
	}
	if server != nil {
		t.Fatal("expected nil server when disabled")
	}

	if server, err := EnableEnhancedServerFromEnv(app); err != nil || server != nil {
		t.Fatalf("expected nil enhanced server, got %v err=%v", server, err)
	}
	if server, err := EnableServerFromEnv(app); err != nil || server != nil {
		t.Fatalf("expected nil server, got %v err=%v", server, err)
	}
}

func TestEnableFromEnvStartsServer(t *testing.T) {
	app := runtime.NewApp(runtime.AppConfig{Backend: sim.New(5, 3)})
	runAppForTest(t, app)

	t.Setenv("FLUFFYUI_AGENT", "unix:"+filepath.Join(t.TempDir(), "agent.sock"))
	t.Setenv("FLUFFYUI_AGENT_TOKEN", "token")
	t.Setenv("FLUFFYUI_AGENT_ALLOW_TEXT", "true")
	t.Setenv("FLUFFYUI_AGENT_DISABLE_HEALTH", "1")
	t.Setenv("FLUFFYUI_AGENT_MAX_SESSIONS", "2")
	t.Setenv("FLUFFYUI_AGENT_RATE_LIMIT", "7")

	server, err := EnableFromEnv(app)
	if err != nil {
		t.Fatalf("enable from env: %v", err)
	}
	if server == nil {
		t.Fatal("expected server")
	}
	t.Cleanup(func() {
		_ = server.Stop()
	})

	if server.opts.Token != "token" {
		t.Fatalf("token = %q", server.opts.Token)
	}
	if !server.opts.AllowText {
		t.Fatal("expected text access enabled")
	}
	if server.opts.EnableHealthCheck {
		t.Fatal("expected health checks disabled")
	}
	if server.opts.SessionPoolLimits.MaxSessions != 2 {
		t.Fatalf("max sessions = %d", server.opts.SessionPoolLimits.MaxSessions)
	}
	if server.opts.SessionPoolLimits.GlobalRateLimit != 7 {
		t.Fatalf("rate limit = %d", server.opts.SessionPoolLimits.GlobalRateLimit)
	}
}

func TestRunWithAgent(t *testing.T) {
	if err := RunWithAgent(nil, context.Background()); err == nil {
		t.Fatal("expected error for nil app")
	}

	app := runtime.NewApp(runtime.AppConfig{Backend: sim.New(5, 3)})
	t.Setenv("FLUFFYUI_AGENT", "0")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := RunWithAgent(app, ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled error, got %v", err)
	}

	if err := RunWithRealTimeAgent(nil, context.Background()); err == nil {
		t.Fatal("expected error for nil app via RunWithRealTimeAgent")
	}
}

func TestServerConfigBuild(t *testing.T) {
	cfg := NewConfig()
	if cfg.maxSessions != 100 || cfg.maxConns != 0 || !cfg.enableHealth {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}

	addr := "unix:" + filepath.Join(t.TempDir(), "server.sock")
	cfg.WithAddress(addr).
		WithWebSocketAddress(":0").
		WithToken("tok").
		WithTextAccess().
		WithTestMode().
		WithMaxSessions(5).
		WithMaxConnections(2).
		WithRequestTimeout(5 * time.Second).
		WithHealthChecks().
		WithoutHealthChecks().
		WithBackgroundMode().
		WithEventFilters(EventFilters{AllEvents: true}).
		WithAllowedOrigins("https://example.com")

	if cfg.addr != addr || cfg.token != "tok" || !cfg.allowText || !cfg.testMode {
		t.Fatalf("unexpected config state: %+v", cfg)
	}
	if cfg.maxSessions != 5 || cfg.maxConns != 2 || cfg.requestTimeout != 5*time.Second {
		t.Fatalf("unexpected limits: %+v", cfg)
	}
	if cfg.enableHealth {
		t.Fatal("expected health checks disabled")
	}
	if !cfg.backgroundMode || !cfg.eventFilters.AllEvents {
		t.Fatal("expected background mode and all events")
	}
	if len(cfg.allowedOrigins) != 1 || cfg.allowedOrigins[0] != "https://example.com" {
		t.Fatalf("unexpected origins: %v", cfg.allowedOrigins)
	}

	if NewAgentConfig() == nil || NewRealTimeConfig() == nil {
		t.Fatal("expected legacy config helpers")
	}

	app := runtime.NewApp(runtime.AppConfig{Backend: sim.New(5, 3)})
	if _, err := cfg.Build(nil); err == nil {
		t.Fatal("expected build error for nil app")
	}
	if _, err := NewConfig().Build(app); err == nil {
		t.Fatal("expected build error for missing address")
	}

	server, err := cfg.Build(app)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if server.opts.SessionLimits.MaxRequestsPerSec != BackgroundSessionLimits().MaxRequestsPerSec {
		t.Fatalf("expected background session limits, got %+v", server.opts.SessionLimits)
	}

	if _, err := cfg.BuildWebSocket(nil); err == nil {
		t.Fatal("expected websocket build error for nil app")
	}
	if _, err := NewConfig().BuildWebSocket(app); err == nil {
		t.Fatal("expected websocket build error for missing address")
	}

	ws, err := cfg.BuildWebSocket(app)
	if err != nil {
		t.Fatalf("build websocket: %v", err)
	}
	if len(ws.allowedOrigins) != 1 || ws.allowedOrigins[0] != "https://example.com" {
		t.Fatalf("unexpected websocket origins: %v", ws.allowedOrigins)
	}
}
