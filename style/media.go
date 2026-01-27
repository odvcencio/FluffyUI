package style

import (
	"fmt"
	"strconv"
	"strings"
)

// MediaContext captures environment for media queries.
type MediaContext struct {
	Width, Height int
	ReducedMotion bool
}

// Orientation represents screen orientation.
type Orientation string

const (
	OrientationPortrait  Orientation = "portrait"
	OrientationLandscape Orientation = "landscape"
)

// MediaQuery captures a single media query (AND of conditions).
type MediaQuery struct {
	MinWidth      int
	MaxWidth      int
	MinHeight     int
	MaxHeight     int
	Orientation   Orientation
	ReducedMotion *bool
	Invalid       bool
}

func (q MediaQuery) Matches(ctx MediaContext) bool {
	if q.Invalid {
		return false
	}
	if q.MinWidth > 0 && ctx.Width < q.MinWidth {
		return false
	}
	if q.MaxWidth > 0 && ctx.Width > q.MaxWidth {
		return false
	}
	if q.MinHeight > 0 && ctx.Height < q.MinHeight {
		return false
	}
	if q.MaxHeight > 0 && ctx.Height > q.MaxHeight {
		return false
	}
	if q.Orientation != "" {
		orient := OrientationPortrait
		if ctx.Width >= ctx.Height {
			orient = OrientationLandscape
		}
		if q.Orientation != orient {
			return false
		}
	}
	if q.ReducedMotion != nil && ctx.ReducedMotion != *q.ReducedMotion {
		return false
	}
	return true
}

func mediaMatches(queries []MediaQuery, ctx MediaContext) bool {
	if len(queries) == 0 {
		return true
	}
	for _, q := range queries {
		if q.Matches(ctx) {
			return true
		}
	}
	return false
}

func parseMediaQueries(input string) ([]MediaQuery, error) {
	parts := splitByComma(input)
	if len(parts) == 0 {
		return nil, nil
	}
	out := make([]MediaQuery, 0, len(parts))
	for _, raw := range parts {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		q, err := parseMediaQuery(raw)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, nil
}

func splitByComma(value string) []string {
	var out []string
	start := 0
	depth := 0
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				out = append(out, value[start:i])
				start = i + 1
			}
		}
	}
	if start <= len(value) {
		out = append(out, value[start:])
	}
	return out
}

func parseMediaQuery(input string) (MediaQuery, error) {
	var q MediaQuery
	fields := splitByAnd(input)
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if strings.HasPrefix(field, "(") && strings.HasSuffix(field, ")") {
			field = strings.TrimSpace(field[1 : len(field)-1])
		}
		lower := strings.ToLower(field)
		if lower == "screen" || lower == "all" {
			continue
		}
		name, value, ok := strings.Cut(field, ":")
		if !ok {
			return MediaQuery{}, errMedia("expected ':' in %q", field)
		}
		name = strings.ToLower(strings.TrimSpace(name))
		value = strings.ToLower(strings.TrimSpace(value))
		switch name {
		case "min-width":
			v, err := parseMediaInt(value)
			if err != nil {
				return MediaQuery{}, errMedia("invalid min-width %q", value)
			}
			q.MinWidth = v
		case "max-width":
			v, err := parseMediaInt(value)
			if err != nil {
				return MediaQuery{}, errMedia("invalid max-width %q", value)
			}
			q.MaxWidth = v
		case "min-height":
			v, err := parseMediaInt(value)
			if err != nil {
				return MediaQuery{}, errMedia("invalid min-height %q", value)
			}
			q.MinHeight = v
		case "max-height":
			v, err := parseMediaInt(value)
			if err != nil {
				return MediaQuery{}, errMedia("invalid max-height %q", value)
			}
			q.MaxHeight = v
		case "orientation":
			switch value {
			case string(OrientationPortrait):
				q.Orientation = OrientationPortrait
			case string(OrientationLandscape):
				q.Orientation = OrientationLandscape
			default:
				return MediaQuery{}, errMedia("invalid orientation %q", value)
			}
		case "prefers-reduced-motion":
			switch value {
			case "reduce":
				v := true
				q.ReducedMotion = &v
			case "no-preference":
				v := false
				q.ReducedMotion = &v
			default:
				return MediaQuery{}, errMedia("invalid prefers-reduced-motion %q", value)
			}
		default:
			return MediaQuery{}, errMedia("unknown media feature %q", name)
		}
	}
	return q, nil
}

func splitByAnd(input string) []string {
	var out []string
	start := 0
	depth := 0
	lower := strings.ToLower(input)
	for i := 0; i < len(lower); i++ {
		switch lower[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 && strings.HasPrefix(lower[i:], "and") {
				before := strings.TrimSpace(input[start:i])
				afterIdx := i + 3
				isBoundary := (i == start || isSpace(lower[i-1])) && (afterIdx >= len(lower) || isSpace(lower[afterIdx]))
				if isBoundary {
					if before != "" {
						out = append(out, before)
					}
					start = afterIdx
					i = afterIdx - 1
				}
			}
		}
	}
	if start <= len(input) {
		out = append(out, input[start:])
	}
	return out
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func parseMediaInt(value string) (int, error) {
	value = strings.TrimSpace(strings.TrimSuffix(value, "px"))
	if value == "" {
		return 0, errMedia("missing number")
	}
	return strconv.Atoi(value)
}

func errMedia(format string, args ...any) error {
	return fmt.Errorf("style: media: "+format, args...)
}
