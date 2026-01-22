package runtime

import (
	"reflect"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/style"
)

// StyleTypeProvider overrides the selector type for a widget.
type StyleTypeProvider interface {
	StyleType() string
}

// StyleIDProvider supplies a selector ID.
type StyleIDProvider interface {
	StyleID() string
}

// StyleClassProvider supplies selector classes.
type StyleClassProvider interface {
	StyleClasses() []string
}

// StyleStateProvider supplies widget pseudo-class state.
type StyleStateProvider interface {
	StyleState() style.WidgetState
}

type styleNode struct {
	typ     string
	id      string
	classes []string
	state   style.WidgetState
}

func (n styleNode) StyleType() string {
	return n.typ
}

func (n styleNode) StyleID() string {
	return n.id
}

func (n styleNode) StyleClasses() []string {
	return n.classes
}

func (n styleNode) StyleState() style.WidgetState {
	return n.state
}

// StyleResolver resolves stylesheet rules against widgets.
type StyleResolver struct {
	sheet   *style.Stylesheet
	parents map[Widget]Widget
	cache   map[Widget]style.Style
}

func newStyleResolver(sheet *style.Stylesheet, roots []Widget) *StyleResolver {
	if sheet == nil {
		return nil
	}
	return &StyleResolver{
		sheet:   sheet,
		parents: buildParentMap(roots),
		cache:   make(map[Widget]style.Style),
	}
}

func buildParentMap(roots []Widget) map[Widget]Widget {
	parents := make(map[Widget]Widget)
	var walk func(parent, node Widget)
	walk = func(parent, node Widget) {
		if node == nil {
			return
		}
		if parent != nil {
			parents[node] = parent
		}
		if container, ok := node.(ChildProvider); ok {
			for _, child := range container.ChildWidgets() {
				walk(node, child)
			}
		}
	}
	for _, root := range roots {
		walk(nil, root)
	}
	return parents
}

// Resolve returns the resolved style for a widget.
func (r *StyleResolver) Resolve(widget Widget, focused bool) style.Style {
	if r == nil || widget == nil {
		return style.Style{}
	}
	if cached, ok := r.cache[widget]; ok {
		return cached
	}
	var parentStyle style.Style
	if parent, ok := r.parents[widget]; ok && parent != nil {
		parentStyle = r.Resolve(parent, focused)
	}
	node := r.nodeFor(widget, focused)
	ancestors := r.ancestors(widget, focused)
	resolved := r.sheet.Resolve(node, ancestors)
	resolved = resolved.Inherit(parentStyle)
	r.cache[widget] = resolved
	return resolved
}

func (r *StyleResolver) ancestors(widget Widget, focused bool) []style.Node {
	var nodes []style.Node
	for parent := r.parents[widget]; parent != nil; parent = r.parents[parent] {
		nodes = append(nodes, r.nodeFor(parent, focused))
	}
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
	return nodes
}

func (r *StyleResolver) nodeFor(widget Widget, focused bool) style.Node {
	state := r.stateFor(widget, focused)
	return styleNode{
		typ:     r.typeFor(widget),
		id:      r.idFor(widget),
		classes: r.classesFor(widget),
		state:   state,
	}
}

func (r *StyleResolver) typeFor(widget Widget) string {
	if widget == nil {
		return ""
	}
	if provider, ok := widget.(StyleTypeProvider); ok {
		if typ := strings.TrimSpace(provider.StyleType()); typ != "" {
			return typ
		}
	}
	typ := reflect.TypeOf(widget)
	for typ != nil && typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ == nil {
		return ""
	}
	return typ.Name()
}

func (r *StyleResolver) idFor(widget Widget) string {
	if widget == nil {
		return ""
	}
	if provider, ok := widget.(StyleIDProvider); ok {
		return strings.TrimSpace(provider.StyleID())
	}
	if provider, ok := widget.(interface{ ID() string }); ok {
		return strings.TrimSpace(provider.ID())
	}
	return ""
}

func (r *StyleResolver) classesFor(widget Widget) []string {
	if widget == nil {
		return nil
	}
	if provider, ok := widget.(StyleClassProvider); ok {
		return normalizeClasses(provider.StyleClasses())
	}
	return nil
}

func (r *StyleResolver) stateFor(widget Widget, focused bool) style.WidgetState {
	state := style.WidgetState{}
	if widget == nil {
		return state
	}
	if provider, ok := widget.(StyleStateProvider); ok {
		state = provider.StyleState()
	} else {
		if focusable, ok := widget.(Focusable); ok && focusable.IsFocused() {
			state.Focused = true
		}
		if accessible, ok := widget.(accessibility.Accessible); ok {
			state.Disabled = accessible.AccessibleState().Disabled
		}
	}
	if !focused {
		state.Focused = false
	}
	state.FirstChild, state.LastChild = r.childPosition(widget)
	return state
}

func (r *StyleResolver) childPosition(widget Widget) (bool, bool) {
	parent := r.parents[widget]
	if parent == nil {
		return false, false
	}
	container, ok := parent.(ChildProvider)
	if !ok {
		return false, false
	}
	children := container.ChildWidgets()
	if len(children) == 0 {
		return false, false
	}
	first := children[0] == widget
	last := children[len(children)-1] == widget
	return first, last
}

func normalizeClasses(classes []string) []string {
	if len(classes) == 0 {
		return nil
	}
	out := make([]string, 0, len(classes))
	for _, cls := range classes {
		name := strings.TrimSpace(cls)
		if name == "" {
			continue
		}
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
