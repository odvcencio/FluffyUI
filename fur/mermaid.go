package fur

import (
	"regexp"
	"strings"
)

// MermaidFlowchart renders a simple mermaid flowchart as ASCII art.
// Supports basic syntax like:
//   A --> B
//   A -->|label| C
//   B --> D
func MermaidFlowchart(diagram string) Renderable {
	return mermaidRenderable{diagram: diagram}
}

type mermaidRenderable struct {
	diagram string
}

// Node represents a node in the flowchart
type mermaidNode struct {
	ID    string
	Label string
	X     int
	Y     int
}

// Edge represents a connection between nodes
type mermaidEdge struct {
	From  string
	To    string
	Label string
}

func (m mermaidRenderable) Render(width int) []Line {
	nodes, edges := m.parseDiagram()
	if len(nodes) == 0 {
		return nil
	}
	
	// Simple layout: arrange nodes in a grid
	layout := m.layoutNodes(nodes, edges)
	
	// Render to ASCII
	return m.renderASCII(layout, edges, width)
}

func (m mermaidRenderable) parseDiagram() (map[string]*mermaidNode, []mermaidEdge) {
	nodes := make(map[string]*mermaidNode)
	var edges []mermaidEdge
	
	lines := strings.Split(m.diagram, "\n")
	
	// Regex patterns
	nodePattern := regexp.MustCompile(`^\s*(\w+)\s*(?:\[(.+?)\]|\((.+?)\)|\{(.+?)\}|\[(.+?)\])\s*$`)
	edgePattern := regexp.MustCompile(`^\s*(\w+)\s*(-->|---|==>|\.->|-\.->)\s*\|?(.+?)\|?\s*(\w+)\s*$`)
	edgeSimplePattern := regexp.MustCompile(`^\s*(\w+)\s*(-->|---|==>|\.->|-\.->)\s*(\w+)\s*$`)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}
		
		// Try to parse edge with label
		if matches := edgePattern.FindStringSubmatch(line); matches != nil {
			from := matches[1]
			label := strings.Trim(matches[3], "|\"")
			to := matches[4]
			edges = append(edges, mermaidEdge{From: from, To: to, Label: label})
			// Ensure nodes exist
			if _, ok := nodes[from]; !ok {
				nodes[from] = &mermaidNode{ID: from, Label: from}
			}
			if _, ok := nodes[to]; !ok {
				nodes[to] = &mermaidNode{ID: to, Label: to}
			}
			continue
		}
		
		// Try simple edge
		if matches := edgeSimplePattern.FindStringSubmatch(line); matches != nil {
			from := matches[1]
			to := matches[3]
			edges = append(edges, mermaidEdge{From: from, To: to})
			if _, ok := nodes[from]; !ok {
				nodes[from] = &mermaidNode{ID: from, Label: from}
			}
			if _, ok := nodes[to]; !ok {
				nodes[to] = &mermaidNode{ID: to, Label: to}
			}
			continue
		}
		
		// Try node definition
		if matches := nodePattern.FindStringSubmatch(line); matches != nil {
			id := matches[1]
			// Get the first non-empty group for the label
			label := id
			for i := 2; i < len(matches); i++ {
				if matches[i] != "" {
					label = matches[i]
					break
				}
			}
			nodes[id] = &mermaidNode{ID: id, Label: label}
		}
	}
	
	return nodes, edges
}

func (m mermaidRenderable) layoutNodes(nodes map[string]*mermaidNode, edges []mermaidEdge) map[string]*mermaidNode {
	// Simple layout: position nodes in levels based on connectivity
	levels := make(map[string]int)
	
	// Find root nodes (no incoming edges)
	hasIncoming := make(map[string]bool)
	for _, edge := range edges {
		hasIncoming[edge.To] = true
	}
	
	// Initialize levels
	for id := range nodes {
		if !hasIncoming[id] {
			levels[id] = 0
		} else {
			levels[id] = -1
		}
	}
	
	// Propagate levels
	changed := true
	for changed {
		changed = false
		for _, edge := range edges {
			fromLevel := levels[edge.From]
			toLevel := levels[edge.To]
			if fromLevel >= 0 && (toLevel < 0 || toLevel <= fromLevel) {
				levels[edge.To] = fromLevel + 1
				changed = true
			}
		}
	}
	
	// Set default level for disconnected nodes
	for id := range nodes {
		if levels[id] < 0 {
			levels[id] = 0
		}
	}
	
	// Assign positions within levels
	levelCounts := make(map[int]int)
	for id, level := range levels {
		nodes[id].X = level * 20 // Horizontal spacing
		nodes[id].Y = levelCounts[level] * 4 // Vertical spacing
		levelCounts[level]++
	}
	
	return nodes
}

func (m mermaidRenderable) renderASCII(nodes map[string]*mermaidNode, edges []mermaidEdge, maxWidth int) []Line {
	if len(nodes) == 0 {
		return nil
	}
	
	// Find bounds
	maxX, maxY := 0, 0
	for _, node := range nodes {
		if node.X > maxX {
			maxX = node.X
		}
		if node.Y > maxY {
			maxY = node.Y
		}
	}
	
	// Create canvas
	width := maxX + 15
	height := maxY + 4
	if maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	
	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = make([]rune, width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}
	
	// Draw nodes
	for _, node := range nodes {
		m.drawNode(canvas, node)
	}
	
	// Draw edges
	for _, edge := range edges {
		m.drawEdge(canvas, nodes[edge.From], nodes[edge.To], edge.Label)
	}
	
	// Convert to lines
	var lines []Line
	for _, row := range canvas {
		text := strings.TrimRight(string(row), " ")
		if text != "" {
			lines = append(lines, Line{{Text: text, Style: DefaultStyle()}})
		}
	}
	
	return lines
}

func (m mermaidRenderable) drawNode(canvas [][]rune, node *mermaidNode) {
	x, y := node.X, node.Y
	label := node.Label
	if len(label) > 12 {
		label = label[:12]
	}
	
	// Draw box: +--------+
	//           | Label  |
	//           +--------+
	width := len(label) + 4
	
	// Top
	if y < len(canvas) && x+width < len(canvas[0]) {
		canvas[y][x] = '+'
		for i := 1; i < width-1; i++ {
			canvas[y][x+i] = '-'
		}
		canvas[y][x+width-1] = '+'
	}
	
	// Middle with label
	if y+1 < len(canvas) && x+width < len(canvas[0]) {
		canvas[y+1][x] = '|'
		canvas[y+1][x+1] = ' '
		for i, ch := range label {
			if x+2+i < len(canvas[0]) {
				canvas[y+1][x+2+i] = ch
			}
		}
		canvas[y+1][x+width-2] = ' '
		canvas[y+1][x+width-1] = '|'
	}
	
	// Bottom
	if y+2 < len(canvas) && x+width < len(canvas[0]) {
		canvas[y+2][x] = '+'
		for i := 1; i < width-1; i++ {
			canvas[y+2][x+i] = '-'
		}
		canvas[y+2][x+width-1] = '+'
	}
}

func (m mermaidRenderable) drawEdge(canvas [][]rune, from, to *mermaidNode, label string) {
	// Simple horizontal line from right of 'from' to left of 'to'
	fromX := from.X + len(from.Label) + 4
	fromY := from.Y + 1
	toX := to.X
	toY := to.Y + 1
	
	// Draw horizontal line
	if fromY == toY && fromY < len(canvas) {
		for x := fromX; x < toX && x < len(canvas[0]); x++ {
			if x > fromX && x < toX-1 {
				canvas[fromY][x] = '-'
			}
		}
		if toX-1 >= 0 && toX-1 < len(canvas[0]) {
			canvas[fromY][toX-1] = '>'
		}
	} else {
		// Draw angled connection (simplified)
		midX := (fromX + toX) / 2
		
		// From node to midpoint
		if fromY < len(canvas) {
			for x := fromX; x < midX && x < len(canvas[0]); x++ {
				canvas[fromY][x] = '-'
			}
		}
		
		// Vertical line
		startY, endY := fromY, toY
		if startY > endY {
			startY, endY = endY, startY
		}
		for y := startY; y <= endY && y < len(canvas) && midX < len(canvas[0]); y++ {
			canvas[y][midX] = '|'
		}
		
		// From midpoint to target
		if toY < len(canvas) {
			for x := midX; x < toX && x < len(canvas[0]); x++ {
				if x == midX {
					continue // Already drawn
				}
				canvas[toY][x] = '-'
			}
			if toX-1 >= 0 && toX-1 < len(canvas[0]) {
				canvas[toY][toX-1] = '>'
			}
		}
	}
}

// SimpleDiagram creates a simple text-based diagram from a description.
// Format: "A -> B -> C" creates a horizontal flow
func SimpleDiagram(description string) Renderable {
	return simpleDiagramRenderable{description: description}
}

type simpleDiagramRenderable struct {
	description string
}

func (s simpleDiagramRenderable) Render(width int) []Line {
	// Parse "A -> B -> C" format
	parts := strings.Split(s.description, "->")
	if len(parts) < 2 {
		return []Line{{{Text: s.description, Style: DefaultStyle()}}}
	}
	
	var boxes []string
	for _, p := range parts {
		boxes = append(boxes, strings.TrimSpace(p))
	}
	
	// Render as: [ Box A ] --> [ Box B ] --> [ Box C ]
	var result strings.Builder
	for i, box := range boxes {
		if i > 0 {
			result.WriteString(" --> ")
		}
		result.WriteString("[ ")
		result.WriteString(box)
		result.WriteString(" ]")
	}
	
	return []Line{{{Text: result.String(), Style: DefaultStyle()}}}
}
