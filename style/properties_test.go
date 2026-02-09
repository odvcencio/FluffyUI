package style

import "testing"

// --- text-align ---

func TestParseTextAlign(t *testing.T) {
	tests := []struct {
		input string
		want  TextAlign
	}{
		{"left", TextAlignLeft},
		{"center", TextAlignCenter},
		{"right", TextAlignRight},
		{"LEFT", TextAlignLeft},
		{"Center", TextAlignCenter},
	}
	for _, tt := range tests {
		got, err := parseTextAlign(tt.input)
		if err != nil {
			t.Fatalf("parseTextAlign(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseTextAlign(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseTextAlignInvalid(t *testing.T) {
	_, err := parseTextAlign("justify")
	if err == nil {
		t.Fatalf("expected error for invalid text-align")
	}
}

func TestParseTextAlignStylesheet(t *testing.T) {
	data := `Button { text-align: center; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.TextAlign != TextAlignCenter {
		t.Fatalf("text-align = %d, want %d (center)", resolved.TextAlign, TextAlignCenter)
	}
}

func TestTextAlignMerge(t *testing.T) {
	base := Style{TextAlign: TextAlignLeft}
	override := Style{TextAlign: TextAlignRight}
	merged := base.Merge(override)
	if merged.TextAlign != TextAlignRight {
		t.Fatalf("merged text-align = %d, want %d (right)", merged.TextAlign, TextAlignRight)
	}
}

func TestTextAlignMergeNoOverride(t *testing.T) {
	base := Style{TextAlign: TextAlignLeft}
	override := Style{}
	merged := base.Merge(override)
	if merged.TextAlign != TextAlignLeft {
		t.Fatalf("merged text-align = %d, want %d (left)", merged.TextAlign, TextAlignLeft)
	}
}

func TestTextAlignInherit(t *testing.T) {
	child := Style{}
	parent := Style{TextAlign: TextAlignCenter}
	inherited := child.Inherit(parent)
	if inherited.TextAlign != TextAlignCenter {
		t.Fatalf("inherited text-align = %d, want %d (center)", inherited.TextAlign, TextAlignCenter)
	}
}

// --- display ---

func TestParseDisplay(t *testing.T) {
	tests := []struct {
		input string
		want  Display
	}{
		{"block", DisplayBlock},
		{"flex", DisplayFlex},
		{"none", DisplayHidden},
		{"BLOCK", DisplayBlock},
	}
	for _, tt := range tests {
		got, err := parseDisplay(tt.input)
		if err != nil {
			t.Fatalf("parseDisplay(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseDisplay(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseDisplayInvalid(t *testing.T) {
	_, err := parseDisplay("inline")
	if err == nil {
		t.Fatalf("expected error for invalid display")
	}
}

func TestParseDisplayStylesheet(t *testing.T) {
	data := `Button { display: none; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Display != DisplayHidden {
		t.Fatalf("display = %d, want %d (hidden)", resolved.Display, DisplayHidden)
	}
}

func TestDisplayMerge(t *testing.T) {
	base := Style{Display: DisplayBlock}
	override := Style{Display: DisplayFlex}
	merged := base.Merge(override)
	if merged.Display != DisplayFlex {
		t.Fatalf("merged display = %d, want %d (flex)", merged.Display, DisplayFlex)
	}
}

func TestDisplayNotInherited(t *testing.T) {
	child := Style{}
	parent := Style{Display: DisplayFlex}
	inherited := child.Inherit(parent)
	if inherited.Display != DisplayNone {
		t.Fatalf("display should not inherit, got %d", inherited.Display)
	}
}

// --- visibility ---

func TestParseVisibility(t *testing.T) {
	tests := []struct {
		input string
		want  Visibility
	}{
		{"visible", VisibilityVisible},
		{"hidden", VisibilityHidden},
		{"VISIBLE", VisibilityVisible},
	}
	for _, tt := range tests {
		got, err := parseVisibility(tt.input)
		if err != nil {
			t.Fatalf("parseVisibility(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseVisibility(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseVisibilityInvalid(t *testing.T) {
	_, err := parseVisibility("collapse")
	if err == nil {
		t.Fatalf("expected error for invalid visibility")
	}
}

func TestParseVisibilityStylesheet(t *testing.T) {
	data := `Button { visibility: hidden; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Visibility != VisibilityHidden {
		t.Fatalf("visibility = %d, want %d (hidden)", resolved.Visibility, VisibilityHidden)
	}
}

func TestVisibilityMerge(t *testing.T) {
	base := Style{Visibility: VisibilityVisible}
	override := Style{Visibility: VisibilityHidden}
	merged := base.Merge(override)
	if merged.Visibility != VisibilityHidden {
		t.Fatalf("merged visibility = %d, want %d (hidden)", merged.Visibility, VisibilityHidden)
	}
}

func TestVisibilityInherit(t *testing.T) {
	child := Style{}
	parent := Style{Visibility: VisibilityHidden}
	inherited := child.Inherit(parent)
	if inherited.Visibility != VisibilityHidden {
		t.Fatalf("inherited visibility = %d, want %d (hidden)", inherited.Visibility, VisibilityHidden)
	}
}

// --- overflow ---

func TestParseOverflow(t *testing.T) {
	tests := []struct {
		input string
		want  Overflow
	}{
		{"visible", OverflowVisible},
		{"hidden", OverflowHidden},
		{"scroll", OverflowScroll},
		{"HIDDEN", OverflowHidden},
	}
	for _, tt := range tests {
		got, err := parseOverflow(tt.input)
		if err != nil {
			t.Fatalf("parseOverflow(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseOverflow(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseOverflowInvalid(t *testing.T) {
	_, err := parseOverflow("auto")
	if err == nil {
		t.Fatalf("expected error for invalid overflow")
	}
}

func TestParseOverflowStylesheet(t *testing.T) {
	data := `Panel { overflow: scroll; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.Overflow != OverflowScroll {
		t.Fatalf("overflow = %d, want %d (scroll)", resolved.Overflow, OverflowScroll)
	}
}

func TestOverflowMerge(t *testing.T) {
	base := Style{Overflow: OverflowVisible}
	override := Style{Overflow: OverflowHidden}
	merged := base.Merge(override)
	if merged.Overflow != OverflowHidden {
		t.Fatalf("merged overflow = %d, want %d (hidden)", merged.Overflow, OverflowHidden)
	}
}

func TestOverflowNotInherited(t *testing.T) {
	child := Style{}
	parent := Style{Overflow: OverflowScroll}
	inherited := child.Inherit(parent)
	if inherited.Overflow != OverflowNone {
		t.Fatalf("overflow should not inherit, got %d", inherited.Overflow)
	}
}

// --- opacity ---

func TestParseOpacity(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"0", 0},
		{"0.0", 0},
		{"0.5", 0.5},
		{"1", 1},
		{"1.0", 1},
		{"0.75", 0.75},
	}
	for _, tt := range tests {
		got, err := parseOpacity(tt.input)
		if err != nil {
			t.Fatalf("parseOpacity(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseOpacity(%q) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestParseOpacityOutOfRange(t *testing.T) {
	_, err := parseOpacity("1.5")
	if err == nil {
		t.Fatalf("expected error for opacity > 1")
	}
	_, err = parseOpacity("-0.1")
	if err == nil {
		t.Fatalf("expected error for opacity < 0")
	}
}

func TestParseOpacityInvalid(t *testing.T) {
	_, err := parseOpacity("abc")
	if err == nil {
		t.Fatalf("expected error for non-numeric opacity")
	}
}

func TestParseOpacityStylesheet(t *testing.T) {
	data := `Button { opacity: 0.5; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Opacity == nil || *resolved.Opacity != 0.5 {
		t.Fatalf("opacity = %v, want 0.5", resolved.Opacity)
	}
}

func TestOpacityMerge(t *testing.T) {
	v1 := 0.5
	v2 := 0.8
	base := Style{Opacity: &v1}
	override := Style{Opacity: &v2}
	merged := base.Merge(override)
	if merged.Opacity == nil || *merged.Opacity != 0.8 {
		t.Fatalf("merged opacity = %v, want 0.8", merged.Opacity)
	}
}

func TestOpacityMergeNoOverride(t *testing.T) {
	v := 0.5
	base := Style{Opacity: &v}
	override := Style{}
	merged := base.Merge(override)
	if merged.Opacity == nil || *merged.Opacity != 0.5 {
		t.Fatalf("merged opacity = %v, want 0.5", merged.Opacity)
	}
}

func TestOpacityNotInherited(t *testing.T) {
	v := 0.5
	child := Style{}
	parent := Style{Opacity: &v}
	inherited := child.Inherit(parent)
	if inherited.Opacity != nil {
		t.Fatalf("opacity should not inherit, got %v", inherited.Opacity)
	}
}

// --- cursor ---

func TestParseCursor(t *testing.T) {
	tests := []struct {
		input string
		want  CursorStyle
	}{
		{"default", CursorDefault},
		{"pointer", CursorPointer},
		{"text", CursorText},
		{"POINTER", CursorPointer},
	}
	for _, tt := range tests {
		got, err := parseCursor(tt.input)
		if err != nil {
			t.Fatalf("parseCursor(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseCursor(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseCursorInvalid(t *testing.T) {
	_, err := parseCursor("crosshair")
	if err == nil {
		t.Fatalf("expected error for invalid cursor")
	}
}

func TestParseCursorStylesheet(t *testing.T) {
	data := `Button { cursor: pointer; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Cursor != CursorPointer {
		t.Fatalf("cursor = %d, want %d (pointer)", resolved.Cursor, CursorPointer)
	}
}

func TestCursorMerge(t *testing.T) {
	base := Style{Cursor: CursorDefault}
	override := Style{Cursor: CursorPointer}
	merged := base.Merge(override)
	if merged.Cursor != CursorPointer {
		t.Fatalf("merged cursor = %d, want %d (pointer)", merged.Cursor, CursorPointer)
	}
}

func TestCursorInherit(t *testing.T) {
	child := Style{}
	parent := Style{Cursor: CursorPointer}
	inherited := child.Inherit(parent)
	if inherited.Cursor != CursorPointer {
		t.Fatalf("inherited cursor = %d, want %d (pointer)", inherited.Cursor, CursorPointer)
	}
}

// --- text-decoration ---

func TestParseTextDecoration(t *testing.T) {
	tests := []struct {
		input string
		want  TextDecorationStyle
	}{
		{"none", TextDecorationOff},
		{"underline", TextDecorationUnderline},
		{"strikethrough", TextDecorationStrikethrough},
		{"line-through", TextDecorationStrikethrough},
		{"UNDERLINE", TextDecorationUnderline},
	}
	for _, tt := range tests {
		got, err := parseTextDecoration(tt.input)
		if err != nil {
			t.Fatalf("parseTextDecoration(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseTextDecoration(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseTextDecorationInvalid(t *testing.T) {
	_, err := parseTextDecoration("overline")
	if err == nil {
		t.Fatalf("expected error for invalid text-decoration")
	}
}

func TestParseTextDecorationStylesheet(t *testing.T) {
	data := `Link { text-decoration: underline; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Link"}
	resolved := sheet.Resolve(node, nil)
	if resolved.TextDecoration != TextDecorationUnderline {
		t.Fatalf("text-decoration = %d, want %d (underline)", resolved.TextDecoration, TextDecorationUnderline)
	}
}

func TestTextDecorationMerge(t *testing.T) {
	base := Style{TextDecoration: TextDecorationUnderline}
	override := Style{TextDecoration: TextDecorationOff}
	merged := base.Merge(override)
	if merged.TextDecoration != TextDecorationOff {
		t.Fatalf("merged text-decoration = %d, want %d (off)", merged.TextDecoration, TextDecorationOff)
	}
}

func TestTextDecorationInherit(t *testing.T) {
	child := Style{}
	parent := Style{TextDecoration: TextDecorationUnderline}
	inherited := child.Inherit(parent)
	if inherited.TextDecoration != TextDecorationUnderline {
		t.Fatalf("inherited text-decoration = %d, want %d (underline)", inherited.TextDecoration, TextDecorationUnderline)
	}
}

// --- white-space ---

func TestParseWhiteSpace(t *testing.T) {
	tests := []struct {
		input string
		want  WhiteSpace
	}{
		{"normal", WhiteSpaceNormal},
		{"nowrap", WhiteSpaceNowrap},
		{"pre", WhiteSpacePre},
		{"NOWRAP", WhiteSpaceNowrap},
	}
	for _, tt := range tests {
		got, err := parseWhiteSpace(tt.input)
		if err != nil {
			t.Fatalf("parseWhiteSpace(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("parseWhiteSpace(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseWhiteSpaceInvalid(t *testing.T) {
	_, err := parseWhiteSpace("break-spaces")
	if err == nil {
		t.Fatalf("expected error for invalid white-space")
	}
}

func TestParseWhiteSpaceStylesheet(t *testing.T) {
	data := `Label { white-space: nowrap; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Label"}
	resolved := sheet.Resolve(node, nil)
	if resolved.WhiteSpace != WhiteSpaceNowrap {
		t.Fatalf("white-space = %d, want %d (nowrap)", resolved.WhiteSpace, WhiteSpaceNowrap)
	}
}

func TestWhiteSpaceMerge(t *testing.T) {
	base := Style{WhiteSpace: WhiteSpaceNormal}
	override := Style{WhiteSpace: WhiteSpacePre}
	merged := base.Merge(override)
	if merged.WhiteSpace != WhiteSpacePre {
		t.Fatalf("merged white-space = %d, want %d (pre)", merged.WhiteSpace, WhiteSpacePre)
	}
}

func TestWhiteSpaceInherit(t *testing.T) {
	child := Style{}
	parent := Style{WhiteSpace: WhiteSpacePre}
	inherited := child.Inherit(parent)
	if inherited.WhiteSpace != WhiteSpacePre {
		t.Fatalf("inherited white-space = %d, want %d (pre)", inherited.WhiteSpace, WhiteSpacePre)
	}
}

// --- IsZero / AffectsLayout integration ---

func TestIsZeroWithNewProperties(t *testing.T) {
	s := Style{}
	if !s.IsZero() {
		t.Fatalf("empty style should be zero")
	}

	tests := []struct {
		name  string
		style Style
	}{
		{"text-align", Style{TextAlign: TextAlignCenter}},
		{"display", Style{Display: DisplayFlex}},
		{"visibility", Style{Visibility: VisibilityHidden}},
		{"overflow", Style{Overflow: OverflowScroll}},
		{"opacity", Style{Opacity: func() *float64 { v := 0.5; return &v }()}},
		{"cursor", Style{Cursor: CursorPointer}},
		{"text-decoration", Style{TextDecoration: TextDecorationUnderline}},
		{"white-space", Style{WhiteSpace: WhiteSpacePre}},
	}

	for _, tt := range tests {
		if tt.style.IsZero() {
			t.Fatalf("style with %s should not be zero", tt.name)
		}
	}
}

func TestAffectsLayoutWithNewProperties(t *testing.T) {
	tests := []struct {
		name   string
		style  Style
		layout bool
	}{
		{"display", Style{Display: DisplayFlex}, true},
		{"visibility", Style{Visibility: VisibilityHidden}, true},
		{"overflow", Style{Overflow: OverflowScroll}, true},
		{"white-space", Style{WhiteSpace: WhiteSpacePre}, true},
		{"text-align", Style{TextAlign: TextAlignCenter}, true},
		// These don't affect layout
		{"opacity", Style{Opacity: func() *float64 { v := 0.5; return &v }()}, false},
		{"cursor", Style{Cursor: CursorPointer}, false},
		{"text-decoration", Style{TextDecoration: TextDecorationUnderline}, false},
	}

	for _, tt := range tests {
		got := tt.style.AffectsLayout()
		if got != tt.layout {
			t.Fatalf("AffectsLayout for %s = %v, want %v", tt.name, got, tt.layout)
		}
	}
}

// --- Combined stylesheet test ---

func TestParseAllNewProperties(t *testing.T) {
	data := `
Button {
  text-align: center;
  display: flex;
  visibility: visible;
  overflow: hidden;
  opacity: 0.8;
  cursor: pointer;
  text-decoration: underline;
  white-space: nowrap;
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)

	if resolved.TextAlign != TextAlignCenter {
		t.Fatalf("text-align = %d, want center", resolved.TextAlign)
	}
	if resolved.Display != DisplayFlex {
		t.Fatalf("display = %d, want flex", resolved.Display)
	}
	if resolved.Visibility != VisibilityVisible {
		t.Fatalf("visibility = %d, want visible", resolved.Visibility)
	}
	if resolved.Overflow != OverflowHidden {
		t.Fatalf("overflow = %d, want hidden", resolved.Overflow)
	}
	if resolved.Opacity == nil || *resolved.Opacity != 0.8 {
		t.Fatalf("opacity = %v, want 0.8", resolved.Opacity)
	}
	if resolved.Cursor != CursorPointer {
		t.Fatalf("cursor = %d, want pointer", resolved.Cursor)
	}
	if resolved.TextDecoration != TextDecorationUnderline {
		t.Fatalf("text-decoration = %d, want underline", resolved.TextDecoration)
	}
	if resolved.WhiteSpace != WhiteSpaceNowrap {
		t.Fatalf("white-space = %d, want nowrap", resolved.WhiteSpace)
	}
}

// --- Important override test ---

func TestNewPropertiesImportant(t *testing.T) {
	data := `
Button.primary {
  text-align: right;
  cursor: text;
}
Button {
  text-align: center !important;
  cursor: pointer !important;
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button", classes: []string{"primary"}}
	resolved := sheet.Resolve(btn, nil)

	if resolved.TextAlign != TextAlignCenter {
		t.Fatalf("text-align = %d, want center (via !important)", resolved.TextAlign)
	}
	if resolved.Cursor != CursorPointer {
		t.Fatalf("cursor = %d, want pointer (via !important)", resolved.Cursor)
	}
}

// --- Variable resolution test ---

func TestNewPropertiesWithVariables(t *testing.T) {
	data := `
:root {
  --align: center;
  --opacity: 0.7;
}
Button {
  text-align: var(--align);
  opacity: var(--opacity);
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)

	if resolved.TextAlign != TextAlignCenter {
		t.Fatalf("text-align = %d, want center (via variable)", resolved.TextAlign)
	}
	if resolved.Opacity == nil || *resolved.Opacity != 0.7 {
		t.Fatalf("opacity = %v, want 0.7 (via variable)", resolved.Opacity)
	}
}

// --- Specificity resolution test ---

func TestNewPropertiesSpecificity(t *testing.T) {
	data := `
Button { text-align: left; display: block; }
Button.primary { text-align: center; }
#submit { text-align: right; }
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{
		typ:     "Button",
		id:      "submit",
		classes: []string{"primary"},
	}
	resolved := sheet.Resolve(node, nil)

	if resolved.TextAlign != TextAlignRight {
		t.Fatalf("text-align = %d, want right (id wins)", resolved.TextAlign)
	}
	if resolved.Display != DisplayBlock {
		t.Fatalf("display = %d, want block (inherited from Button rule)", resolved.Display)
	}
}

// --- Descendant selector test ---

func TestNewPropertiesDescendant(t *testing.T) {
	data := `Dialog Button { display: flex; overflow: scroll; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	dialog := &testNode{typ: "Dialog"}
	btn := &testNode{typ: "Button"}

	resolved := sheet.Resolve(btn, []Node{dialog})
	if resolved.Display != DisplayFlex {
		t.Fatalf("display = %d, want flex", resolved.Display)
	}
	if resolved.Overflow != OverflowScroll {
		t.Fatalf("overflow = %d, want scroll", resolved.Overflow)
	}

	// Without ancestor, should not match
	resolved = sheet.Resolve(btn, nil)
	if resolved.Display != DisplayNone {
		t.Fatalf("display = %d, want none (no ancestor)", resolved.Display)
	}
}

// --- Pseudo-class test ---

func TestNewPropertiesPseudo(t *testing.T) {
	data := `
Button { opacity: 1.0; }
Button:disabled { opacity: 0.5; cursor: default; }
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.Opacity == nil || *resolved.Opacity != 1.0 {
		t.Fatalf("opacity = %v, want 1.0", resolved.Opacity)
	}
	if resolved.Cursor != CursorNone {
		t.Fatalf("cursor = %d, want none", resolved.Cursor)
	}

	btn.state = WidgetState{Disabled: true}
	resolved = sheet.Resolve(btn, nil)
	if resolved.Opacity == nil || *resolved.Opacity != 0.5 {
		t.Fatalf("opacity = %v, want 0.5 (disabled)", resolved.Opacity)
	}
	if resolved.Cursor != CursorDefault {
		t.Fatalf("cursor = %d, want default (disabled)", resolved.Cursor)
	}
}

// --- display: none in stylesheet with display: block ---

func TestDisplayBlockAndFlex(t *testing.T) {
	data := `
Panel { display: block; }
Panel.hidden { display: none; }
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.Display != DisplayBlock {
		t.Fatalf("display = %d, want block", resolved.Display)
	}

	node.classes = []string{"hidden"}
	resolved = sheet.Resolve(node, nil)
	if resolved.Display != DisplayHidden {
		t.Fatalf("display = %d, want hidden (display:none)", resolved.Display)
	}
}

// --- text-decoration: none resets ---

func TestTextDecorationNoneResets(t *testing.T) {
	data := `
Link { text-decoration: underline; }
Link.plain { text-decoration: none; }
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	node := &testNode{typ: "Link", classes: []string{"plain"}}
	resolved := sheet.Resolve(node, nil)
	if resolved.TextDecoration != TextDecorationOff {
		t.Fatalf("text-decoration = %d, want off (none)", resolved.TextDecoration)
	}
}

// --- text-decoration: line-through alias ---

func TestTextDecorationLineThroughAlias(t *testing.T) {
	data := `Label { text-decoration: line-through; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	node := &testNode{typ: "Label"}
	resolved := sheet.Resolve(node, nil)
	if resolved.TextDecoration != TextDecorationStrikethrough {
		t.Fatalf("text-decoration = %d, want strikethrough", resolved.TextDecoration)
	}
}

// --- min-width / max-width / min-height / max-height ---

func intPtr(n int) *int { return &n }

func TestParseMinWidth(t *testing.T) {
	data := `Button { min-width: 10; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.MinWidth == nil || *resolved.MinWidth != 10 {
		t.Fatalf("min-width = %v, want 10", resolved.MinWidth)
	}
}

func TestParseMaxWidth(t *testing.T) {
	data := `Button { max-width: 80; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if resolved.MaxWidth == nil || *resolved.MaxWidth != 80 {
		t.Fatalf("max-width = %v, want 80", resolved.MaxWidth)
	}
}

func TestParseMinHeight(t *testing.T) {
	data := `Panel { min-height: 5; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.MinHeight == nil || *resolved.MinHeight != 5 {
		t.Fatalf("min-height = %v, want 5", resolved.MinHeight)
	}
}

func TestParseMaxHeight(t *testing.T) {
	data := `Panel { max-height: 40; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.MaxHeight == nil || *resolved.MaxHeight != 40 {
		t.Fatalf("max-height = %v, want 40", resolved.MaxHeight)
	}
}

func TestMinMaxDimensionsMerge(t *testing.T) {
	base := Style{MinWidth: intPtr(10), MaxWidth: intPtr(80)}
	override := Style{MinWidth: intPtr(20)}
	merged := base.Merge(override)
	if merged.MinWidth == nil || *merged.MinWidth != 20 {
		t.Fatalf("merged min-width = %v, want 20", merged.MinWidth)
	}
	if merged.MaxWidth == nil || *merged.MaxWidth != 80 {
		t.Fatalf("merged max-width = %v, want 80 (from base)", merged.MaxWidth)
	}
}

func TestMinMaxDimensionsNotInherited(t *testing.T) {
	child := Style{}
	parent := Style{MinWidth: intPtr(10), MaxHeight: intPtr(40)}
	inherited := child.Inherit(parent)
	if inherited.MinWidth != nil {
		t.Fatalf("min-width should not inherit, got %v", inherited.MinWidth)
	}
	if inherited.MaxHeight != nil {
		t.Fatalf("max-height should not inherit, got %v", inherited.MaxHeight)
	}
}

func TestMinMaxDimensionsAffectLayout(t *testing.T) {
	tests := []struct {
		name  string
		style Style
	}{
		{"min-width", Style{MinWidth: intPtr(10)}},
		{"max-width", Style{MaxWidth: intPtr(80)}},
		{"min-height", Style{MinHeight: intPtr(5)}},
		{"max-height", Style{MaxHeight: intPtr(40)}},
	}
	for _, tt := range tests {
		if !tt.style.AffectsLayout() {
			t.Fatalf("%s should affect layout", tt.name)
		}
	}
}

func TestMinMaxDimensionsIsZero(t *testing.T) {
	tests := []struct {
		name  string
		style Style
	}{
		{"min-width", Style{MinWidth: intPtr(10)}},
		{"max-width", Style{MaxWidth: intPtr(80)}},
		{"min-height", Style{MinHeight: intPtr(5)}},
		{"max-height", Style{MaxHeight: intPtr(40)}},
	}
	for _, tt := range tests {
		if tt.style.IsZero() {
			t.Fatalf("style with %s should not be zero", tt.name)
		}
	}
}

func TestAllDimensionsCombined(t *testing.T) {
	data := `Panel { min-width: 10; max-width: 80; min-height: 3; max-height: 40; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.MinWidth == nil || *resolved.MinWidth != 10 {
		t.Fatalf("min-width = %v, want 10", resolved.MinWidth)
	}
	if resolved.MaxWidth == nil || *resolved.MaxWidth != 80 {
		t.Fatalf("max-width = %v, want 80", resolved.MaxWidth)
	}
	if resolved.MinHeight == nil || *resolved.MinHeight != 3 {
		t.Fatalf("min-height = %v, want 3", resolved.MinHeight)
	}
	if resolved.MaxHeight == nil || *resolved.MaxHeight != 40 {
		t.Fatalf("max-height = %v, want 40", resolved.MaxHeight)
	}
}

// --- content-align ---

func TestParseContentAlign(t *testing.T) {
	tests := []struct {
		input string
		want  ContentAlign
	}{
		{"start", ContentAlignStart},
		{"center", ContentAlignCenter},
		{"end", ContentAlignEnd},
		{"CENTER", ContentAlignCenter},
		{"Start", ContentAlignStart},
	}
	for _, tt := range tests {
		got := parseContentAlign(tt.input)
		if got != tt.want {
			t.Fatalf("parseContentAlign(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseContentAlignInvalid(t *testing.T) {
	got := parseContentAlign("justify")
	if got != ContentAlignNone {
		t.Fatalf("expected ContentAlignNone for invalid, got %d", got)
	}
}

func TestParseContentAlignStylesheet(t *testing.T) {
	data := `Panel { content-align: center; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.ContentAlign != ContentAlignCenter {
		t.Fatalf("content-align = %d, want center", resolved.ContentAlign)
	}
}

func TestContentAlignMerge(t *testing.T) {
	base := Style{ContentAlign: ContentAlignStart}
	override := Style{ContentAlign: ContentAlignEnd}
	merged := base.Merge(override)
	if merged.ContentAlign != ContentAlignEnd {
		t.Fatalf("merged content-align = %d, want end", merged.ContentAlign)
	}
}

func TestContentAlignMergeNoOverride(t *testing.T) {
	base := Style{ContentAlign: ContentAlignStart}
	override := Style{}
	merged := base.Merge(override)
	if merged.ContentAlign != ContentAlignStart {
		t.Fatalf("merged content-align = %d, want start", merged.ContentAlign)
	}
}

func TestContentAlignAffectsLayout(t *testing.T) {
	s := Style{ContentAlign: ContentAlignCenter}
	if !s.AffectsLayout() {
		t.Fatal("content-align should affect layout")
	}
}

func TestContentAlignIsZero(t *testing.T) {
	s := Style{ContentAlign: ContentAlignCenter}
	if s.IsZero() {
		t.Fatal("style with content-align should not be zero")
	}
}

// --- dock ---

func TestParseDock(t *testing.T) {
	tests := []struct {
		input string
		want  Dock
	}{
		{"top", DockTop},
		{"bottom", DockBottom},
		{"left", DockLeft},
		{"right", DockRight},
		{"TOP", DockTop},
		{"Bottom", DockBottom},
	}
	for _, tt := range tests {
		got := parseDock(tt.input)
		if got != tt.want {
			t.Fatalf("parseDock(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseDockInvalid(t *testing.T) {
	got := parseDock("center")
	if got != DockNone {
		t.Fatalf("expected DockNone for invalid, got %d", got)
	}
}

func TestParseDockStylesheet(t *testing.T) {
	data := `StatusBar { dock: bottom; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "StatusBar"}
	resolved := sheet.Resolve(node, nil)
	if resolved.Dock != DockBottom {
		t.Fatalf("dock = %d, want bottom", resolved.Dock)
	}
}

func TestDockMerge(t *testing.T) {
	base := Style{Dock: DockTop}
	override := Style{Dock: DockBottom}
	merged := base.Merge(override)
	if merged.Dock != DockBottom {
		t.Fatalf("merged dock = %d, want bottom", merged.Dock)
	}
}

func TestDockMergeNoOverride(t *testing.T) {
	base := Style{Dock: DockTop}
	override := Style{}
	merged := base.Merge(override)
	if merged.Dock != DockTop {
		t.Fatalf("merged dock = %d, want top", merged.Dock)
	}
}

func TestDockAffectsLayout(t *testing.T) {
	s := Style{Dock: DockLeft}
	if !s.AffectsLayout() {
		t.Fatal("dock should affect layout")
	}
}

func TestDockIsZero(t *testing.T) {
	s := Style{Dock: DockRight}
	if s.IsZero() {
		t.Fatal("style with dock should not be zero")
	}
}

// --- offset-x / offset-y ---

func TestParseOffsetX(t *testing.T) {
	data := `Panel { offset-x: 5; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.OffsetX == nil || *resolved.OffsetX != 5 {
		t.Fatalf("offset-x = %v, want 5", resolved.OffsetX)
	}
}

func TestParseOffsetY(t *testing.T) {
	data := `Panel { offset-y: -3; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.OffsetY == nil || *resolved.OffsetY != -3 {
		t.Fatalf("offset-y = %v, want -3", resolved.OffsetY)
	}
}

func TestOffsetMerge(t *testing.T) {
	base := Style{OffsetX: intPtr(5), OffsetY: intPtr(10)}
	override := Style{OffsetX: intPtr(2)}
	merged := base.Merge(override)
	if merged.OffsetX == nil || *merged.OffsetX != 2 {
		t.Fatalf("merged offset-x = %v, want 2", merged.OffsetX)
	}
	if merged.OffsetY == nil || *merged.OffsetY != 10 {
		t.Fatalf("merged offset-y = %v, want 10 (from base)", merged.OffsetY)
	}
}

func TestOffsetAffectsLayout(t *testing.T) {
	if !(Style{OffsetX: intPtr(1)}).AffectsLayout() {
		t.Fatal("offset-x should affect layout")
	}
	if !(Style{OffsetY: intPtr(1)}).AffectsLayout() {
		t.Fatal("offset-y should affect layout")
	}
}

func TestOffsetIsZero(t *testing.T) {
	if (Style{OffsetX: intPtr(0)}).IsZero() {
		t.Fatal("style with offset-x should not be zero")
	}
	if (Style{OffsetY: intPtr(0)}).IsZero() {
		t.Fatal("style with offset-y should not be zero")
	}
}

// --- z-index ---

func TestParseZIndex(t *testing.T) {
	data := `Dialog { z-index: 100; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Dialog"}
	resolved := sheet.Resolve(node, nil)
	if resolved.ZIndex == nil || *resolved.ZIndex != 100 {
		t.Fatalf("z-index = %v, want 100", resolved.ZIndex)
	}
}

func TestParseZIndexNegative(t *testing.T) {
	data := `Panel { z-index: -1; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)
	if resolved.ZIndex == nil || *resolved.ZIndex != -1 {
		t.Fatalf("z-index = %v, want -1", resolved.ZIndex)
	}
}

func TestZIndexMerge(t *testing.T) {
	base := Style{ZIndex: intPtr(10)}
	override := Style{ZIndex: intPtr(100)}
	merged := base.Merge(override)
	if merged.ZIndex == nil || *merged.ZIndex != 100 {
		t.Fatalf("merged z-index = %v, want 100", merged.ZIndex)
	}
}

func TestZIndexMergeNoOverride(t *testing.T) {
	base := Style{ZIndex: intPtr(10)}
	override := Style{}
	merged := base.Merge(override)
	if merged.ZIndex == nil || *merged.ZIndex != 10 {
		t.Fatalf("merged z-index = %v, want 10", merged.ZIndex)
	}
}

func TestZIndexAffectsLayout(t *testing.T) {
	s := Style{ZIndex: intPtr(10)}
	if !s.AffectsLayout() {
		t.Fatal("z-index should affect layout")
	}
}

func TestZIndexIsZero(t *testing.T) {
	s := Style{ZIndex: intPtr(0)}
	if s.IsZero() {
		t.Fatal("style with z-index should not be zero")
	}
}

// --- Combined layout properties stylesheet ---

func TestAllLayoutPropertiesCombined(t *testing.T) {
	data := `
Panel {
  min-width: 10;
  max-width: 80;
  min-height: 3;
  max-height: 40;
  content-align: center;
  dock: top;
  offset-x: 2;
  offset-y: -1;
  z-index: 50;
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)

	if resolved.MinWidth == nil || *resolved.MinWidth != 10 {
		t.Fatalf("min-width = %v, want 10", resolved.MinWidth)
	}
	if resolved.MaxWidth == nil || *resolved.MaxWidth != 80 {
		t.Fatalf("max-width = %v, want 80", resolved.MaxWidth)
	}
	if resolved.MinHeight == nil || *resolved.MinHeight != 3 {
		t.Fatalf("min-height = %v, want 3", resolved.MinHeight)
	}
	if resolved.MaxHeight == nil || *resolved.MaxHeight != 40 {
		t.Fatalf("max-height = %v, want 40", resolved.MaxHeight)
	}
	if resolved.ContentAlign != ContentAlignCenter {
		t.Fatalf("content-align = %d, want center", resolved.ContentAlign)
	}
	if resolved.Dock != DockTop {
		t.Fatalf("dock = %d, want top", resolved.Dock)
	}
	if resolved.OffsetX == nil || *resolved.OffsetX != 2 {
		t.Fatalf("offset-x = %v, want 2", resolved.OffsetX)
	}
	if resolved.OffsetY == nil || *resolved.OffsetY != -1 {
		t.Fatalf("offset-y = %v, want -1", resolved.OffsetY)
	}
	if resolved.ZIndex == nil || *resolved.ZIndex != 50 {
		t.Fatalf("z-index = %v, want 50", resolved.ZIndex)
	}
}

// --- Layout properties with specificity ---

func TestLayoutPropertiesSpecificity(t *testing.T) {
	data := `
Panel { dock: top; z-index: 1; }
Panel.sidebar { dock: left; }
#main { z-index: 100; }
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel", id: "main", classes: []string{"sidebar"}}
	resolved := sheet.Resolve(node, nil)

	if resolved.Dock != DockLeft {
		t.Fatalf("dock = %d, want left (class wins over type)", resolved.Dock)
	}
	if resolved.ZIndex == nil || *resolved.ZIndex != 100 {
		t.Fatalf("z-index = %v, want 100 (id wins)", resolved.ZIndex)
	}
}

// --- Layout properties with variables ---

func TestLayoutPropertiesWithVariables(t *testing.T) {
	data := `
:root {
  --sidebar-width: 30;
  --layer: 10;
}
Panel {
  min-width: var(--sidebar-width);
  z-index: var(--layer);
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	node := &testNode{typ: "Panel"}
	resolved := sheet.Resolve(node, nil)

	if resolved.MinWidth == nil || *resolved.MinWidth != 30 {
		t.Fatalf("min-width = %v, want 30 (via variable)", resolved.MinWidth)
	}
	if resolved.ZIndex == nil || *resolved.ZIndex != 10 {
		t.Fatalf("z-index = %v, want 10 (via variable)", resolved.ZIndex)
	}
}

// --- Media query combined with new properties ---

func TestNewPropertiesMediaQuery(t *testing.T) {
	data := `
Button { text-align: left; }
@media (min-width: 80) {
  Button { text-align: center; }
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	btn := &testNode{typ: "Button"}

	resolved := sheet.ResolveWithContext(btn, nil, MediaContext{Width: 60, Height: 24})
	if resolved.TextAlign != TextAlignLeft {
		t.Fatalf("text-align = %d, want left (narrow)", resolved.TextAlign)
	}

	resolved = sheet.ResolveWithContext(btn, nil, MediaContext{Width: 100, Height: 24})
	if resolved.TextAlign != TextAlignCenter {
		t.Fatalf("text-align = %d, want center (wide)", resolved.TextAlign)
	}
}
