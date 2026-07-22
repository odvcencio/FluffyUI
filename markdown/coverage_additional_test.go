package markdown

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"m31labs.dev/fluffyui/compositor"
	"m31labs.dev/fluffyui/theme"
)

func TestParserWalk(t *testing.T) {
	p := NewParser()
	root := p.ParseString("# Title\n\nHello")
	count := 0
	err := Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			count++
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected nodes to be visited")
	}
}

func TestHighlighterFallback(t *testing.T) {
	h := NewHighlighter(nil)
	lines := h.Highlight("line1\nline2", "unknown", nil)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !lines[0].IsCode || !lines[1].IsCode {
		t.Fatalf("expected code lines")
	}
	blank := h.Highlight("", "unknown", nil)
	if len(blank) != 1 || !blank[0].BlankLine {
		t.Fatalf("expected blank code line")
	}
}

func TestFallbackCodeLines(t *testing.T) {
	cfg := DefaultStyleConfig(nil)
	lines := fallbackCodeLines("", "go", cfg)
	if len(lines) != 1 || !lines[0].BlankLine {
		t.Fatalf("expected blank line")
	}
	lines = fallbackCodeLines("a\n", "go", cfg)
	if len(lines) != 2 || !lines[1].BlankLine {
		t.Fatalf("expected trailing blank line")
	}
}

func TestEnhancedTableRender(t *testing.T) {
	p := NewParser()
	src := "| Name | Age |\n| --- | --- |\n| Ada | 30 |\n"
	root := p.ParseString(src)
	var table *extast.Table
	_ = Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if t, ok := node.(*extast.Table); ok {
				table = t
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	})
	if table == nil {
		t.Fatalf("expected table node")
	}

	cfg := DefaultTableRendererConfig(theme.DefaultTheme())
	renderer := NewEnhancedTableRenderer(cfg)
	lines := renderer.RenderTable(table, []byte(src), 40)
	if len(lines) == 0 {
		t.Fatalf("expected rendered table lines")
	}
	text := spansText(lines[0].Spans)
	if !strings.Contains(text, cfg.BoxDrawings.TopLeft) {
		t.Fatalf("expected top border, got %q", text)
	}
	for _, line := range lines {
		if width := runewidth.StringWidth(spansText(line.Spans)); width > 40 {
			t.Fatalf("table line width = %d, want <= 40: %q", width, spansText(line.Spans))
		}
	}
}

func TestRendererRenderWidth_ConstrainsWideTable(t *testing.T) {
	r := NewRenderer(theme.DefaultTheme())
	src := "| Name | Description |\n| --- | --- |\n| Ada | A deliberately long description that must fit |\n"
	lines := r.RenderWidth("assistant", src, 28)
	for _, line := range lines {
		if width := runewidth.StringWidth(spansText(line.Spans)); width > 28 {
			t.Fatalf("rendered line width = %d, want <= 28: %q", width, spansText(line.Spans))
		}
	}
}

func TestRendererRenderWidth_WrapsTableCellsWithoutLosingContent(t *testing.T) {
	r := NewRenderer(theme.DefaultTheme())
	src := "| Item | Description |\n| --- | --- |\n| Parser | A deliberately long description that must retain its final fallback marker |\n"
	lines := r.RenderWidth("assistant", src, 34)
	var rendered strings.Builder
	for _, line := range lines {
		text := spansText(line.Spans)
		if width := runewidth.StringWidth(text); width > 34 {
			t.Fatalf("rendered line width = %d, want <= 34: %q", width, text)
		}
		rendered.WriteString(text)
		rendered.WriteByte('\n')
	}
	if !strings.Contains(rendered.String(), "fallback") || !strings.Contains(rendered.String(), "marker") {
		t.Fatalf("wrapped table lost cell content:\n%s", rendered.String())
	}
}

func TestStyleConfigAndMerge(t *testing.T) {
	cfg := DefaultStyleConfig(theme.DefaultTheme())
	base := compositor.DefaultStyle().WithBold(true)
	next := cfg.WithBaseText(base)
	if next == nil || !next.Text.Bold || !next.Bold.Bold {
		t.Fatalf("expected base text to propagate")
	}
	inline := compositor.DefaultStyle().WithUnderline(true)
	merged := MergeStyle(next.Text, inline)
	if !merged.Underline || !merged.Bold {
		t.Fatalf("expected merged style to keep underline + bold")
	}
}

func TestTableStyles(t *testing.T) {
	compact := TableStyleCompact(nil)
	if compact.Padding != 0 {
		t.Fatalf("expected compact padding 0")
	}
	minimal := TableStyleMinimal(nil)
	if minimal.BoxDrawings.TopLeft != "┌" {
		t.Fatalf("expected minimal box drawing")
	}
	heavy := TableStyleHeavy(nil)
	if heavy.BoxDrawings.TopLeft != HeavyBoxDrawings.TopLeft {
		t.Fatalf("expected heavy box drawing")
	}
}
