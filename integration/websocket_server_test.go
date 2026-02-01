package integration

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func TestAgentWebSocketServerIntegration(t *testing.T) {
	be := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              widgets.NewLabel("Hello"),
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})
	runApp(t, app)

	wsServer, err := agent.NewWebSocketServerWithOptions(agent.WebSocketOptions{
		ServerOptions: agent.ServerOptions{
			App:       app,
			Token:     "secret",
			AllowText: false,
		},
		AllowedOrigins: []string{"*"},
	})
	if err != nil {
		t.Fatalf("new websocket server: %v", err)
	}

	httpServer := httptest.NewServer(wsServer)
	t.Cleanup(httpServer.Close)

	url := "ws" + strings.TrimPrefix(httpServer.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	send := func(t *testing.T, req serverRequest) serverResponse {
		t.Helper()
		if err := conn.SetWriteDeadline(time.Now().Add(appTimeout)); err != nil {
			t.Fatalf("write deadline: %v", err)
		}
		if err := conn.WriteJSON(req); err != nil {
			t.Fatalf("write json: %v", err)
		}
		if err := conn.SetReadDeadline(time.Now().Add(appTimeout)); err != nil {
			t.Fatalf("read deadline: %v", err)
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read message: %v", err)
		}
		var resp serverResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return resp
	}

	resp := send(t, serverRequest{ID: 1, Type: "ping"})
	if resp.Error != "unauthorized" {
		t.Fatalf("expected unauthorized, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 2, Type: "hello", Token: "secret"})
	if !resp.OK {
		t.Fatalf("expected hello ok, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 3, Type: "snapshot", IncludeText: true})
	if resp.Error != "text_disabled" {
		t.Fatalf("expected text_disabled, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 4, Type: "snapshot"})
	if !resp.OK || resp.Snapshot == nil {
		t.Fatalf("expected snapshot, got %+v", resp)
	}
}
