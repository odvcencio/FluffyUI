package mcp

import (
	"context"
	"encoding/json"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
)

func newThemeTestServer(t *testing.T) *Server {
	t.Helper()
	app := runtime.NewApp(runtime.AppConfig{
		Backend: sim.New(20, 6),
	})
	s := &Server{
		app:       app,
		opts:      runtime.MCPOptions{},
		sessions:  make(map[string]*sessionState),
		mcpServer: mcpserver.NewMCPServer("test", "dev"),
	}
	registerTools(s)
	return s
}

func TestThemeListThemes(t *testing.T) {
	s := newThemeTestServer(t)
	result, err := s.handleListThemes(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("list_themes error: %v", err)
	}
	if result.IsError {
		t.Fatalf("list_themes returned error result")
	}
	sc, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result.StructuredContent)
	}
	data, ok := sc["data"]
	if !ok {
		t.Fatal("expected data in structured content")
	}
	themes, ok := data.([]string)
	if !ok {
		t.Fatalf("expected []string data, got %T", data)
	}
	if len(themes) != 4 {
		t.Fatalf("expected 4 themes, got %d", len(themes))
	}
	expected := map[string]bool{"dark": true, "light": true, "monokai": true, "nord": true}
	for _, name := range themes {
		if !expected[name] {
			t.Fatalf("unexpected theme name: %s", name)
		}
	}
}

func TestThemeSetTheme(t *testing.T) {
	s := newThemeTestServer(t)
	ctx := context.Background()

	for _, name := range []string{"dark", "light", "monokai", "nord"} {
		args, _ := json.Marshal(themeArgs{Name: name})
		req := mcp.CallToolRequest{}
		req.Params.Arguments = make(map[string]any)
		_ = json.Unmarshal(args, &req.Params.Arguments)
		result, err := s.handleSetTheme(ctx, req)
		if err != nil {
			t.Fatalf("set_theme(%s) error: %v", name, err)
		}
		if result.IsError {
			t.Fatalf("set_theme(%s) returned error", name)
		}
	}
}

func TestThemeSetThemeUnknown(t *testing.T) {
	s := newThemeTestServer(t)
	ctx := context.Background()

	args, _ := json.Marshal(themeArgs{Name: "nonexistent"})
	req := mcp.CallToolRequest{}
	req.Params.Arguments = make(map[string]any)
	_ = json.Unmarshal(args, &req.Params.Arguments)

	result, err := s.handleSetTheme(ctx, req)
	if err != nil {
		t.Fatalf("set_theme(unknown) should not return go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("set_theme(unknown) should return tool error")
	}
}
