package mcp

import (
	"bufio"
	"bytes"
	"sync"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/odvcencio/fluffyui/runtime"
)

func TestSocketSessionBasics(t *testing.T) {
	session := newSocketSession(0)
	if session.Initialized() {
		t.Fatalf("expected uninitialized session")
	}
	if session.GetLogLevel() != mcp.LoggingLevelError {
		t.Fatalf("expected default log level")
	}
	called := 0
	session.setCloseFn(func() { called++ })
	_ = session.Close()
	_ = session.Close()
	if called != 1 {
		t.Fatalf("expected close called once")
	}

	session.Initialize()
	if !session.Initialized() {
		t.Fatalf("expected initialized")
	}
	session.SetLogLevel(mcp.LoggingLevelDebug)
	if session.GetLogLevel() != mcp.LoggingLevelDebug {
		t.Fatalf("expected log level debug")
	}

	tools := map[string]mcpserver.ServerTool{"t": {Tool: mcp.Tool{Name: "t"}}}
	session.SetSessionTools(tools)
	if len(session.GetSessionTools()) != 1 {
		t.Fatalf("expected tools")
	}
	resources := map[string]mcpserver.ServerResource{"r": {Resource: mcp.Resource{URI: "u"}}}
	session.SetSessionResources(resources)
	if len(session.GetSessionResources()) != 1 {
		t.Fatalf("expected resources")
	}
	templates := map[string]mcpserver.ServerResourceTemplate{"rt": {Template: mcp.ResourceTemplate{URI: "t"}}}
	session.SetSessionResourceTemplates(templates)
	if len(session.GetSessionResourceTemplates()) != 1 {
		t.Fatalf("expected templates")
	}

	info := mcp.Implementation{Name: "client"}
	session.SetClientInfo(info)
	if session.GetClientInfo().Name != "client" {
		t.Fatalf("expected client info")
	}
	caps := mcp.ClientCapabilities{}
	session.SetClientCapabilities(caps)
	if session.GetClientCapabilities() != caps {
		t.Fatalf("expected client capabilities")
	}
}

func TestWriteJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := bufio.NewWriter(buf)
	mu := &sync.Mutex{}
	if err := writeJSON(writer, mu, map[string]string{"ok": "yes"}); err != nil {
		t.Fatalf("writeJSON error: %v", err)
	}
	if !bytes.HasSuffix(buf.Bytes(), []byte("\n")) {
		t.Fatalf("expected newline suffix")
	}
}

func TestStartTransportUnsupported(t *testing.T) {
	srv := &Server{opts: runtime.MCPOptions{Transport: "bogus"}}
	if err := srv.startTransport(); err == nil {
		t.Fatalf("expected unsupported transport error")
	}
}
