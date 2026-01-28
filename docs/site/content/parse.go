package content

import (
	"strings"

	"github.com/odvcencio/fluffyui/markdown"
	"github.com/yuin/goldmark/ast"
)

func extractHeadings(root ast.Node, source []byte) []Heading {
	if root == nil {
		return nil
	}
	var headings []Heading
	seen := map[string]int{}
	_ = markdown.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := node.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		text := strings.TrimSpace(collectText(h, source))
		if text == "" {
			return ast.WalkContinue, nil
		}
		slug := headingID(h, text)
		if slug == "" {
			slug = "section"
		}
		seen[slug]++
		if seen[slug] > 1 {
			slug = slug + "-" + itoa(seen[slug])
		}
		headings = append(headings, Heading{
			Level: h.Level,
			Text:  text,
			ID:    slug,
		})
		return ast.WalkContinue, nil
	})
	return headings
}

func headingID(h *ast.Heading, fallback string) string {
	if h != nil {
		if value, ok := h.AttributeString("id"); ok {
			switch v := value.(type) {
			case string:
				if v != "" {
					return v
				}
			case []byte:
				if len(v) > 0 {
					return string(v)
				}
			}
		}
	}
	return slugify(fallback)
}

func firstParagraphText(root ast.Node, source []byte) string {
	if root == nil {
		return ""
	}
	summary := ""
	_ = markdown.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if para, ok := node.(*ast.Paragraph); ok {
			text := strings.TrimSpace(normalizeWhitespace(collectText(para, source)))
			if text != "" {
				summary = trimToLength(text, 200)
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	})
	return summary
}

func collectText(node ast.Node, source []byte) string {
	if node == nil {
		return ""
	}
	var sb strings.Builder
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		if n == nil {
			return
		}
		if text, ok := n.(*ast.Text); ok {
			sb.Write(text.Segment.Value(source))
			return
		}
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			walk(child)
		}
	}
	walk(node)
	return sb.String()
}

func plainTextFromLines(lines []markdown.StyledLine) string {
	if len(lines) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, line := range lines {
		if line.BlankLine {
			if i > 0 {
				sb.WriteString("\n")
			}
			continue
		}
		for _, span := range line.Prefix {
			sb.WriteString(span.Text)
		}
		for _, span := range line.Spans {
			sb.WriteString(span.Text)
		}
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}
	return normalizeWhitespace(sb.String())
}

func normalizeWhitespace(text string) string {
	parts := strings.Fields(text)
	return strings.Join(parts, " ")
}

func trimToLength(text string, max int) string {
	text = strings.TrimSpace(text)
	if max <= 0 || len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}

func slugify(text string) string {
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return ""
	}
	var out []rune
	wasDash := false
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out = append(out, r)
			wasDash = false
			continue
		}
		if r == '-' || r == '_' || r == ' ' || r == '.' || r == '/' {
			if !wasDash && len(out) > 0 {
				out = append(out, '-')
				wasDash = true
			}
		}
	}
	result := strings.Trim(string(out), "-")
	return result
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var out [20]byte
	pos := len(out)
	for v > 0 {
		pos--
		out[pos] = byte('0' + (v % 10))
		v /= 10
	}
	return string(out[pos:])
}
