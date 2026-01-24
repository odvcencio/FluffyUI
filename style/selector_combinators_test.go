package style

import "testing"

func TestChildCombinator(t *testing.T) {
	data := `
Panel > Button { bold: true; }
Panel Button { underline: true; }
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	panel := &testNode{typ: "Panel"}
	container := &testNode{typ: "Container"}
	button := &testNode{typ: "Button"}

	resolvedDirect := sheet.Resolve(button, []Node{panel})
	if resolvedDirect.Bold == nil || !*resolvedDirect.Bold {
		t.Fatalf("direct child bold = %v, want true", resolvedDirect.Bold)
	}
	if resolvedDirect.Underline == nil || !*resolvedDirect.Underline {
		t.Fatalf("direct child underline = %v, want true", resolvedDirect.Underline)
	}

	resolvedNested := sheet.Resolve(button, []Node{panel, container})
	if resolvedNested.Bold != nil {
		t.Fatalf("nested bold = %v, want nil", resolvedNested.Bold)
	}
	if resolvedNested.Underline == nil || !*resolvedNested.Underline {
		t.Fatalf("nested underline = %v, want true", resolvedNested.Underline)
	}
}
