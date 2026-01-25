package fur

import (
	"strings"
	"testing"
)

func TestExportText(t *testing.T) {
	r := Text("Hello, World!")
	output := ExportText(r, 40)

	if output != "Hello, World!" {
		t.Errorf("expected plain text, got %q", output)
	}
}

func TestExportTextNil(t *testing.T) {
	output := ExportText(nil, 40)
	if output != "" {
		t.Errorf("expected empty string for nil, got %q", output)
	}
}

func TestExportTextMultiline(t *testing.T) {
	r := Text("Line1\nLine2\nLine3")
	output := ExportText(r, 40)

	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestExportHTML(t *testing.T) {
	r := Markup("[bold]Bold[/] text")
	output := ExportHTML(r, 40)

	if !strings.Contains(output, "<pre") {
		t.Error("expected pre tag in HTML output")
	}
	if !strings.Contains(output, "Bold") {
		t.Error("expected text content in HTML output")
	}
	if !strings.Contains(output, "</pre>") {
		t.Error("expected closing pre tag")
	}
}

func TestExportHTMLEscaping(t *testing.T) {
	r := Text("<script>alert('xss')</script>")
	output := ExportHTML(r, 80)

	if strings.Contains(output, "<script>") {
		t.Error("HTML should escape script tags")
	}
	if !strings.Contains(output, "&lt;script&gt;") {
		t.Error("expected escaped script tag")
	}
}

func TestExportHTMLNil(t *testing.T) {
	output := ExportHTML(nil, 40)
	if output != "" {
		t.Errorf("expected empty string for nil, got %q", output)
	}
}

func TestExportHTMLWithStyles(t *testing.T) {
	r := Markup("[red]Red[/] and [bold]Bold[/]")
	output := ExportHTML(r, 40)

	if !strings.Contains(output, "style=") {
		t.Error("expected inline styles in HTML output")
	}
}

func TestExportSVG(t *testing.T) {
	r := Text("SVG Test")
	output := ExportSVG(r, 40)

	if !strings.HasPrefix(output, "<svg") {
		t.Error("expected svg element")
	}
	if !strings.Contains(output, "SVG Test") {
		t.Error("expected text content")
	}
	if !strings.HasSuffix(output, "</svg>") {
		t.Error("expected closing svg tag")
	}
}

func TestExportSVGNil(t *testing.T) {
	output := ExportSVG(nil, 40)
	if output != "" {
		t.Errorf("expected empty string for nil, got %q", output)
	}
}

func TestExportSVGEscaping(t *testing.T) {
	r := Text("Test <>&\"")
	output := ExportSVG(r, 40)

	if strings.Contains(output, "Test <>&\"") {
		t.Error("SVG should escape special characters")
	}
}

func TestExportGeneric(t *testing.T) {
	r := Text("Test")

	textOut := Export(r, 40, ExportTextFormat)
	if textOut != "Test" {
		t.Errorf("text export failed: %q", textOut)
	}

	htmlOut := Export(r, 40, ExportHTMLFormat)
	if !strings.Contains(htmlOut, "<pre") {
		t.Error("HTML export should contain pre tag")
	}

	svgOut := Export(r, 40, ExportSVGFormat)
	if !strings.HasPrefix(svgOut, "<svg") {
		t.Error("SVG export should start with svg tag")
	}
}

func TestExportDefaultFormat(t *testing.T) {
	r := Text("Default")
	output := Export(r, 40, "unknown")

	// Unknown format should fall back to text
	if output != "Default" {
		t.Errorf("expected text fallback, got %q", output)
	}
}
