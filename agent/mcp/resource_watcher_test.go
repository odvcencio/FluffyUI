package mcp

import (
	"context"
	"fmt"
	"testing"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/odvcencio/fluffyui/runtime"
)

func TestResourceWatcherHandleWidgetResourceRemoved(t *testing.T) {
	srv, session, ctx := newWatcherServer(t)
	uri := "fluffy://widget/layer0:button:0:save"
	subscribeURI(t, srv, ctx, uri)

	watcher := &resourceWatcher{srv: srv}
	id := "layer0:button:0:save"
	ref := resourceRef{uri: uri, kind: resourceWidget, id: id}
	before := map[string]WidgetInfo{id: {ID: id}}
	watcher.handleWidgetResource(ref, before, map[string]WidgetInfo{}, map[string]WidgetChange{}, map[string][]string{})

	notif := expectNotification(t, session.ch)
	if notif.Method != mcp.MethodNotificationResourceUpdated {
		t.Fatalf("unexpected method: %s", notif.Method)
	}
	if notif.Params.AdditionalFields["uri"].(string) != uri {
		t.Fatalf("unexpected uri")
	}
}

func TestResourceWatcherHandleWidgetResourceIDChanged(t *testing.T) {
	srv, session, ctx := newWatcherServer(t)
	oldID := "layer0:button:0:save#2"
	newID := "layer0:button:1:save"
	uri := fmt.Sprintf("fluffy://widget/%s", oldID)
	subscribeURI(t, srv, ctx, uri)

	watcher := &resourceWatcher{srv: srv}
	ref := resourceRef{uri: uri, kind: resourceWidget, id: oldID}
	before := map[string]WidgetInfo{oldID: {ID: oldID}}
	explicit := map[string][]string{"save": {newID}}
	watcher.handleWidgetResource(ref, before, map[string]WidgetInfo{}, map[string]WidgetChange{}, explicit)

	notif := expectNotification(t, session.ch)
	fields := notif.Params.AdditionalFields
	if fields["reason"].(string) != "widget_id_changed" {
		t.Fatalf("unexpected reason")
	}
	if fields["new_uri"].(string) != fmt.Sprintf("fluffy://widget/%s", newID) {
		t.Fatalf("unexpected new uri")
	}
}

func TestResourceWatcherHandleWidgetResourceValueAndChildren(t *testing.T) {
	srv, session, ctx := newWatcherServer(t)
	valueID := "layer0:input:0:name"
	valueURI := fmt.Sprintf("fluffy://widget/%s/value", valueID)
	childrenID := "layer0:stack:0"
	childrenURI := fmt.Sprintf("fluffy://widget/%s/children", childrenID)
	subscribeURI(t, srv, ctx, valueURI)
	subscribeURI(t, srv, ctx, childrenURI)

	watcher := &resourceWatcher{srv: srv}

	beforeValue := map[string]WidgetInfo{valueID: {ID: valueID, Value: "a"}}
	afterValue := map[string]WidgetInfo{valueID: {ID: valueID, Value: "b"}}
	refValue := resourceRef{uri: valueURI, kind: resourceWidgetValue, id: valueID, subresource: "value"}
	watcher.handleWidgetResource(refValue, beforeValue, afterValue, map[string]WidgetChange{}, map[string][]string{})

	beforeChildren := map[string]WidgetInfo{childrenID: {ID: childrenID, ChildrenIDs: []string{"child1"}}}
	afterChildren := map[string]WidgetInfo{childrenID: {ID: childrenID, ChildrenIDs: []string{"child1"}}}
	changes := map[string]WidgetChange{childrenID: {ID: childrenID, Changes: map[string]ValueChange{"children_ids": {}}}}
	refChildren := resourceRef{uri: childrenURI, kind: resourceWidgetChildren, id: childrenID, subresource: "children"}
	watcher.handleWidgetResource(refChildren, beforeChildren, afterChildren, changes, map[string][]string{})

	// Expect two notifications in any order.
	n1 := expectNotification(t, session.ch)
	n2 := expectNotification(t, session.ch)
	uris := map[string]bool{
		n1.Params.AdditionalFields["uri"].(string): true,
		n2.Params.AdditionalFields["uri"].(string): true,
	}
	if !uris[valueURI] || !uris[childrenURI] {
		t.Fatalf("unexpected notification uris")
	}
}

func newWatcherServer(t *testing.T) (*Server, testSession, context.Context) {
	t.Helper()
	server := &Server{
		opts:      runtime.MCPOptions{AllowText: true, AllowClipboard: true},
		sessions:  make(map[string]*sessionState),
		mcpServer: mcpserver.NewMCPServer("test", "dev", mcpserver.WithResourceCapabilities(true, false)),
	}
	session := testSession{id: "sess", initialized: true, ch: make(chan mcp.JSONRPCNotification, 5)}
	ctx := server.mcpServer.WithContext(context.Background(), session)
	if err := server.mcpServer.RegisterSession(ctx, session); err != nil {
		t.Fatalf("register session error: %v", err)
	}
	return server, session, ctx
}

func subscribeURI(t *testing.T, server *Server, ctx context.Context, uri string) {
	t.Helper()
	req := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"resources/subscribe","params":{"uri":"%s"}}`, uri)
	resp := server.mcpServer.HandleMessage(ctx, []byte(req))
	if errResp, ok := resp.(mcp.JSONRPCError); ok {
		t.Fatalf("subscribe error: %s", errResp.Error.Message)
	}
}

func expectNotification(t *testing.T, ch <-chan mcp.JSONRPCNotification) mcp.JSONRPCNotification {
	t.Helper()
	select {
	case note := <-ch:
		return note
	case <-time.After(250 * time.Millisecond):
		t.Fatalf("timeout waiting for notification")
		return mcp.JSONRPCNotification{}
	}
}
