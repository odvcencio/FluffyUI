package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
)

func readWSResponse(t *testing.T, conn *websocket.Conn, wantID string) wsResponse {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn.SetReadDeadline(deadline)
		var resp wsResponse
		if err := conn.ReadJSON(&resp); err != nil {
			t.Fatalf("read json: %v", err)
		}
		if resp.Type == "event" {
			continue
		}
		if wantID == "" || resp.ID == wantID {
			return resp
		}
	}
}

func TestRealTimeWebSocketServerFlows(t *testing.T) {
	input := &testInput{label: "Name"}
	button := &testButton{label: "Submit"}
	root := runtime.VBox(runtime.Fixed(input), runtime.Fixed(button)).WithGap(1)

	be := sim.New(60, 12)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              root,
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})
	runAppForTest(t, app)

	opts := RealTimeWSOptions{
		EnhancedServerOptions: EnhancedServerOptions{
			Addr:              "unix:" + filepath.Join(t.TempDir(), "realtime.sock"),
			App:               app,
			AllowText:         true,
			EnableHealthCheck: false,
		},
		AllowedOrigins: []string{"https://example.com"},
	}

	wsServer, err := NewRealTimeWebSocketServer(opts)
	if err != nil {
		t.Fatalf("new realtime ws server: %v", err)
	}

	wsServer.server.notifier.minInterval = time.Hour
	wsServer.server.notifier.maxInterval = time.Hour

	if err := wsServer.Start(); err != nil {
		t.Fatalf("start realtime ws server: %v", err)
	}
	t.Cleanup(func() { _ = wsServer.Stop() })

	req := httptest.NewRequest("GET", "http://example", nil)
	req.Header.Set("Origin", "https://example.com")
	if !wsServer.checkOrigin(req) {
		t.Fatal("expected origin allowed")
	}
	wsServer.allowedOrigins = []string{"https://allowed.example"}
	req.Header.Set("Origin", "https://evil.example")
	if wsServer.checkOrigin(req) {
		t.Fatal("expected origin denied")
	}
	wsServer.allowedOrigins = nil
	req.Header.Del("Origin")
	if !wsServer.checkOrigin(req) {
		t.Fatal("expected empty origin allowed")
	}
	wsServer.allowedOrigins = []string{"https://example.com"}

	handler, err := RealTimeHandler(opts)
	if err != nil || handler == nil {
		t.Fatalf("RealTimeHandler error: %v", err)
	}

	srv := httptest.NewServer(wsServer)
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": []string{"https://example.com"}})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	// Consume initial event
	var initial wsResponse
	if err := conn.ReadJSON(&initial); err != nil {
		t.Fatalf("read initial event: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for wsServer.ConnectionCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if wsServer.ConnectionCount() == 0 {
		t.Fatal("expected active connection")
	}

	if err := conn.WriteJSON(wsMessage{Type: "ping", ID: "ping"}); err != nil {
		t.Fatalf("write ping: %v", err)
	}
	resp := readWSResponse(t, conn, "ping")
	if resp.Type != "pong" || !resp.OK {
		t.Fatalf("expected pong, got %+v", resp)
	}

	if err := conn.WriteJSON(wsMessage{Type: "snapshot", ID: "snap"}); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
	resp = readWSResponse(t, conn, "snap")
	if resp.Type != "snapshot" || !resp.OK {
		t.Fatalf("expected snapshot, got %+v", resp)
	}

	findPayload, _ := json.Marshal(map[string]string{"by": "label", "value": "Name"})
	if err := conn.WriteJSON(wsMessage{Type: "find", ID: "find", Payload: findPayload}); err != nil {
		t.Fatalf("write find: %v", err)
	}
	resp = readWSResponse(t, conn, "find")
	if resp.Type != "find_result" || !resp.OK {
		t.Fatalf("expected find_result, got %+v", resp)
	}
	findBadPayload, _ := json.Marshal(map[string]string{"by": "unknown", "value": "Name"})
	if err := conn.WriteJSON(wsMessage{Type: "find", ID: "find-bad", Payload: findBadPayload}); err != nil {
		t.Fatalf("write find bad: %v", err)
	}
	resp = readWSResponse(t, conn, "find-bad")
	if resp.Type != "error" || resp.Error == "" {
		t.Fatalf("expected error for find, got %+v", resp)
	}

	actionPayload, _ := json.Marshal(map[string]string{"label": "Name", "action": "focus"})
	if err := conn.WriteJSON(wsMessage{Type: "action", ID: "focus", Payload: actionPayload}); err != nil {
		t.Fatalf("write action: %v", err)
	}
	resp = readWSResponse(t, conn, "focus")
	if resp.Type != "action_complete" || !resp.OK {
		t.Fatalf("expected action_complete, got %+v", resp)
	}
	actionBadPayload, _ := json.Marshal(map[string]string{"label": "Name", "action": "bogus"})
	if err := conn.WriteJSON(wsMessage{Type: "action", ID: "action-bad", Payload: actionBadPayload}); err != nil {
		t.Fatalf("write action bad: %v", err)
	}
	resp = readWSResponse(t, conn, "action-bad")
	if resp.Type != "error" || resp.Error == "" {
		t.Fatalf("expected error for action, got %+v", resp)
	}

	waitPayload, _ := json.Marshal(map[string]any{"for": "widget", "value": "Name", "timeout": 100})
	if err := conn.WriteJSON(wsMessage{Type: "wait", ID: "wait", Payload: waitPayload}); err != nil {
		t.Fatalf("write wait: %v", err)
	}
	resp = readWSResponse(t, conn, "wait")
	if resp.Type != "wait_complete" || !resp.OK {
		t.Fatalf("expected wait_complete, got %+v", resp)
	}
	waitBadPayload, _ := json.Marshal(map[string]any{"for": "value", "value": "Name", "timeout": 10})
	if err := conn.WriteJSON(wsMessage{Type: "wait", ID: "wait-bad", Payload: waitBadPayload}); err != nil {
		t.Fatalf("write wait bad: %v", err)
	}
	resp = readWSResponse(t, conn, "wait-bad")
	if resp.Type != "error" || resp.Error == "" {
		t.Fatalf("expected error for wait, got %+v", resp)
	}

	subPayload, _ := json.Marshal(EventFilters{AllEvents: true})
	if err := conn.WriteJSON(wsMessage{Type: "subscribe", ID: "sub", Payload: subPayload}); err != nil {
		t.Fatalf("write subscribe: %v", err)
	}
	resp = readWSResponse(t, conn, "sub")
	if resp.Type != "subscribed" || !resp.OK {
		t.Fatalf("expected subscribed, got %+v", resp)
	}

	if err := conn.WriteJSON(wsMessage{Type: "unsubscribe", ID: "unsub"}); err != nil {
		t.Fatalf("write unsubscribe: %v", err)
	}
	resp = readWSResponse(t, conn, "unsub")
	if resp.Type != "unsubscribed" || !resp.OK {
		t.Fatalf("expected unsubscribed, got %+v", resp)
	}

	keyPayload, _ := json.Marshal(map[string]string{"key": "enter"})
	if err := conn.WriteJSON(wsMessage{Type: "key", ID: "key", Payload: keyPayload}); err != nil {
		t.Fatalf("write key: %v", err)
	}
	resp = readWSResponse(t, conn, "key")
	if resp.Type != "key_sent" || !resp.OK {
		t.Fatalf("expected key_sent, got %+v", resp)
	}
	keyBadPayload, _ := json.Marshal(map[string]string{"key": "g g"})
	if err := conn.WriteJSON(wsMessage{Type: "key", ID: "key-bad", Payload: keyBadPayload}); err != nil {
		t.Fatalf("write key bad: %v", err)
	}
	resp = readWSResponse(t, conn, "key-bad")
	if resp.Type != "error" || resp.Error == "" {
		t.Fatalf("expected error for key, got %+v", resp)
	}

	textPayload, _ := json.Marshal(map[string]string{"text": "hello"})
	if err := conn.WriteJSON(wsMessage{Type: "text", ID: "text", Payload: textPayload}); err != nil {
		t.Fatalf("write text: %v", err)
	}
	resp = readWSResponse(t, conn, "text")
	if resp.Type != "text_sent" || !resp.OK {
		t.Fatalf("expected text_sent, got %+v", resp)
	}

	mousePayload, _ := json.Marshal(map[string]any{"x": 0, "y": 0, "button": "left", "action": "press"})
	if err := conn.WriteJSON(wsMessage{Type: "mouse", ID: "mouse", Payload: mousePayload}); err != nil {
		t.Fatalf("write mouse: %v", err)
	}
	resp = readWSResponse(t, conn, "mouse")
	if resp.Type != "mouse_sent" || !resp.OK {
		t.Fatalf("expected mouse_sent, got %+v", resp)
	}

	taskPayload, _ := json.Marshal(map[string]string{"action": "submit"})
	if err := conn.WriteJSON(wsMessage{Type: "task", ID: "task-submit", Payload: taskPayload}); err != nil {
		t.Fatalf("write task submit: %v", err)
	}
	resp = readWSResponse(t, conn, "task-submit")
	if resp.Type != "task_submitted" || !resp.OK {
		t.Fatalf("expected task_submitted, got %+v", resp)
	}
	taskBadPayload, _ := json.Marshal(map[string]string{"action": "bogus"})
	if err := conn.WriteJSON(wsMessage{Type: "task", ID: "task-bad", Payload: taskBadPayload}); err != nil {
		t.Fatalf("write task bad: %v", err)
	}
	resp = readWSResponse(t, conn, "task-bad")
	if resp.Type != "error" || resp.Error == "" {
		t.Fatalf("expected error for task bad action, got %+v", resp)
	}

	block := make(chan struct{})
	task, err := wsServer.server.taskManager.Submit("task-id", "task", "", "", func(ctx context.Context, task *BackgroundTask) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-block:
			return nil
		}
	})
	if err != nil {
		t.Fatalf("submit task: %v", err)
	}

	taskStatusPayload, _ := json.Marshal(map[string]string{"action": "status", "task_id": task.ID})
	if err := conn.WriteJSON(wsMessage{Type: "task", ID: "task-status", Payload: taskStatusPayload}); err != nil {
		t.Fatalf("write task status: %v", err)
	}
	resp = readWSResponse(t, conn, "task-status")
	if resp.Type != "task_status" || !resp.OK {
		t.Fatalf("expected task_status, got %+v", resp)
	}

	taskCancelPayload, _ := json.Marshal(map[string]string{"action": "cancel", "task_id": task.ID})
	if err := conn.WriteJSON(wsMessage{Type: "task", ID: "task-cancel", Payload: taskCancelPayload}); err != nil {
		t.Fatalf("write task cancel: %v", err)
	}
	resp = readWSResponse(t, conn, "task-cancel")
	if resp.Type != "task_cancelled" || !resp.OK {
		t.Fatalf("expected task_cancelled, got %+v", resp)
	}

	close(block)

	wsServer.Broadcast(map[string]string{"type": "broadcast", "msg": "ok"})
	deadline = time.Now().Add(time.Second)
	found := false
	for time.Now().Before(deadline) {
		conn.SetReadDeadline(deadline)
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read broadcast: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}
		if payload["type"] == "broadcast" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected broadcast message")
	}
}
