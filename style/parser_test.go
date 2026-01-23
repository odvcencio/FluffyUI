package style

import "testing"

func TestParseStylesheet(t *testing.T) {
	data := `
/* comment */
Button {
  padding: 1;
  border: single;
}

Button.primary {
  foreground: #FFB74D;
  background: #0C0C10;
  bold: true;
}

Button:focus {
  border: double #ffb74d;
}

Dialog Input {
  padding: 2;
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	btn := &testNode{typ: "Button", classes: []string{"primary"}}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Padding == nil || resolved.Padding.Top != 1 || resolved.Padding.Right != 1 {
		t.Fatalf("padding = %#v, want 1", resolved.Padding)
	}
	if resolved.Foreground != RGB(255, 183, 77) {
		t.Fatalf("foreground = %#v, want #ffb74d", resolved.Foreground)
	}
	if resolved.Background != RGB(12, 12, 16) {
		t.Fatalf("background = %#v, want #0c0c10", resolved.Background)
	}
	if resolved.Bold == nil || !*resolved.Bold {
		t.Fatalf("bold = %v, want true", resolved.Bold)
	}

	btn.state = WidgetState{Focused: true}
	resolved = sheet.Resolve(btn, nil)
	if resolved.Border == nil || resolved.Border.Style != BorderDouble {
		t.Fatalf("border = %#v, want double", resolved.Border)
	}
	if resolved.Border.Color != RGB(255, 183, 77) {
		t.Fatalf("border color = %#v, want #ffb74d", resolved.Border.Color)
	}

	input := &testNode{typ: "Input"}
	dialog := &testNode{typ: "Dialog"}
	resolved = sheet.Resolve(input, []Node{dialog})
	if resolved.Padding == nil || resolved.Padding.Top != 2 {
		t.Fatalf("padding = %#v, want 2", resolved.Padding)
	}
}

func TestParseUnknownProperty(t *testing.T) {
	data := `Button { nope: 1; padding: 2; }`
	sheet, err := Parse(data)
	if err == nil {
		t.Fatalf("expected error for unknown property")
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Padding == nil || resolved.Padding.Top != 2 {
		t.Fatalf("padding = %#v, want 2", resolved.Padding)
	}
}

func TestParseImportantOverrides(t *testing.T) {
	data := `
Button.primary {
  foreground: #00ff00;
}

Button {
  foreground: #ff0000 !important;
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Button", classes: []string{"primary"}}
	resolved := sheet.Resolve(node, nil)
	if resolved.Foreground != RGB(255, 0, 0) {
		t.Fatalf("foreground = %#v, want #ff0000", resolved.Foreground)
	}
}

func TestParseBorderSubproperties(t *testing.T) {
	data := `
Button {
  border-style: single;
}

Button.primary {
  border-color: #ffb74d;
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Button", classes: []string{"primary"}}
	resolved := sheet.Resolve(node, nil)
	if resolved.Border == nil || resolved.Border.Style != BorderSingle {
		t.Fatalf("border style = %#v, want single", resolved.Border)
	}
	if resolved.Border.Color != RGB(255, 183, 77) {
		t.Fatalf("border color = %#v, want #ffb74d", resolved.Border.Color)
	}
}

func TestParseAttributeSelector(t *testing.T) {
	data := `Input[type="password"] { bold: true; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &attrNode{
		testNode: testNode{typ: "Input"},
		attrs: map[string]string{
			"type": "password",
		},
	}
	resolved := sheet.Resolve(node, nil)
	if resolved.Bold == nil || !*resolved.Bold {
		t.Fatalf("bold = %v, want true", resolved.Bold)
	}
}
