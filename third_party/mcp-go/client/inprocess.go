package client

import (
	"context"

	"m31labs.dev/fluffyui/third_party/mcp-go/client/transport"
	"m31labs.dev/fluffyui/third_party/mcp-go/mcp"
	"m31labs.dev/fluffyui/third_party/mcp-go/server"
)

// NewInProcessClient connect directly to a mcp server object in the same process
func NewInProcessClient(server *server.MCPServer) (*Client, error) {
	inProcessTransport := transport.NewInProcessTransport(server)
	return NewClient(inProcessTransport), nil
}

// NewInProcessClientWithSamplingHandler creates an in-process client with sampling support
func NewInProcessClientWithSamplingHandler(server *server.MCPServer, handler SamplingHandler) (*Client, error) {
	// Create a wrapper that implements server.SamplingHandler
	serverHandler := &inProcessSamplingHandlerWrapper{handler: handler}

	inProcessTransport := transport.NewInProcessTransportWithOptions(server,
		transport.WithSamplingHandler(serverHandler))

	client := NewClient(inProcessTransport)
	client.samplingHandler = handler

	return client, nil
}

// inProcessSamplingHandlerWrapper wraps client.SamplingHandler to implement server.SamplingHandler
type inProcessSamplingHandlerWrapper struct {
	handler SamplingHandler
}

func (w *inProcessSamplingHandlerWrapper) CreateMessage(ctx context.Context, request mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	return w.handler.CreateMessage(ctx, request)
}
