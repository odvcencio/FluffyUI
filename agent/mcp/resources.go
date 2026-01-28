package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/odvcencio/fluffyui/agent"
)

const (
	resourceScreenURI     = "fluffy://screen"
	resourceWidgetsURI    = "fluffy://widgets"
	resourceFocusedURI    = "fluffy://focused"
	resourceClipboardURI  = "fluffy://clipboard"
	resourceDimensionsURI = "fluffy://dimensions"
)

func registerResources(s *Server) {
	if s == nil || s.mcpServer == nil {
		return
	}

	s.mcpServer.AddResource(
		mcp.NewResource(resourceScreenURI, "screen",
			mcp.WithResourceDescription("Current screen text."),
			mcp.WithMIMEType("text/plain"),
		),
		s.handleResourceScreen,
	)
	s.mcpServer.AddResource(
		mcp.NewResource(resourceWidgetsURI, "widgets",
			mcp.WithResourceDescription("Widget tree snapshot."),
			mcp.WithMIMEType("application/json"),
		),
		s.handleResourceWidgets,
	)
	s.mcpServer.AddResource(
		mcp.NewResource(resourceFocusedURI, "focused",
			mcp.WithResourceDescription("Focused widget details."),
			mcp.WithMIMEType("application/json"),
		),
		s.handleResourceFocused,
	)
	s.mcpServer.AddResource(
		mcp.NewResource(resourceClipboardURI, "clipboard",
			mcp.WithResourceDescription("Clipboard contents."),
			mcp.WithMIMEType("text/plain"),
		),
		s.handleResourceClipboard,
	)
	s.mcpServer.AddResource(
		mcp.NewResource(resourceDimensionsURI, "dimensions",
			mcp.WithResourceDescription("Screen dimensions."),
			mcp.WithMIMEType("application/json"),
		),
		s.handleResourceDimensions,
	)

	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate("fluffy://widget/{id}", "widget",
			mcp.WithTemplateDescription("Widget details by ID."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		s.handleResourceWidget,
	)
	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate("fluffy://widget/{id}/value", "widget-value",
			mcp.WithTemplateDescription("Widget value by ID."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		s.handleResourceWidgetValue,
	)
	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate("fluffy://widget/{id}/children", "widget-children",
			mcp.WithTemplateDescription("Widget children by ID."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		s.handleResourceWidgetChildren,
	)
	s.mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate("fluffy://layer/{n}", "layer",
			mcp.WithTemplateDescription("Widget tree for a layer."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		s.handleResourceLayer,
	)
}

func (s *Server) handleResourceScreen(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	if !s.textAllowed() {
		return nil, textDeniedError("resources/read")
	}
	return resourceText(req.Params.URI, s.agent.CaptureText(), "text/plain"), nil
}

func (s *Server) handleResourceWidgets(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	snap, err := s.agent.SnapshotWithContext(ctx, agent.SnapshotOptions{})
	if err != nil {
		return nil, err
	}
	tree := treeSnapshotFromAgent(snap, false)
	return resourceJSON(req.Params.URI, tree)
}

func (s *Server) handleResourceFocused(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return nil, err
	}
	if snap.FocusedID == "" {
		return resourceJSON(req.Params.URI, (*WidgetInfo)(nil))
	}
	info := findWidgetByID(snap.Widgets, snap.FocusedID, false)
	return resourceJSON(req.Params.URI, info)
}

func (s *Server) handleResourceClipboard(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	if !s.clipboardAllowed() {
		return nil, clipboardDeniedError("resources/read")
	}
	text, err := s.clipboard().Read()
	if err != nil {
		return nil, err
	}
	return resourceText(req.Params.URI, text, "text/plain"), nil
}

func (s *Server) handleResourceDimensions(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return resourceJSON(req.Params.URI, s.currentDimensions())
}

func (s *Server) handleResourceWidget(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ref := parseResourceURI(req.Params.URI)
	if ref.kind != resourceWidget || ref.id == "" {
		return nil, newMCPError(mcp.INVALID_PARAMS, "invalid widget URI", nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return nil, err
	}
	info := findWidgetByID(snap.Widgets, ref.id, false)
	if info == nil {
		return nil, newMCPError(mcp.RESOURCE_NOT_FOUND, "widget not found", nil)
	}
	return resourceJSON(req.Params.URI, info)
}

func (s *Server) handleResourceWidgetValue(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ref := parseResourceURI(req.Params.URI)
	if ref.kind != resourceWidgetValue || ref.id == "" {
		return nil, newMCPError(mcp.INVALID_PARAMS, "invalid widget value URI", nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return nil, err
	}
	info := findWidgetByID(snap.Widgets, ref.id, false)
	if info == nil {
		return nil, newMCPError(mcp.RESOURCE_NOT_FOUND, "widget not found", nil)
	}
	return resourceJSON(req.Params.URI, info.Value)
}

func (s *Server) handleResourceWidgetChildren(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ref := parseResourceURI(req.Params.URI)
	if ref.kind != resourceWidgetChildren || ref.id == "" {
		return nil, newMCPError(mcp.INVALID_PARAMS, "invalid widget children URI", nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return nil, err
	}
	index := indexWidgets(snap.Widgets)
	parent, ok := index[ref.id]
	if !ok {
		return nil, newMCPError(mcp.RESOURCE_NOT_FOUND, "widget not found", nil)
	}
	children := collectWidgets(index, parent.ChildrenIDs)
	return resourceJSON(req.Params.URI, children)
}

func (s *Server) handleResourceLayer(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ref := parseResourceURI(req.Params.URI)
	if ref.kind != resourceLayer || ref.layer < 0 {
		return nil, newMCPError(mcp.INVALID_PARAMS, "invalid layer URI", nil)
	}
	snap, err := s.agent.SnapshotWithContext(ctx, agent.SnapshotOptions{})
	if err != nil {
		return nil, err
	}
	for _, widget := range snap.Widgets {
		if layerFromID(widget.ID) == ref.layer {
			node := widgetNodeFromAgent(widget)
			return resourceJSON(req.Params.URI, node)
		}
	}
	return nil, newMCPError(mcp.RESOURCE_NOT_FOUND, "layer not found", nil)
}

func resourceJSON(uri string, data any) ([]mcp.ResourceContents, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(encoded),
		},
	}, nil
}

func resourceText(uri, text, mime string) []mcp.ResourceContents {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: mime,
			Text:     text,
		},
	}
}

type resourceKind int

const (
	resourceUnknown resourceKind = iota
	resourceScreen
	resourceWidgets
	resourceFocused
	resourceClipboard
	resourceDimensions
	resourceWidget
	resourceWidgetValue
	resourceWidgetChildren
	resourceLayer
)

type resourceRef struct {
	uri         string
	kind        resourceKind
	id          string
	layer       int
	subresource string
}

func parseResourceURI(uri string) resourceRef {
	ref := resourceRef{uri: uri, kind: resourceUnknown, layer: -1}
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Scheme != "fluffy" {
		return ref
	}
	switch parsed.Host {
	case "screen":
		ref.kind = resourceScreen
		return ref
	case "widgets":
		ref.kind = resourceWidgets
		return ref
	case "focused":
		ref.kind = resourceFocused
		return ref
	case "clipboard":
		ref.kind = resourceClipboard
		return ref
	case "dimensions":
		ref.kind = resourceDimensions
		return ref
	case "widget":
		path := strings.TrimPrefix(parsed.Path, "/")
		if path == "" {
			return ref
		}
		parts := strings.Split(path, "/")
		ref.id = parts[0]
		ref.subresource = ""
		if len(parts) > 1 {
			switch parts[1] {
			case "value":
				ref.kind = resourceWidgetValue
				ref.subresource = "value"
				return ref
			case "children":
				ref.kind = resourceWidgetChildren
				ref.subresource = "children"
				return ref
			}
		}
		ref.kind = resourceWidget
		return ref
	case "layer":
		path := strings.TrimPrefix(parsed.Path, "/")
		if path == "" {
			return ref
		}
		layer, err := strconv.Atoi(path)
		if err != nil {
			return ref
		}
		ref.kind = resourceLayer
		ref.layer = layer
		return ref
	default:
		return ref
	}
}

type resourceWatcher struct {
	srv           *Server
	pollInterval  time.Duration
	prevSnapshot  Snapshot
	prevClipboard string
	initialized   bool
}

func newResourceWatcher(s *Server) *resourceWatcher {
	return &resourceWatcher{
		srv:          s,
		pollInterval: 100 * time.Millisecond,
	}
}

func (w *resourceWatcher) Run(ctx context.Context) {
	if w == nil || w.srv == nil {
		return
	}
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *resourceWatcher) tick(ctx context.Context) {
	s := w.srv
	if s == nil || s.mcpServer == nil {
		return
	}
	uris := s.mcpServer.SubscribedURIs()
	if len(uris) == 0 {
		w.initialized = false
		w.prevSnapshot = Snapshot{}
		w.prevClipboard = ""
		return
	}
	refs, needsText, needsClipboard := parseResourceRefs(uris)
	if needsText && !s.textAllowed() {
		needsText = false
	}
	snap, err := s.currentSnapshot(ctx, needsText)
	if err != nil {
		return
	}
	if !w.initialized {
		w.prevSnapshot = snap
		if needsClipboard && s.clipboardAllowed() {
			w.prevClipboard = w.readClipboard()
		}
		w.initialized = true
		return
	}

	diff := diffSnapshots(w.prevSnapshot, snap)
	modified := make(map[string]WidgetChange, len(diff.WidgetsModified))
	for _, change := range diff.WidgetsModified {
		modified[change.ID] = change
	}
	beforeIndex := indexWidgets(w.prevSnapshot.Widgets)
	afterIndex := indexWidgets(snap.Widgets)
	explicitIndex := explicitIndexFromWidgets(snap.Widgets)
	changedLayers := computeChangedLayers(w.prevSnapshot, snap, diff)

	clipboardChanged := false
	clipboardText := w.prevClipboard
	if needsClipboard && s.clipboardAllowed() {
		clipboardText = w.readClipboard()
		clipboardChanged = clipboardText != w.prevClipboard
	}

	focusedChanged := diff.FocusChanged
	focusedWidgetChanged := false
	if snap.FocusedID != "" && !diff.FocusChanged {
		if change, ok := modified[snap.FocusedID]; ok && len(change.Changes) > 0 {
			focusedWidgetChanged = true
		}
	}

	widgetsChanged := diff.LayerCountChanged || diff.DimensionsChanged || diff.FocusChanged
	if len(diff.WidgetsAdded) > 0 || len(diff.WidgetsRemoved) > 0 || len(diff.WidgetsModified) > 0 {
		widgetsChanged = true
	}

	for _, ref := range refs {
		switch ref.kind {
		case resourceScreen:
			if diff.TextChanged {
				s.mcpServer.SendResourceUpdated(ref.uri)
			}
		case resourceWidgets:
			if widgetsChanged {
				s.mcpServer.SendResourceUpdated(ref.uri)
			}
		case resourceFocused:
			if focusedChanged || focusedWidgetChanged {
				s.mcpServer.SendResourceUpdated(ref.uri)
			}
		case resourceClipboard:
			if clipboardChanged {
				s.mcpServer.SendResourceUpdated(ref.uri)
			}
		case resourceDimensions:
			if diff.DimensionsChanged {
				s.mcpServer.SendResourceUpdated(ref.uri)
			}
		case resourceLayer:
			if ref.layer >= 0 && changedLayers[ref.layer] {
				s.mcpServer.SendResourceUpdated(ref.uri)
			}
		case resourceWidget, resourceWidgetValue, resourceWidgetChildren:
			w.handleWidgetResource(ref, beforeIndex, afterIndex, modified, explicitIndex)
		}
	}

	w.prevSnapshot = snap
	if needsClipboard && s.clipboardAllowed() {
		w.prevClipboard = clipboardText
	}
}

func (w *resourceWatcher) handleWidgetResource(
	ref resourceRef,
	beforeIndex map[string]WidgetInfo,
	afterIndex map[string]WidgetInfo,
	modified map[string]WidgetChange,
	explicitIndex map[string][]string,
) {
	s := w.srv
	if s == nil || s.mcpServer == nil {
		return
	}
	if ref.id == "" {
		return
	}
	beforeWidget, beforeOK := beforeIndex[ref.id]
	afterWidget, afterOK := afterIndex[ref.id]
	if !afterOK {
		if newURI := widgetIDChangedURI(ref, explicitIndex); newURI != "" {
			s.mcpServer.SendResourceUpdatedWithParams(ref.uri, map[string]any{
				"reason":  "widget_id_changed",
				"new_uri": newURI,
			})
			return
		}
		if beforeOK {
			s.mcpServer.SendResourceUpdated(ref.uri)
		}
		return
	}
	if !beforeOK {
		s.mcpServer.SendResourceUpdated(ref.uri)
		return
	}
	change, changed := modified[ref.id]
	if !changed && ref.kind == resourceWidget {
		return
	}
	switch ref.kind {
	case resourceWidget:
		if changed {
			s.mcpServer.SendResourceUpdated(ref.uri)
		}
	case resourceWidgetValue:
		if !beforeOK || !afterOK || beforeWidget.Value != afterWidget.Value {
			s.mcpServer.SendResourceUpdated(ref.uri)
		}
	case resourceWidgetChildren:
		if !beforeOK || !afterOK || !equalStringSlice(beforeWidget.ChildrenIDs, afterWidget.ChildrenIDs) || changedChildren(change) {
			s.mcpServer.SendResourceUpdated(ref.uri)
		}
	}
}

func changedChildren(change WidgetChange) bool {
	if len(change.Changes) == 0 {
		return false
	}
	if _, ok := change.Changes["children_ids"]; ok {
		return true
	}
	return false
}

func widgetIDChangedURI(ref resourceRef, explicitIndex map[string][]string) string {
	explicit := explicitIDFromWidgetID(ref.id)
	base := explicitBaseID(explicit)
	if base == "" {
		return ""
	}
	candidates := explicitIndex[base]
	if len(candidates) != 1 {
		return ""
	}
	newID := candidates[0]
	if newID == ref.id {
		return ""
	}
	if ref.subresource == "" {
		return fmt.Sprintf("fluffy://widget/%s", newID)
	}
	return fmt.Sprintf("fluffy://widget/%s/%s", newID, ref.subresource)
}

func explicitIndexFromWidgets(widgets []WidgetInfo) map[string][]string {
	out := make(map[string][]string)
	for _, widget := range widgets {
		explicit := explicitIDFromWidgetID(widget.ID)
		base := explicitBaseID(explicit)
		if base == "" {
			continue
		}
		out[base] = append(out[base], widget.ID)
	}
	return out
}

func computeChangedLayers(before Snapshot, after Snapshot, diff Diff) map[int]bool {
	changed := make(map[int]bool)
	if diff.LayerCountChanged {
		maxLayer := max(before.LayerCount, after.LayerCount)
		for i := 0; i < maxLayer; i++ {
			changed[i] = true
		}
	}
	for _, id := range diff.WidgetsAdded {
		changed[layerFromID(id)] = true
	}
	for _, id := range diff.WidgetsRemoved {
		changed[layerFromID(id)] = true
	}
	for _, change := range diff.WidgetsModified {
		changed[layerFromID(change.ID)] = true
	}
	return changed
}

func parseResourceRefs(uris []string) ([]resourceRef, bool, bool) {
	refs := make([]resourceRef, 0, len(uris))
	needsText := false
	needsClipboard := false
	for _, uri := range uris {
		ref := parseResourceURI(uri)
		refs = append(refs, ref)
		switch ref.kind {
		case resourceScreen:
			needsText = true
		case resourceClipboard:
			needsClipboard = true
		}
	}
	return refs, needsText, needsClipboard
}

func (w *resourceWatcher) readClipboard() string {
	if w == nil || w.srv == nil {
		return ""
	}
	text, err := w.srv.clipboard().Read()
	if err != nil {
		return ""
	}
	return text
}
