package style

import "testing"

func TestDarkThemeNotNil(t *testing.T) {
	s := DarkTheme()
	if s == nil {
		t.Fatal("DarkTheme() returned nil")
	}
}

func TestLightThemeNotNil(t *testing.T) {
	s := LightTheme()
	if s == nil {
		t.Fatal("LightTheme() returned nil")
	}
}

func TestMonokaiThemeNotNil(t *testing.T) {
	s := MonokaiTheme()
	if s == nil {
		t.Fatal("MonokaiTheme() returned nil")
	}
}

func TestNordThemeNotNil(t *testing.T) {
	s := NordTheme()
	if s == nil {
		t.Fatal("NordTheme() returned nil")
	}
}

func TestDarkThemeResolvesButton(t *testing.T) {
	s := DarkTheme()
	node := &testNode{typ: "Button"}
	resolved := s.Resolve(node, nil)
	if resolved.IsZero() {
		t.Fatal("DarkTheme Button style resolved to zero")
	}
	if resolved.Bold == nil || !*resolved.Bold {
		t.Fatal("DarkTheme Button should be bold")
	}
}

func TestLightThemeResolvesButton(t *testing.T) {
	s := LightTheme()
	node := &testNode{typ: "Button"}
	resolved := s.Resolve(node, nil)
	if resolved.IsZero() {
		t.Fatal("LightTheme Button style resolved to zero")
	}
	if resolved.Bold == nil || !*resolved.Bold {
		t.Fatal("LightTheme Button should be bold")
	}
}

func TestMonokaiThemeResolvesButton(t *testing.T) {
	s := MonokaiTheme()
	node := &testNode{typ: "Button"}
	resolved := s.Resolve(node, nil)
	if resolved.IsZero() {
		t.Fatal("MonokaiTheme Button style resolved to zero")
	}
}

func TestNordThemeResolvesButton(t *testing.T) {
	s := NordTheme()
	node := &testNode{typ: "Button"}
	resolved := s.Resolve(node, nil)
	if resolved.IsZero() {
		t.Fatal("NordTheme Button style resolved to zero")
	}
}

func TestDarkAndLightDiffer(t *testing.T) {
	dark := DarkTheme()
	light := LightTheme()

	node := &testNode{typ: "Button"}
	darkBtn := dark.Resolve(node, nil)
	lightBtn := light.Resolve(node, nil)

	// Dark has black fg on white bg; Light has white fg on blue bg.
	// At minimum the foreground or background must differ.
	fgSame := darkBtn.Foreground == lightBtn.Foreground
	bgSame := darkBtn.Background == lightBtn.Background
	if fgSame && bgSame {
		t.Fatalf("DarkTheme and LightTheme Button styles should differ; dark=%+v light=%+v", darkBtn, lightBtn)
	}
}

func TestThemeResolvesWidgetTypes(t *testing.T) {
	themes := map[string]*Stylesheet{
		"Dark":    DarkTheme(),
		"Light":   LightTheme(),
		"Monokai": MonokaiTheme(),
		"Nord":    NordTheme(),
	}

	widgetTypes := []string{
		"Button", "Input", "Panel", "Dialog",
		"Tabs", "Alert", "Table", "Progress",
		"Label", "Text",
	}

	for name, sheet := range themes {
		for _, wt := range widgetTypes {
			node := &testNode{typ: wt}
			resolved := sheet.Resolve(node, nil)
			if resolved.IsZero() {
				t.Errorf("%s theme: %s resolved to zero style", name, wt)
			}
		}
	}
}

func TestThemeAlertClasses(t *testing.T) {
	themes := map[string]*Stylesheet{
		"Dark":    DarkTheme(),
		"Light":   LightTheme(),
		"Monokai": MonokaiTheme(),
		"Nord":    NordTheme(),
	}

	classes := []string{"info", "warning", "error"}

	for name, sheet := range themes {
		for _, cls := range classes {
			node := &testNode{typ: "Alert", classes: []string{cls}}
			resolved := sheet.Resolve(node, nil)
			if resolved.IsZero() {
				t.Errorf("%s theme: Alert.%s resolved to zero style", name, cls)
			}
		}
	}
}

func TestThemeButtonFocus(t *testing.T) {
	themes := map[string]*Stylesheet{
		"Dark":    DarkTheme(),
		"Light":   LightTheme(),
		"Monokai": MonokaiTheme(),
		"Nord":    NordTheme(),
	}

	for name, sheet := range themes {
		normal := &testNode{typ: "Button"}
		focused := &testNode{typ: "Button", state: WidgetState{Focused: true}}

		normalStyle := sheet.Resolve(normal, nil)
		focusStyle := sheet.Resolve(focused, nil)

		// Focus should change at least the background
		if normalStyle.Background == focusStyle.Background {
			t.Errorf("%s theme: Button and Button:focus should have different backgrounds", name)
		}
	}
}

func TestThemeInputFocus(t *testing.T) {
	themes := map[string]*Stylesheet{
		"Dark":    DarkTheme(),
		"Light":   LightTheme(),
		"Monokai": MonokaiTheme(),
		"Nord":    NordTheme(),
	}

	for name, sheet := range themes {
		normal := &testNode{typ: "Input"}
		focused := &testNode{typ: "Input", state: WidgetState{Focused: true}}

		normalStyle := sheet.Resolve(normal, nil)
		focusStyle := sheet.Resolve(focused, nil)

		// Focus should change the border color
		if normalStyle.Border == nil {
			t.Errorf("%s theme: Input should have a border", name)
			continue
		}
		if focusStyle.Border == nil {
			t.Errorf("%s theme: Input:focus should have a border", name)
			continue
		}
		if normalStyle.Border.Color == focusStyle.Border.Color {
			t.Errorf("%s theme: Input and Input:focus should have different border colors", name)
		}
	}
}

func TestThemeVariables(t *testing.T) {
	tests := []struct {
		name     string
		sheet    *Stylesheet
		wantName string
	}{
		{"Dark", DarkTheme(), "dark"},
		{"Light", LightTheme(), "light"},
		{"Monokai", MonokaiTheme(), "monokai"},
		{"Nord", NordTheme(), "nord"},
	}

	for _, tt := range tests {
		v, ok := tt.sheet.GetVariable("theme-name")
		if !ok {
			t.Errorf("%s theme: missing theme-name variable", tt.name)
			continue
		}
		if v != tt.wantName {
			t.Errorf("%s theme: theme-name = %q, want %q", tt.name, v, tt.wantName)
		}
	}
}

func TestThemeMerge(t *testing.T) {
	base := DarkTheme()
	override := NewStylesheet()
	override.Add(Select("Button"), Style{
		Foreground: ColorRed,
	})

	merged := Merge(base, override)
	node := &testNode{typ: "Button"}
	resolved := merged.Resolve(node, nil)

	// The override should win for foreground
	if resolved.Foreground != ColorRed {
		t.Fatalf("merged Button foreground = %v, want ColorRed", resolved.Foreground)
	}
	// But base bold should still be inherited from first rule
	if resolved.Bold == nil || !*resolved.Bold {
		t.Fatal("merged Button should still be bold from base theme")
	}
}
