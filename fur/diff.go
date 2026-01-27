package fur

import (
	"bufio"
	"fmt"
	"strings"
)

// Diff renders a unified diff with syntax highlighting.
func Diff(content string) Renderable {
	return diffRenderable{content: content}
}

type diffRenderable struct {
	content string
}

func (d diffRenderable) Render(width int) []Line {
	var lines []Line
	scanner := bufio.NewScanner(strings.NewReader(d.content))
	
	for scanner.Scan() {
		line := scanner.Text()
		styled := d.styleLine(line)
		lines = append(lines, styled)
	}
	
	return lines
}

func (d diffRenderable) styleLine(line string) Line {
	if len(line) == 0 {
		return Line{}
	}
	
	switch {
	case strings.HasPrefix(line, "+++ "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorGreen).Bold()}}
	case strings.HasPrefix(line, "--- "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorRed).Bold()}}
	case strings.HasPrefix(line, "@@ "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorCyan)}} // Hunk header
	case strings.HasPrefix(line, "+"):
		return Line{{Text: line, Style: Style{}.Foreground(ColorGreen)}} // Added
	case strings.HasPrefix(line, "-"):
		return Line{{Text: line, Style: Style{}.Foreground(ColorRed)}} // Deleted
	case strings.HasPrefix(line, " "):
		return Line{{Text: line, Style: DefaultStyle()}} // Context
	case strings.HasPrefix(line, "diff "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorMagenta).Bold()}}
	case strings.HasPrefix(line, "index "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorYellow).Dim()}}
	case strings.HasPrefix(line, "mode "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorYellow).Dim()}}
	case strings.HasPrefix(line, "new file "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorGreen).Bold()}}
	case strings.HasPrefix(line, "deleted file "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorRed).Bold()}}
	case strings.HasPrefix(line, "rename "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorBlue).Bold()}}
	case strings.HasPrefix(line, "similarity "):
		return Line{{Text: line, Style: Style{}.Foreground(ColorYellow)}} // Similarity index
	default:
		return Line{{Text: line, Style: DefaultStyle()}}
	}
}

// DiffStats shows statistics for a diff.
func DiffStats(added, deleted, modified int) Renderable {
	return diffStatsRenderable{added: added, deleted: deleted, modified: modified}
}

type diffStatsRenderable struct {
	added    int
	deleted  int
	modified int
}

func (d diffStatsRenderable) Render(width int) []Line {
	var parts []string
	if d.added > 0 {
		parts = append(parts, fmt.Sprintf("+[green]%d[-] added", d.added))
	}
	if d.deleted > 0 {
		parts = append(parts, fmt.Sprintf("-[red]%d[-] deleted", d.deleted))
	}
	if d.modified > 0 {
		parts = append(parts, fmt.Sprintf("~[yellow]%d[-] modified", d.modified))
	}
	
	text := strings.Join(parts, ", ")
	if text == "" {
		text = "No changes"
	}
	
	return Markup(text).Render(width)
}
