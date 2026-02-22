package html

import (
	"fmt"
	"html/template"
	"strings"
)

func composePage(title, css, body, js string) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
        html, body { height: 100%%; font-family: 'Consolas', 'Courier New', monospace; }
        .fluffy-List { list-style: none; }
        .fluffy-List li { padding: 4px 8px; cursor: pointer; }
        .fluffy-List li.selected { font-weight: bold; }
        .fluffy-Divider { border: none; border-top: 1px solid; opacity: 0.3; margin: 4px 0; }
        .fluffy-VDivider { width: 1px; align-self: stretch; opacity: 0.3; }
        .fluffy-Search { width: 100%%; padding: 6px 8px; border: 1px solid; border-radius: 4px; font-family: inherit; font-size: inherit; }
        .fluffy-Chip { border: 1px solid transparent; border-radius: 12px; cursor: pointer; font-family: inherit; font-size: 0.85em; }
        .fluffy-Chip.active { font-weight: bold; }
        .fluffy-MarkdownViewer { line-height: 1.6; }
        .fluffy-MarkdownViewer h1 { font-size: 1.8em; margin: 0.5em 0; }
        .fluffy-MarkdownViewer h2 { font-size: 1.4em; margin: 0.5em 0; }
        .fluffy-MarkdownViewer h3 { font-size: 1.2em; margin: 0.4em 0; }
        .fluffy-MarkdownViewer p { margin: 0.5em 0; }
        .fluffy-MarkdownViewer ul, .fluffy-MarkdownViewer ol { padding-left: 2em; margin: 0.5em 0; }
        .fluffy-MarkdownViewer code { padding: 2px 4px; border-radius: 3px; font-size: 0.9em; }
        .fluffy-MarkdownViewer pre { padding: 12px; border-radius: 6px; overflow-x: auto; margin: 0.5em 0; }
        .fluffy-MarkdownViewer pre code { padding: 0; }
        .fluffy-MarkdownViewer a { text-decoration: underline; }
        .fluffy-Link { text-decoration: underline; }
        .fluffy-Splitter-first { padding: 12px; }
        .fluffy-Splitter-second { padding: 16px; }
%s
    </style>
</head>
<body>
%s
    <script>%s</script>
</body>
</html>`, template.HTMLEscapeString(title), css, body, js))

	return []byte(b.String())
}
