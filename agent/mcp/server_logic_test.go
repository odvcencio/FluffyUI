package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
	mcp "github.com/odvcencio/fluffyui/third_party/mcp-go/mcp"
	mcpserver "github.com/odvcencio/fluffyui/third_party/mcp-go/server"
)

type testSession struct {
	id          string
	initialized bool
	ch          chan mcp.JSONRPCNotification
}

func (t testSession) SessionID() string { return t.id }
func (t testSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return t.ch
}
func (t testSession) Initialize() {}
func (t testSession) Initialized() bool {
	return t.initialized
}

func newTestServer(opts runtime.MCPOptions) *Server {
	return &Server{
		opts:      opts,
		sessions:  make(map[string]*sessionState),
		mcpServer: mcpserver.NewMCPServer("test", "dev"),
	}
}

func TestServerHelperFunctions(t *testing.T) {
	if bearerToken("") != "" {
		t.Fatalf("expected empty bearer token")
	}
	if bearerToken("Token abc") != "" {
		t.Fatalf("expected invalid bearer token")
	}
	if bearerToken("Bearer abc") != "abc" {
		t.Fatalf("expected bearer token to parse")
	}
	if bearerToken("bearer  abc ") != "abc" {
		t.Fatalf("expected bearer token to trim")
	}

	if hashToken("") != "" {
		t.Fatalf("expected empty hash for empty token")
	}
	if hashToken("secret") == "" {
		t.Fatalf("expected hash for token")
	}

	opts := applyMCPDefaults(runtime.MCPOptions{})
	if opts.SessionTimeout <= 0 || opts.MaxSessions <= 0 || opts.MaxPendingEvents <= 0 {
		t.Fatalf("expected defaults applied")
	}
	if opts.SlowClientPolicy == "" {
		t.Fatalf("expected slow client policy default")
	}

	params := map[string]any{"auth": map[string]any{"token": "  tok "}, "uri": "  fluffy://screen  "}
	if authTokenFromParams(params) != "tok" {
		t.Fatalf("expected auth token parsed")
	}
	if paramString(params, "uri") != "fluffy://screen" {
		t.Fatalf("expected param string trimmed")
	}
	if paramString(nil, "uri") != "" {
		t.Fatalf("expected empty param string")
	}

	method, out := parseRawRequest(json.RawMessage(`{"method":"ping","params":{"uri":"fluffy://screen"}}`))
	if method != "ping" || out["uri"].(string) != "fluffy://screen" {
		t.Fatalf("unexpected parseRawRequest result")
	}
	if method, out := parseRawRequest([]byte("not-json")); method != "" || out != nil {
		t.Fatalf("expected parseRawRequest to fail")
	}
}

func TestServerContextAuthHelpers(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	req.Header.Set("Authorization", "Bearer token123")

	srv := newTestServer(runtime.MCPOptions{})
	ctx := srv.sseContext(context.Background(), req)
	if authTokenFromContext(ctx) != "token123" {
		t.Fatalf("expected auth token from context")
	}
}

func TestServerSessionRegistration(t *testing.T) {
	srv := newTestServer(runtime.MCPOptions{MaxSessions: 1})
	s1 := testSession{id: "s1", initialized: true, ch: make(chan mcp.JSONRPCNotification, 1)}
	s2 := testSession{id: "s2", initialized: true, ch: make(chan mcp.JSONRPCNotification, 1)}

	srv.onRegisterSession(context.Background(), s1)
	srv.onRegisterSession(context.Background(), s2)

	state1 := srv.sessions[s1.id]
	state2 := srv.sessions[s2.id]
	if state1 == nil || state2 == nil {
		t.Fatalf("expected session state entries")
	}
	if state1.rejected {
		t.Fatalf("expected first session not rejected")
	}
	if !state2.rejected {
		t.Fatalf("expected second session rejected")
	}

	srv.onUnregisterSession(context.Background(), s1)
	if _, ok := srv.sessions[s1.id]; ok {
		t.Fatalf("expected session removed")
	}
}

func TestServerRequestInitializationAuthFlow(t *testing.T) {
	srv := newTestServer(runtime.MCPOptions{Token: "secret", AllowText: false, AllowClipboard: false})
	session := testSession{id: "s1", initialized: true, ch: make(chan mcp.JSONRPCNotification, 1)}
	srv.onRegisterSession(context.Background(), session)
	ctx := srv.mcpServer.WithContext(context.Background(), session)

	// No auth, non-initialize request should fail.
	err := srv.onRequestInitialization(ctx, nil, json.RawMessage(`{"method":"ping"}`))
	assertMCPError(t, err, authErrorCode, "authentication required")

	// Initialize with wrong token.
	err = srv.onRequestInitialization(ctx, nil, json.RawMessage(`{"method":"initialize","params":{"auth":{"token":"bad"}}}`))
	assertMCPError(t, err, authErrorCode, "authentication failed")

	// Initialize with correct token.
	err = srv.onRequestInitialization(ctx, nil, json.RawMessage(`{"method":"initialize","params":{"auth":{"token":"secret"}}}`))
	if err != nil {
		t.Fatalf("unexpected initialize error: %v", err)
	}
	if !srv.sessions[session.id].authed {
		t.Fatalf("expected session authed")
	}

	// Subscription gating for text.
	err = srv.onRequestInitialization(ctx, nil, json.RawMessage(`{"method":"resources/subscribe","params":{"uri":"fluffy://screen"}}`))
	assertMCPError(t, err, accessErrorCode, "text access denied")

	// Subscription gating for clipboard.
	err = srv.onRequestInitialization(ctx, nil, json.RawMessage(`{"method":"resources/subscribe","params":{"uri":"fluffy://clipboard"}}`))
	assertMCPError(t, err, accessErrorCode, "clipboard access denied")
}

func TestServerRequestInitializationRateLimitAndExpiry(t *testing.T) {
	srv := newTestServer(runtime.MCPOptions{RateLimit: 1, BurstLimit: 1, SessionTimeout: 50 * time.Millisecond})
	session := testSession{id: "s2", initialized: true, ch: make(chan mcp.JSONRPCNotification, 1)}
	srv.onRegisterSession(context.Background(), session)
	ctx := srv.mcpServer.WithContext(context.Background(), session)

	msg := json.RawMessage(`{"method":"ping"}`)
	if err := srv.onRequestInitialization(ctx, nil, msg); err != nil {
		t.Fatalf("unexpected first request error: %v", err)
	}
	if err := srv.onRequestInitialization(ctx, nil, msg); err == nil {
		t.Fatalf("expected rate limit error")
	}

	state := srv.sessions[session.id]
	state.lastSeen = time.Now().Add(-time.Minute)
	if err := srv.onRequestInitialization(ctx, nil, msg); err == nil {
		t.Fatalf("expected session expired error")
	} else if !strings.Contains(err.Error(), "session expired") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertMCPError(t *testing.T, err error, code int, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error")
	}
	var mcpErr *mcpError
	if !errors.As(err, &mcpErr) {
		t.Fatalf("expected mcpError, got %T", err)
	}
	if mcpErr.code != code {
		t.Fatalf("expected code %d, got %d", code, mcpErr.code)
	}
	if !strings.Contains(mcpErr.message, msg) {
		t.Fatalf("expected message %q, got %q", msg, mcpErr.message)
	}
}

// TestTokenComparisonIsConstantTime verifies that token comparison uses
// constant-time comparison (fix for 2.2 timing attack).
// Note: This test cannot truly verify constant-time behavior, but it ensures
// the code path works correctly for both matching and non-matching tokens.
func TestTokenComparisonIsConstantTime(t *testing.T) {
	srv := newTestServer(runtime.MCPOptions{Token: "correct-token"})
	session := testSession{id: "s1", initialized: true, ch: make(chan mcp.JSONRPCNotification, 1)}
	srv.onRegisterSession(context.Background(), session)
	ctx := srv.mcpServer.WithContext(context.Background(), session)

	// Test with wrong token - should fail
	err := srv.onRequestInitialization(ctx, nil, json.RawMessage(`{"method":"initialize","params":{"auth":{"token":"wrong-token"}}}`))
	assertMCPError(t, err, authErrorCode, "authentication failed")

	// Test with correct token - should succeed
	err = srv.onRequestInitialization(ctx, nil, json.RawMessage(`{"method":"initialize","params":{"auth":{"token":"correct-token"}}}`))
	if err != nil {
		t.Fatalf("unexpected error with correct token: %v", err)
	}

	// Test with similar-length wrong token (timing attack scenario)
	srv2 := newTestServer(runtime.MCPOptions{Token: "abcdefghijklmnop"})
	session2 := testSession{id: "s2", initialized: true, ch: make(chan mcp.JSONRPCNotification, 1)}
	srv2.onRegisterSession(context.Background(), session2)
	ctx2 := srv2.mcpServer.WithContext(context.Background(), session2)

	err = srv2.onRequestInitialization(ctx2, nil, json.RawMessage(`{"method":"initialize","params":{"auth":{"token":"abcdefghijklmnox"}}}`))
	assertMCPError(t, err, authErrorCode, "authentication failed")
}

// TestTextGatingOnToolHandlers verifies that text-returning tools are
// properly gated by textAllowed() (fix for 2.3).
func TestTextGatingOnToolHandlers(t *testing.T) {
	// Server with text access disabled
	srv := newTestServer(runtime.MCPOptions{AllowText: false, TestBypassTextGating: false})

	// Verify textAllowed returns false
	if srv.textAllowed() {
		t.Fatal("expected textAllowed to return false")
	}

	// Server with text access enabled via bypass
	srv2 := newTestServer(runtime.MCPOptions{AllowText: false, TestBypassTextGating: true})
	if !srv2.textAllowed() {
		t.Fatal("expected textAllowed to return true with bypass")
	}

	// Server with text access enabled
	srv3 := newTestServer(runtime.MCPOptions{AllowText: true})
	if !srv3.textAllowed() {
		t.Fatal("expected textAllowed to return true when enabled")
	}
}

// TestClipboardGating verifies clipboard gating.
func TestClipboardGating(t *testing.T) {
	srv := newTestServer(runtime.MCPOptions{AllowClipboard: false, TestBypassClipboardGating: false})
	if srv.clipboardAllowed() {
		t.Fatal("expected clipboardAllowed to return false")
	}

	srv2 := newTestServer(runtime.MCPOptions{AllowClipboard: false, TestBypassClipboardGating: true})
	if !srv2.clipboardAllowed() {
		t.Fatal("expected clipboardAllowed to return true with bypass")
	}

	srv3 := newTestServer(runtime.MCPOptions{AllowClipboard: true})
	if !srv3.clipboardAllowed() {
		t.Fatal("expected clipboardAllowed to return true when enabled")
	}
}
