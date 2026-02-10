package main

import (
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

var proxyUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// proxyWebSocket relays WebSocket frames between the browser client
// and the user's compiled FluffyUI binary running with web backend.
func proxyWebSocket(w http.ResponseWriter, r *http.Request, targetURL string) error {
	// Connect to user binary's WebSocket
	backConn, _, err := websocket.DefaultDialer.Dial(targetURL, nil)
	if err != nil {
		http.Error(w, "backend unavailable", http.StatusBadGateway)
		return err
	}
	defer backConn.Close()

	// Upgrade browser connection
	frontConn, err := proxyUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer frontConn.Close()

	// Bidirectional relay
	errc := make(chan error, 2)
	go func() { errc <- relayWS(frontConn, backConn) }()
	go func() { errc <- relayWS(backConn, frontConn) }()
	<-errc
	return nil
}

// relayWS copies WebSocket messages from src to dst until an error occurs.
func relayWS(dst, src *websocket.Conn) error {
	for {
		mt, r, err := src.NextReader()
		if err != nil {
			return err
		}
		w, err := dst.NextWriter(mt)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, r); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
	}
}
