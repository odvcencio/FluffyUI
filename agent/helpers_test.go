package agent

import (
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

type simpleAccessible struct {
	label string
	value *accessibility.ValueInfo
}

func (s simpleAccessible) Measure(runtime.Constraints) runtime.Size { return runtime.Size{} }
func (s simpleAccessible) Layout(runtime.Rect)                      {}
func (s simpleAccessible) Render(runtime.RenderContext)             {}
func (s simpleAccessible) HandleMessage(runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (s simpleAccessible) AccessibleRole() accessibility.Role { return accessibility.RoleText }
func (s simpleAccessible) AccessibleLabel() string            { return s.label }
func (s simpleAccessible) AccessibleDescription() string      { return "" }
func (s simpleAccessible) AccessibleState() accessibility.StateSet {
	return accessibility.StateSet{}
}
func (s simpleAccessible) AccessibleValue() *accessibility.ValueInfo { return s.value }
func (s simpleAccessible) AccessibleLive() accessibility.Live           { return "" }
func (s simpleAccessible) AccessibleRelevant() accessibility.Relevant   { return "" }
func (s simpleAccessible) AccessibleAtomic() bool                       { return false }
func (s simpleAccessible) AccessibleLandmark() accessibility.Landmark   { return "" }
func (s simpleAccessible) AccessibleLabelledBy() string                 { return "" }
func (s simpleAccessible) AccessibleDescribedBy() string                { return "" }
func (s simpleAccessible) AccessibleControls() string                   { return "" }
func (s simpleAccessible) AccessibleOwns() string                      { return "" }
func (s simpleAccessible) AccessibleFlowTo() string                    { return "" }
func (s simpleAccessible) AccessibleLevel() int                        { return 0 }
func (s simpleAccessible) AccessibleOrientation() string               { return "" }
func (s simpleAccessible) AccessibleActiveDescendant() string          { return "" }
func (s simpleAccessible) AccessiblePosInSet() int                     { return 0 }
func (s simpleAccessible) AccessibleSetSize() int                      { return 0 }
func (s simpleAccessible) AccessibleHasPopup() string                  { return "" }
func (s simpleAccessible) AccessibleErrorMessage() string              { return "" }
func (s simpleAccessible) AccessibleCurrent() string                   { return "" }
func (s simpleAccessible) AccessibleAutocomplete() string              { return "" }
func (s simpleAccessible) AccessiblePlaceholder() string               { return "" }
func (s simpleAccessible) AccessibleSort() string                      { return "" }
func (s simpleAccessible) AccessibleKeyShortcuts() string              { return "" }
func (s simpleAccessible) AccessibleDetails() string                   { return "" }
func (s simpleAccessible) AccessibleRoleDescription() string           { return "" }

func TestAccessibleChoice(t *testing.T) {
	if accessibleChoice(nil) != "" {
		t.Fatalf("expected empty for nil")
	}
	acc := simpleAccessible{label: "Label"}
	if accessibleChoice(acc) != "Label" {
		t.Fatalf("expected label choice")
	}
	acc = simpleAccessible{label: "Label", value: &accessibility.ValueInfo{Text: "Value"}}
	if accessibleChoice(acc) != "Value" {
		t.Fatalf("expected value choice")
	}
}

func TestAccessibleFromWidget(t *testing.T) {
	if accessibleFromWidget(nil) != nil {
		t.Fatalf("expected nil for nil widget")
	}
	acc := simpleAccessible{label: "ok"}
	if accessibleFromWidget(acc) == nil {
		t.Fatalf("expected accessible from widget")
	}
}
