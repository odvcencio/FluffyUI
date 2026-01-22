package style

import "testing"

type testNode struct {
	typ     string
	id      string
	classes []string
	state   WidgetState
}

func (n *testNode) StyleType() string {
	return n.typ
}

func (n *testNode) StyleID() string {
	return n.id
}

func (n *testNode) StyleClasses() []string {
	return n.classes
}

func (n *testNode) StyleState() WidgetState {
	return n.state
}

func TestResolveSpecificity(t *testing.T) {
	sheet := NewStylesheet().
		Add(Select("Button"), Style{Foreground: RGB(10, 20, 30)}).
		Add(Select("Button").Class("primary"), Style{Foreground: RGB(40, 50, 60)}).
		Add(SelectID("submit"), Style{Foreground: RGB(70, 80, 90)})

	node := &testNode{
		typ:     "Button",
		id:      "submit",
		classes: []string{"primary"},
	}

	resolved := sheet.Resolve(node, nil)
	if resolved.Foreground != RGB(70, 80, 90) {
		t.Fatalf("foreground = %#v, want %#v", resolved.Foreground, RGB(70, 80, 90))
	}
}

func TestResolveOrderTiebreaker(t *testing.T) {
	sheet := NewStylesheet().
		Add(Select("Button"), Style{Bold: Bool(true)}).
		Add(Select("Button"), Style{Bold: Bool(false)})

	node := &testNode{typ: "Button"}
	resolved := sheet.Resolve(node, nil)
	if resolved.Bold == nil || *resolved.Bold {
		t.Fatalf("bold = %v, want false", resolved.Bold)
	}
}

func TestResolveDescendantAndPseudo(t *testing.T) {
	sheet := NewStylesheet().
		Add(Select("Input").Inside("Dialog").Pseudo(PseudoFocus), Style{Underline: Bool(true)})

	dialog := &testNode{typ: "Dialog"}
	input := &testNode{
		typ: "Input",
		state: WidgetState{
			Focused: true,
		},
	}

	resolved := sheet.Resolve(input, []Node{dialog})
	if resolved.Underline == nil || !*resolved.Underline {
		t.Fatalf("underline = %v, want true", resolved.Underline)
	}
}
