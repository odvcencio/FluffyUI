package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func TestWebSocketServerServeAndBroadcast(t *testing.T) {
	be := sim.New(20, 6)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              widgets.NewLabel("Hello"),
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})
	runAppForTest(t, app)

	opts := WebSocketOptions{
		ServerOptions: ServerOptions{
			App:       app,
			Token:     "secret",
			AllowText: false,
		},
		AllowedOrigins: []string{"https://example.com"},
	}
	wsServer, err := NewWebSocketServerWithOptions(opts)
	if err != nil {
		t.Fatalf("new websocket server: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example", nil)
	req.Header.Set("Origin", "https://example.com")
	if !wsServer.checkOrigin(req) {
		t.Fatal("expected origin allowed")
	}
	req.Header.Set("Origin", "https://evil.example")
	if wsServer.checkOrigin(req) {
		t.Fatal("expected origin rejected")
	}
	req.Header.Del("Origin")
	if !wsServer.checkOrigin(req) {
		t.Fatal("expected empty origin allowed")
	}

	srv := httptest.NewServer(wsServer)
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, http.Header{"Origin": []string{"https://example.com"}})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	send := func(t *testing.T, req request) response {
		t.Helper()
		if err := conn.WriteJSON(req); err != nil {
			t.Fatalf("write json: %v", err)
		}
		var resp response
		if err := conn.ReadJSON(&resp); err != nil {
			t.Fatalf("read json: %v", err)
		}
		return resp
	}

	resp := send(t, request{ID: 1, Type: "ping"})
	if resp.Error != "unauthorized" {
		t.Fatalf("expected unauthorized, got %+v", resp)
	}
	resp = send(t, request{ID: 2, Type: "hello", Token: "wrong"})
	if resp.Error != "unauthorized" {
		t.Fatalf("expected unauthorized, got %+v", resp)
	}
	resp = send(t, request{ID: 3, Type: "hello", Token: "secret"})
	if !resp.OK || resp.Capabilities == nil {
		t.Fatalf("expected capabilities, got %+v", resp)
	}
	resp = send(t, request{ID: 4, Type: "snapshot", IncludeText: true})
	if resp.Error != "text_disabled" {
		t.Fatalf("expected text_disabled, got %+v", resp)
	}
	resp = send(t, request{ID: 5, Type: "snapshot"})
	if !resp.OK || resp.Snapshot == nil {
		t.Fatalf("expected snapshot, got %+v", resp)
	}
	resp = send(t, request{ID: 6, Type: "text", Text: ""})
	if resp.Error != "missing_text" {
		t.Fatalf("expected missing_text, got %+v", resp)
	}
	resp = send(t, request{ID: 7, Type: "mouse", Button: "bad", Action: "press"})
	if resp.Error != "invalid_mouse_button" {
		t.Fatalf("expected invalid_mouse_button, got %+v", resp)
	}
	resp = send(t, request{ID: 8, Type: "unknown"})
	if resp.Error != "unknown_type" {
		t.Fatalf("expected unknown_type, got %+v", resp)
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte("{bad_json}")); err != nil {
		t.Fatalf("write bad json: %v", err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read bad json response: %v", err)
	}
	var badResp response
	if err := json.Unmarshal(msg, &badResp); err != nil {
		t.Fatalf("unmarshal bad json response: %v", err)
	}
	if badResp.Error != "bad_json" {
		t.Fatalf("expected bad_json, got %+v", badResp)
	}

	wsServer.Broadcast(map[string]any{"type": "broadcast", "ok": true})
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read broadcast: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal broadcast: %v", err)
	}
	if payload["type"] != "broadcast" {
		t.Fatalf("expected broadcast payload, got %v", payload)
	}

	wsServer.NotifyChange("widgets", map[string]string{"status": "ok"})
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, raw, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read notify: %v", err)
	}
	payload = map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal notify: %v", err)
	}
	if payload["type"] != "change" {
		t.Fatalf("expected change payload, got %v", payload)
	}
}

func TestWebSocketServerOriginDefaults(t *testing.T) {
	wsServer, err := NewWebSocketServerWithOptions(WebSocketOptions{
		ServerOptions: ServerOptions{
			Agent: New(Config{Sim: sim.New(5, 3)}),
		},
	})
	if err != nil {
		t.Fatalf("new websocket server: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example", nil)
	req.Header.Set("Origin", "https://any.example")
	if wsServer.checkOrigin(req) {
		t.Fatal("expected default origin reject for browser origin")
	}
	req.Header.Del("Origin")
	if !wsServer.checkOrigin(req) {
		t.Fatal("expected default origin allow for non-browser client")
	}

	if wsServer.handleMessage(context.Background(), &session{}, []byte("{bad_json}")) == nil {
		t.Fatal("expected bad json response")
	}
}
