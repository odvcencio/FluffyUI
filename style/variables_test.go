package style

import "testing"

func TestVariablesResolve(t *testing.T) {
	data := `
:root {
  --primary: #ff0000;
  --spacing: 2;
}

Button {
  foreground: var(--primary);
  padding: var(--spacing);
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Foreground != RGB(255, 0, 0) {
		t.Fatalf("foreground = %#v, want #ff0000", resolved.Foreground)
	}
	if resolved.Padding == nil || resolved.Padding.Top != 2 || resolved.Padding.Right != 2 {
		t.Fatalf("padding = %#v, want 2", resolved.Padding)
	}
}

func TestVariablesFallback(t *testing.T) {
	data := `Button { foreground: var(--missing, #00ff00); }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Foreground != RGB(0, 255, 0) {
		t.Fatalf("foreground = %#v, want #00ff00", resolved.Foreground)
	}
}
