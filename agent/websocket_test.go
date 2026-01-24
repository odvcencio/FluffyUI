package agent

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebSocketServerPing(t *testing.T) {
	wsServer, err := NewWebSocketServer(ServerOptions{Agent: New(Config{})})
	if err != nil {
		t.Fatalf("NewWebSocketServer error: %v", err)
	}
	server := httptest.NewServer(wsServer)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	req := request{ID: 1, Type: "ping"}
	data, _ := json.Marshal(req)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("write error: %v", err)
	}
	_, respData, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	var resp response
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.OK || resp.ID != 1 {
		t.Fatalf("response = %#v, want ok ping", resp)
	}
}

func TestWebSocketServerAuth(t *testing.T) {
	wsServer, err := NewWebSocketServer(ServerOptions{Agent: New(Config{}), Token: "secret"})
	if err != nil {
		t.Fatalf("NewWebSocketServer error: %v", err)
	}
	server := httptest.NewServer(wsServer)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	req := request{ID: 1, Type: "ping"}
	data, _ := json.Marshal(req)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("write error: %v", err)
	}
	_, respData, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	var resp response
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.OK || resp.Error != "unauthorized" {
		t.Fatalf("response = %#v, want unauthorized", resp)
	}

	hello := request{ID: 2, Type: "hello", Token: "secret"}
	helloData, _ := json.Marshal(hello)
	if err := conn.WriteMessage(websocket.TextMessage, helloData); err != nil {
		t.Fatalf("write error: %v", err)
	}
	_, respData, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if err := json.Unmarshal(respData, &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !resp.OK || resp.ID != 2 {
		t.Fatalf("response = %#v, want ok hello", resp)
	}
}

func wsURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http")
}
