# Tutorial 04: AI Agent Integration

This tutorial exposes a FluffyUI app to external automation.

## Start the Agent Server

```go
srv, _ := agent.NewServer(agent.ServerOptions{
    Addr:      "tcp:127.0.0.1:7777",
    App:       app,
    AllowText: true,
})

go srv.Serve(context.Background())
```

## Send Commands

Use JSONL over TCP to send requests:

```json
{"id":1,"type":"hello"}
{"id":2,"type":"snapshot","include_text":true}
```

## Reference

- `examples/ai-agent-demo` includes a runnable demo and Python client.
- `agent/mcp` provides MCP server tooling for LLM integration.
