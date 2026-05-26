package integration

import (
	"encoding/json"
	"net"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"m31labs.dev/fluffyui/agent"
	"m31labs.dev/fluffyui/backend/sim"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/widgets"
)

type wsMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type wsResponse struct {
	Type  string          `json:"type"`
	ID    string          `json:"id,omitempty"`
	OK    bool            `json:"ok,omitempty"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

func TestAgentRealTimeWebSocketIntegration(t *testing.T) {
	be := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              widgets.NewLabel("Hello"),
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})
	runApp(t, app)

	socketPath := filepath.Join(t.TempDir(), "realtime.sock")
	wsServer, err := agent.NewRealTimeWebSocketServer(agent.RealTimeWSOptions{
		EnhancedServerOptions: agent.EnhancedServerOptions{
			Addr:              "unix:" + socketPath,
			App:               app,
			AllowText:         false,
			EnableHealthCheck: false,
		},
		AllowedOrigins: []string{"*"},
	})
	if err != nil {
		t.Fatalf("new realtime websocket server: %v", err)
	}
	if err := wsServer.Start(); err != nil {
		t.Fatalf("start realtime server: %v", err)
	}
	t.Cleanup(func() {
		_ = wsServer.Stop()
	})

	httpServer := httptest.NewServer(wsServer)
	t.Cleanup(httpServer.Close)

	url := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial realtime websocket: %v", err)
	}
	defer conn.Close()

	readResponse := func(t *testing.T, match func(wsResponse) bool) wsResponse {
		t.Helper()
		deadline := time.Now().Add(appTimeout)
		for time.Now().Before(deadline) {
			if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
				t.Fatalf("read deadline: %v", err)
			}
			var resp wsResponse
			if err := conn.ReadJSON(&resp); err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				t.Fatalf("read json: %v", err)
			}
			if match(resp) {
				return resp
			}
		}
		t.Fatalf("timeout waiting for websocket response")
		return wsResponse{}
	}

	if err := conn.WriteJSON(wsMessage{Type: "ping", ID: "ping-1"}); err != nil {
		t.Fatalf("send ping: %v", err)
	}
	resp := readResponse(t, func(r wsResponse) bool {
		return r.Type == "pong" && r.ID == "ping-1"
	})
	if !resp.OK {
		t.Fatalf("expected pong ok, got %+v", resp)
	}

	if err := conn.WriteJSON(wsMessage{Type: "snapshot", ID: "snap-1"}); err != nil {
		t.Fatalf("send snapshot: %v", err)
	}
	resp = readResponse(t, func(r wsResponse) bool {
		return r.Type == "snapshot" && r.ID == "snap-1"
	})
	if !resp.OK || len(resp.Data) == 0 {
		t.Fatalf("expected snapshot data, got %+v", resp)
	}
}
