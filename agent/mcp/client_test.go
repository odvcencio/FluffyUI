package mcp

import (
	"reflect"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
)

type samplePayload struct {
	ID string
}

func TestParseSubscriptionHandler(t *testing.T) {
	if _, err := parseSubscriptionHandler(123); err == nil {
		t.Fatalf("expected error for non-function handler")
	}
	if _, err := parseSubscriptionHandler(func(a, b int) {}); err == nil {
		t.Fatalf("expected error for wrong arity handler")
	}
	if _, err := parseSubscriptionHandler(func(a int) int { return a }); err == nil {
		t.Fatalf("expected error for handler with return")
	}

	gotEvent := false
	handler := func(evt ResourceEvent) { gotEvent = evt.URI != "" }
	sub, err := parseSubscriptionHandler(handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sub.expectsEvent {
		t.Fatalf("expected event handler")
	}
	if sub.argType != reflect.TypeOf(ResourceEvent{}) {
		t.Fatalf("unexpected arg type")
	}

	ptrHandler := func(evt *ResourceEvent) { gotEvent = evt != nil }
	sub, err = parseSubscriptionHandler(ptrHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sub.expectsEvent || sub.argType.Kind() != reflect.Pointer {
		t.Fatalf("expected pointer event handler")
	}
	_ = gotEvent
}

func TestInvokeHandlerConversions(t *testing.T) {
	client := &Client{}

	called := false
	handler := func(p samplePayload) {
		called = true
		if p.ID != "ok" {
			t.Fatalf("unexpected payload value")
		}
	}
	sub, err := parseSubscriptionHandler(handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	client.invokeHandler(sub, reflect.ValueOf(samplePayload{ID: "ok"}))
	if !called {
		t.Fatalf("expected handler to be called")
	}

	called = false
	ptrHandler := func(p *samplePayload) {
		called = true
		if p == nil || p.ID != "ptr" {
			t.Fatalf("unexpected pointer payload")
		}
	}
	sub, err = parseSubscriptionHandler(ptrHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	client.invokeHandler(sub, reflect.ValueOf(samplePayload{ID: "ptr"}))
	if !called {
		t.Fatalf("expected pointer handler to be called")
	}

	called = false
	sub, _ = parseSubscriptionHandler(handler)
	client.invokeHandler(sub, reflect.ValueOf(&samplePayload{ID: "ok"}))
	if !called {
		t.Fatalf("expected value handler to accept pointer")
	}

	called = false
	client.invokeHandler(sub, reflect.ValueOf(123))
	if called {
		t.Fatalf("expected handler to ignore mismatched type")
	}
	client.invokeHandler(sub, reflect.Value{})
	if called {
		t.Fatalf("expected handler to ignore invalid value")
	}
}

func TestHandleNotificationEvent(t *testing.T) {
	client := &Client{}
	ch := make(chan ResourceEvent, 1)
	handler := func(evt ResourceEvent) { ch <- evt }
	sub, err := parseSubscriptionHandler(handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	client.subscribers = map[string][]subscriptionHandler{
		"fluffy://widgets": {sub},
	}

	note := mcp.JSONRPCNotification{
		Notification: mcp.Notification{
			Method: mcp.MethodNotificationResourceUpdated,
			Params: mcp.NotificationParams{AdditionalFields: map[string]any{
				"uri":     "fluffy://widgets",
				"reason":  "changed",
				"new_uri": "fluffy://widgets/updated",
			}},
		},
	}
	client.handleNotification(note)

	select {
	case evt := <-ch:
		if evt.URI != "fluffy://widgets" || evt.Reason != "changed" || evt.NewURI != "fluffy://widgets/updated" {
			t.Fatalf("unexpected event: %#v", evt)
		}
	default:
		t.Fatalf("expected handler to be invoked")
	}
}

func TestResourceTextFromResult(t *testing.T) {
	if _, _, err := resourceTextFromResult(nil); err == nil {
		t.Fatalf("expected error for nil result")
	}

	result := &mcp.ReadResourceResult{Contents: []mcp.ResourceContents{mcp.BlobResourceContents{Blob: "AA=="}}}
	if _, _, err := resourceTextFromResult(result); err == nil {
		t.Fatalf("expected error for non-text resource")
	}

	textResult := &mcp.ReadResourceResult{Contents: []mcp.ResourceContents{mcp.TextResourceContents{Text: "ok", MIMEType: "text/plain"}}}
	text, mime, err := resourceTextFromResult(textResult)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "ok" || mime != "text/plain" {
		t.Fatalf("unexpected resource text result")
	}

	ptrResult := &mcp.ReadResourceResult{Contents: []mcp.ResourceContents{&mcp.TextResourceContents{Text: "ptr", MIMEType: "text/plain"}}}
	text, mime, err = resourceTextFromResult(ptrResult)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "ptr" || mime != "text/plain" {
		t.Fatalf("unexpected resource text result for pointer")
	}
}
