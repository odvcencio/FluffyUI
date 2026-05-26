package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mcpclient "m31labs.dev/fluffyui/third_party/mcp-go/client"
	mcptr "m31labs.dev/fluffyui/third_party/mcp-go/client/transport"
	mcp "m31labs.dev/fluffyui/third_party/mcp-go/mcp"
)

type initTransport struct {
	started bool
	closed  bool
}

func (t *initTransport) Start(context.Context) error {
	t.started = true
	return nil
}

func (t *initTransport) SendRequest(_ context.Context, request mcptr.JSONRPCRequest) (*mcptr.JSONRPCResponse, error) {
	switch request.Method {
	case "initialize":
		result := mcp.InitializeResult{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ServerCapabilities{},
			ServerInfo:      mcp.Implementation{Name: "test", Version: "dev"},
		}
		raw, _ := json.Marshal(result)
		return &mcptr.JSONRPCResponse{JSONRPC: mcp.JSONRPC_VERSION, ID: request.ID, Result: raw}, nil
	default:
		raw, _ := json.Marshal(map[string]any{})
		return &mcptr.JSONRPCResponse{JSONRPC: mcp.JSONRPC_VERSION, ID: request.ID, Result: raw}, nil
	}
}

func (t *initTransport) SendNotification(context.Context, mcp.JSONRPCNotification) error { return nil }
func (t *initTransport) SetNotificationHandler(func(mcp.JSONRPCNotification))            {}
func (t *initTransport) Close() error {
	t.closed = true
	return nil
}
func (t *initTransport) GetSessionId() string { return "init" }

func TestClientInitializeAndStartTransport(t *testing.T) {
	transport := &initTransport{}
	inner := mcpclient.NewClient(transport)
	if err := startTransport(inner, 5*time.Millisecond); err != nil {
		t.Fatalf("start transport error: %v", err)
	}
	if !transport.started {
		t.Fatalf("expected transport started")
	}

	client := &Client{inner: inner}
	if err := client.initialize(ClientOptions{Token: "token", SchemaVersion: "v1"}); err != nil {
		t.Fatalf("initialize error: %v", err)
	}
}
