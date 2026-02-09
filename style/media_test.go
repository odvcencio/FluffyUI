package style

import "testing"

// ---------------------------------------------------------------------------
// prefers-color-scheme parsing
// ---------------------------------------------------------------------------

func TestParseMediaQueryColorSchemeDark(t *testing.T) {
	queries, err := parseMediaQueries("(prefers-color-scheme: dark)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(queries))
	}
	q := queries[0]
	if q.ColorScheme == nil || *q.ColorScheme != "dark" {
		t.Fatalf("ColorScheme = %v, want *dark", q.ColorScheme)
	}
}

func TestParseMediaQueryColorSchemeLight(t *testing.T) {
	queries, err := parseMediaQueries("(prefers-color-scheme: light)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(queries))
	}
	q := queries[0]
	if q.ColorScheme == nil || *q.ColorScheme != "light" {
		t.Fatalf("ColorScheme = %v, want *light", q.ColorScheme)
	}
}

func TestParseMediaQueryColorSchemeInvalid(t *testing.T) {
	_, err := parseMediaQueries("(prefers-color-scheme: sepia)")
	if err == nil {
		t.Fatal("expected error for invalid prefers-color-scheme value")
	}
}

func TestParseMediaQueryColorSchemeUpperCase(t *testing.T) {
	queries, err := parseMediaQueries("(prefers-color-scheme: Dark)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queries) != 1 || queries[0].ColorScheme == nil || *queries[0].ColorScheme != "dark" {
		t.Fatalf("expected dark, got %v", queries[0].ColorScheme)
	}
}

// ---------------------------------------------------------------------------
// prefers-color-scheme matching
// ---------------------------------------------------------------------------

func TestMediaQueryMatchesColorSchemeDark(t *testing.T) {
	dark := "dark"
	q := MediaQuery{ColorScheme: &dark}

	if !q.Matches(MediaContext{ColorScheme: "dark"}) {
		t.Fatal("expected dark query to match dark context")
	}
	if q.Matches(MediaContext{ColorScheme: "light"}) {
		t.Fatal("expected dark query NOT to match light context")
	}
}

func TestMediaQueryMatchesColorSchemeLight(t *testing.T) {
	light := "light"
	q := MediaQuery{ColorScheme: &light}

	if !q.Matches(MediaContext{ColorScheme: "light"}) {
		t.Fatal("expected light query to match light context")
	}
	if q.Matches(MediaContext{ColorScheme: "dark"}) {
		t.Fatal("expected light query NOT to match dark context")
	}
}

func TestMediaQueryColorSchemeDefaultDark(t *testing.T) {
	// Empty ColorScheme in context should be treated as "dark".
	dark := "dark"
	light := "light"

	qDark := MediaQuery{ColorScheme: &dark}
	qLight := MediaQuery{ColorScheme: &light}

	if !qDark.Matches(MediaContext{}) {
		t.Fatal("expected dark query to match empty context (defaults to dark)")
	}
	if qLight.Matches(MediaContext{}) {
		t.Fatal("expected light query NOT to match empty context")
	}
}

func TestMediaQueryNoColorSchemeAlwaysMatches(t *testing.T) {
	// A query without ColorScheme should match any context.
	q := MediaQuery{}
	if !q.Matches(MediaContext{ColorScheme: "light"}) {
		t.Fatal("expected query without ColorScheme to match light context")
	}
	if !q.Matches(MediaContext{ColorScheme: "dark"}) {
		t.Fatal("expected query without ColorScheme to match dark context")
	}
}

// ---------------------------------------------------------------------------
// Combined media queries with color scheme
// ---------------------------------------------------------------------------

func TestMediaQueryCombinedColorSchemeAndWidth(t *testing.T) {
	queries, err := parseMediaQueries("(min-width: 80) and (prefers-color-scheme: dark)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(queries))
	}
	q := queries[0]
	if q.MinWidth != 80 {
		t.Fatalf("MinWidth = %d, want 80", q.MinWidth)
	}
	if q.ColorScheme == nil || *q.ColorScheme != "dark" {
		t.Fatalf("ColorScheme = %v, want *dark", q.ColorScheme)
	}

	// Match: wide + dark
	if !q.Matches(MediaContext{Width: 100, ColorScheme: "dark"}) {
		t.Fatal("expected match for wide+dark")
	}
	// No match: wide + light
	if q.Matches(MediaContext{Width: 100, ColorScheme: "light"}) {
		t.Fatal("expected no match for wide+light")
	}
	// No match: narrow + dark
	if q.Matches(MediaContext{Width: 40, ColorScheme: "dark"}) {
		t.Fatal("expected no match for narrow+dark")
	}
}

func TestMediaQueryCommaColorScheme(t *testing.T) {
	// Comma-separated: either dark or light narrow
	queries, err := parseMediaQueries("(prefers-color-scheme: dark), (prefers-color-scheme: light) and (max-width: 40)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(queries) != 2 {
		t.Fatalf("expected 2 queries, got %d", len(queries))
	}

	// Dark context should match via first query
	if !mediaMatches(queries, MediaContext{Width: 100, ColorScheme: "dark"}) {
		t.Fatal("expected match for dark context via first query")
	}
	// Light narrow should match via second query
	if !mediaMatches(queries, MediaContext{Width: 30, ColorScheme: "light"}) {
		t.Fatal("expected match for light+narrow via second query")
	}
	// Light wide should NOT match either
	if mediaMatches(queries, MediaContext{Width: 100, ColorScheme: "light"}) {
		t.Fatal("expected no match for light+wide")
	}
}

// ---------------------------------------------------------------------------
// mergeMedia with ColorScheme
// ---------------------------------------------------------------------------

func TestMergeMediaColorSchemeAgreement(t *testing.T) {
	dark := "dark"
	a := MediaQuery{ColorScheme: &dark}
	b := MediaQuery{ColorScheme: &dark}
	merged := mergeMedia(a, b)
	if merged.Invalid {
		t.Fatal("expected merged to be valid")
	}
	if merged.ColorScheme == nil || *merged.ColorScheme != "dark" {
		t.Fatalf("ColorScheme = %v, want *dark", merged.ColorScheme)
	}
}

func TestMergeMediaColorSchemeConflict(t *testing.T) {
	dark := "dark"
	light := "light"
	a := MediaQuery{ColorScheme: &dark}
	b := MediaQuery{ColorScheme: &light}
	merged := mergeMedia(a, b)
	if !merged.Invalid {
		t.Fatal("expected merged to be invalid due to conflicting color schemes")
	}
}

func TestMergeMediaColorSchemeOneNil(t *testing.T) {
	dark := "dark"
	a := MediaQuery{}
	b := MediaQuery{ColorScheme: &dark}
	merged := mergeMedia(a, b)
	if merged.Invalid {
		t.Fatal("expected merged to be valid")
	}
	if merged.ColorScheme == nil || *merged.ColorScheme != "dark" {
		t.Fatalf("ColorScheme = %v, want *dark", merged.ColorScheme)
	}
}

// ---------------------------------------------------------------------------
// Full FSS parse integration
// ---------------------------------------------------------------------------

func TestFSSParseColorScheme(t *testing.T) {
	fss := `
@media (prefers-color-scheme: dark) {
	Button {
		foreground: white;
	}
}
@media (prefers-color-scheme: light) {
	Button {
		foreground: black;
	}
}
`
	sheet, err := Parse(fss)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	node := &testNode{typ: "Button"}

	darkStyle := sheet.ResolveWithContext(node, nil, MediaContext{ColorScheme: "dark"})
	if darkStyle.Foreground != ColorWhite {
		t.Fatalf("dark: foreground = %v, want white", darkStyle.Foreground)
	}

	lightStyle := sheet.ResolveWithContext(node, nil, MediaContext{ColorScheme: "light"})
	if lightStyle.Foreground != ColorBlack {
		t.Fatalf("light: foreground = %v, want black", lightStyle.Foreground)
	}
}

// ---------------------------------------------------------------------------
// DetectColorScheme
// ---------------------------------------------------------------------------

func TestDetectColorSchemeFromEnvDark(t *testing.T) {
	tests := []struct {
		name       string
		colorfgbg  string
		nocolor    string
		wantScheme string
	}{
		{"empty env defaults to dark", "", "", "dark"},
		{"COLORFGBG bg=0 is dark", "15;0", "", "dark"},
		{"COLORFGBG bg=1 is dark", "15;1", "", "dark"},
		{"COLORFGBG bg=7 is dark", "15;7", "", "dark"},
		{"COLORFGBG bg=8 is light", "0;8", "", "light"},
		{"COLORFGBG bg=15 is light", "0;15", "", "light"},
		{"COLORFGBG three-part bg=0 is dark", "15;0;0", "", "dark"},
		{"COLORFGBG three-part bg=15 is light", "15;0;15", "", "light"},
		{"NO_COLOR set means light", "", "1", "light"},
		{"COLORFGBG takes precedence over NO_COLOR", "15;0", "1", "dark"},
		{"COLORFGBG non-numeric ignored, NO_COLOR fallback", "bad", "1", "light"},
		{"COLORFGBG non-numeric ignored, no NO_COLOR defaults dark", "bad", "", "dark"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectColorSchemeFromEnv(tt.colorfgbg, tt.nocolor)
			if got != tt.wantScheme {
				t.Fatalf("detectColorSchemeFromEnv(%q, %q) = %q, want %q",
					tt.colorfgbg, tt.nocolor, got, tt.wantScheme)
			}
		})
	}
}
