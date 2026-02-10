package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestProxyWebSocket(t *testing.T) {
	// Start a mock "user binary" WebSocket server (echo server)
	echoUpgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := echoUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}))
	defer mockBackend.Close()

	targetURL := "ws" + strings.TrimPrefix(mockBackend.URL, "http") + "/ws"

	// Start a proxy server that relays to the mock backend
	proxySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = proxyWebSocket(w, r, targetURL)
	}))
	defer proxySrv.Close()

	// Connect to the proxy as a browser client
	proxyURL := "ws" + strings.TrimPrefix(proxySrv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(proxyURL, nil)
	if err != nil {
		t.Fatal("dial proxy:", err)
	}
	defer conn.Close()

	// Send a message through the proxy
	msg := []byte("hello fluffyui")
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatal("write:", err)
	}

	// Read the echoed response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	mt, resp, err := conn.ReadMessage()
	if err != nil {
		t.Fatal("read:", err)
	}
	if mt != websocket.TextMessage {
		t.Fatalf("expected text message, got %d", mt)
	}
	if string(resp) != string(msg) {
		t.Fatalf("expected %q, got %q", msg, resp)
	}
}

func TestProxyWebSocket_BackendUnavailable(t *testing.T) {
	// Try to proxy to a non-existent backend
	proxySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = proxyWebSocket(w, r, "ws://127.0.0.1:1/ws")
	}))
	defer proxySrv.Close()

	// Make a regular HTTP request (not WebSocket) — should get 502
	resp, err := http.Get(proxySrv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", resp.StatusCode)
	}
}

func TestRelayWS_BinaryMessage(t *testing.T) {
	// Test binary message relay
	echoUpgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := echoUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}))
	defer mockBackend.Close()

	targetURL := "ws" + strings.TrimPrefix(mockBackend.URL, "http")

	proxySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = proxyWebSocket(w, r, targetURL)
	}))
	defer proxySrv.Close()

	proxyURL := "ws" + strings.TrimPrefix(proxySrv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(proxyURL, nil)
	if err != nil {
		t.Fatal("dial:", err)
	}
	defer conn.Close()

	// Send binary ANSI data (simulating terminal output)
	data := []byte("\x1b[2J\x1b[HHello World")
	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		t.Fatal("write:", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	mt, resp, err := conn.ReadMessage()
	if err != nil {
		t.Fatal("read:", err)
	}
	if mt != websocket.BinaryMessage {
		t.Fatalf("expected binary message, got %d", mt)
	}
	if string(resp) != string(data) {
		t.Fatalf("binary data mismatch")
	}
}
