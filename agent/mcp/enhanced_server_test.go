package mcp

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"m31labs.dev/fluffyui/agent"
	"m31labs.dev/fluffyui/backend/sim"
	"m31labs.dev/fluffyui/runtime"
	mcp "m31labs.dev/fluffyui/third_party/mcp-go/mcp"
)

func newEnhancedTestServer(t *testing.T, opts EnhancedOptions) *EnhancedServer {
	t.Helper()
	app := runtime.NewApp(runtime.AppConfig{
		Backend: sim.New(20, 6),
	})
	server, err := NewEnhancedServer(app, runtime.MCPOptions{
		Transport: "unix",
		Addr:      filepath.Join(t.TempDir(), "mcp.sock"),
	}, opts)
	if err != nil {
		t.Fatalf("new enhanced server: %v", err)
	}
	t.Cleanup(func() {
		_ = server.Close()
	})
	return server
}

func TestEnhancedServerStatsAndHealth(t *testing.T) {
	es := newEnhancedTestServer(t, EnhancedOptions{EnableHealthCheck: false})
	stats := es.Stats()
	if stats.Running {
		t.Fatalf("expected server not running")
	}

	if err := es.Start(); err != nil {
		t.Fatalf("start enhanced server: %v", err)
	}
	if !es.Stats().Running {
		t.Fatalf("expected running stats after start")
	}

	es.updateHealth()
	health := es.Health()
	if !health.Healthy {
		t.Fatalf("expected healthy status, got %+v", health)
	}

	req := &agent.Request{}
	es.onRequestStart(req)
	es.onRequestDone(req, 10*time.Millisecond, nil)
	if es.AverageLatency() == 0 {
		t.Fatalf("expected avg latency updated")
	}
}

func TestEnhancedServerRequestHandlerRateLimit(t *testing.T) {
	es := newEnhancedTestServer(t, EnhancedOptions{
		PoolLimits: agent.PoolLimits{
			GlobalRateLimit:  1,
			GlobalBurstLimit: 1,
		},
	})

	if _, err := es.CreateSession("s1", agent.ModeNormal, agent.SessionLimits{MaxRequestsPerSec: 1, BurstLimit: 1}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	handler := func(ctx context.Context) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	}

	if _, err := es.MCPRequestHandler("s1", handler); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if _, err := es.MCPRequestHandler("s1", handler); err == nil {
		t.Fatalf("expected rate limit error")
	} else {
		assertMCPError(t, err, rateErrorCode, "rate limit exceeded")
	}
}

func TestEnhancedServerToolHandlers(t *testing.T) {
	es := newEnhancedTestServer(t, EnhancedOptions{EnableAsyncTools: false})
	es.RegisterEnhancedTools()

	_, err := es.handleSubmitBackgroundTask(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"name": "task"}},
	})
	assertMCPError(t, err, -32007, "async tools disabled")

	es.enhancedOpts.EnableAsyncTools = true

	_, err = es.handleSubmitBackgroundTask(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{}},
	})
	assertMCPError(t, err, mcp.INVALID_PARAMS, "name is required")

	result, err := es.handleSubmitBackgroundTask(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"name": "task", "description": "demo"}},
	})
	if err != nil || result == nil {
		t.Fatalf("expected submit task result, got err=%v", err)
	}

	_, err = es.handleGetTaskStatus(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{}},
	})
	assertMCPError(t, err, mcp.INVALID_PARAMS, "task_id is required")

	_, err = es.handleGetTaskStatus(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"task_id": "missing"}},
	})
	assertMCPError(t, err, mcp.INVALID_PARAMS, "task not found")

	_, err = es.handleCancelTask(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{}},
	})
	assertMCPError(t, err, mcp.INVALID_PARAMS, "task_id is required")
}
