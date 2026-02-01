package agent

import (
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

type testWidget struct {
	bounds   runtime.Rect
	id       string
	focus    bool
	a11y     accessibility.Base
	text     string
	canFocus bool
}

func (t *testWidget) Measure(runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 1, Height: 1}
}
func (t *testWidget) Layout(bounds runtime.Rect)   { t.bounds = bounds }
func (t *testWidget) Render(runtime.RenderContext) {}
func (t *testWidget) HandleMessage(runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (t *testWidget) Bounds() runtime.Rect { return t.bounds }

func (t *testWidget) ID() string { return t.id }

func (t *testWidget) IsFocused() bool { return t.focus }
func (t *testWidget) SetFocused(v bool) {
	t.focus = v
}
func (t *testWidget) CanFocus() bool { return t.canFocus }

func (t *testWidget) AccessibleRole() accessibility.Role { return t.a11y.Role }
func (t *testWidget) AccessibleLabel() string            { return t.a11y.Label }
func (t *testWidget) AccessibleDescription() string      { return t.a11y.Description }
func (t *testWidget) AccessibleState() accessibility.StateSet {
	return t.a11y.State
}
func (t *testWidget) AccessibleValue() *accessibility.ValueInfo { return t.a11y.Value }

func (t *testWidget) Text() string { return t.text }

func TestBuildWidgetIDExplicitCollision(t *testing.T) {
	w := &testWidget{id: "btn"}
	counts := map[string]int{}
	id1 := buildWidgetID(w, 0, []int{0}, counts, false)
	id2 := buildWidgetID(w, 0, []int{1}, counts, false)
	if id1 == id2 {
		t.Fatalf("expected unique ids, got %q", id1)
	}
	if id2 != "layer0:testwidget:*:btn#2" {
		t.Fatalf("id2 = %q", id2)
	}
}

func TestBuildWidgetIDPath(t *testing.T) {
	w := &testWidget{}
	counts := map[string]int{}
	id := buildWidgetID(w, 2, []int{0, 3, 1}, counts, false)
	if id != "layer2:testwidget:0.3.1" {
		t.Fatalf("id = %q", id)
	}
}

func TestActionsForRole(t *testing.T) {
	if actionsForRole(accessibility.RoleButton, accessibility.StateSet{})[0] != "activate" {
		t.Fatalf("expected activate action")
	}
	if actionsForRole(accessibility.RoleTextbox, accessibility.StateSet{})[1] != "clear" {
		t.Fatalf("expected clear action")
	}
	if actionsForRole(accessibility.RoleList, accessibility.StateSet{})[2] != "scroll" {
		t.Fatalf("expected scroll action")
	}
	if actionsForRole(accessibility.RoleMenuItem, accessibility.StateSet{})[0] != "activate" {
		t.Fatalf("expected activate action for menu item")
	}
	if actionsForRole(accessibility.RoleButton, accessibility.StateSet{Disabled: true}) != nil {
		t.Fatalf("expected disabled to return no actions")
	}
}

func TestWidgetHelpers(t *testing.T) {
	if widgetTypeName(nil) != "widget" {
		t.Fatalf("expected widget for nil type")
	}
	if formatWidgetPath(nil) != "0" {
		t.Fatalf("expected default path")
	}
	if formatWidgetPath([]int{1, 2, 3}) != "1.2.3" {
		t.Fatalf("unexpected path")
	}

	w := &testWidget{id: " id "}
	if widgetExplicitID(w) != "id" {
		t.Fatalf("expected trimmed id")
	}
	if defaultWidgetLabel(w) != "testWidget" {
		t.Fatalf("expected default label")
	}
}

func TestFindByLabelAndFocused(t *testing.T) {
	widgets := []WidgetInfo{
		{Label: "Root", Children: []WidgetInfo{{Label: "Child"}}},
		{Label: "Another", Focused: true},
	}
	found := findByLabelIn(widgets, "child")
	if found == nil || found.Label != "Child" {
		t.Fatalf("expected to find child")
	}
	focused := findFocusedInfo(&widgets[1])
	if focused == nil || focused.Label != "Another" {
		t.Fatalf("expected focused widget")
	}
}

func TestRemoveAction(t *testing.T) {
	actions := []string{"focus", "activate", "focus"}
	out := removeAction(actions, "focus")
	if len(out) != 1 || out[0] != "activate" {
		t.Fatalf("unexpected removeAction result: %v", out)
	}
	unchanged := removeAction(actions, " ")
	if len(unchanged) != len(actions) {
		t.Fatalf("expected no change for empty action")
	}
	for i := range actions {
		if unchanged[i] != actions[i] {
			t.Fatalf("expected no change for empty action")
		}
	}
}
