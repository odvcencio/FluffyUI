// Package i18n bidi.go provides bidirectional text utilities for RTL language support.
// This implements a simplified Unicode Bidirectional Algorithm (UBA) suitable for
// terminal rendering of Arabic, Hebrew, and other RTL scripts.
package i18n

import (
	"strings"
	"unicode"
)

// String returns the direction as a human-readable string.
func (d Direction) String() string {
	if d == DirectionRTL {
		return "RTL"
	}
	return "LTR"
}

// DetectDirection returns the dominant text direction of a string
// by checking for RTL Unicode ranges (Arabic, Hebrew, etc.).
// If the string has more RTL characters than LTR strong characters,
// it returns DirectionRTL. Otherwise, it returns DirectionLTR.
func DetectDirection(text string) Direction {
	var ltrCount, rtlCount int
	for _, r := range text {
		if IsRTLRune(r) {
			rtlCount++
		} else if isStrongLTR(r) {
			ltrCount++
		}
	}
	if rtlCount > ltrCount {
		return DirectionRTL
	}
	return DirectionLTR
}

// IsRTLRune returns true if the rune is in an RTL script range.
// Covers Arabic, Hebrew, Syriac, Thaana, NKo, Samaritan, Mandaic,
// and various directional formatting characters.
func IsRTLRune(r rune) bool {
	return unicode.Is(unicode.Arabic, r) ||
		unicode.Is(unicode.Hebrew, r) ||
		unicode.Is(unicode.Syriac, r) ||
		unicode.Is(unicode.Thaana, r) ||
		unicode.Is(unicode.Nko, r) ||
		unicode.Is(unicode.Samaritan, r) ||
		unicode.Is(unicode.Mandaic, r) ||
		// RTL directional marks
		r == 0x200F || // RIGHT-TO-LEFT MARK
		r == 0x202B || // RIGHT-TO-LEFT EMBEDDING
		r == 0x202E || // RIGHT-TO-LEFT OVERRIDE
		r == 0x2067 // RIGHT-TO-LEFT ISOLATE
}

// isStrongLTR returns true if the rune is a strong LTR character (Latin, CJK, etc.)
func isStrongLTR(r rune) bool {
	return unicode.Is(unicode.Latin, r) ||
		unicode.Is(unicode.Greek, r) ||
		unicode.Is(unicode.Cyrillic, r) ||
		unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Thai, r) ||
		unicode.Is(unicode.Devanagari, r) ||
		r == 0x200E || // LEFT-TO-RIGHT MARK
		r == 0x202A || // LEFT-TO-RIGHT EMBEDDING
		r == 0x202D || // LEFT-TO-RIGHT OVERRIDE
		r == 0x2066 // LEFT-TO-RIGHT ISOLATE
}

// ReverseString reverses a string rune-by-rune (for RTL rendering).
func ReverseString(s string) string {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}

// bracketMirrors maps opening brackets to closing and vice versa.
var bracketMirrors = map[rune]rune{
	'(':  ')',
	')':  '(',
	'[':  ']',
	']':  '[',
	'{':  '}',
	'}':  '{',
	'<':  '>',
	'>':  '<',
	// Unicode brackets
	'\u00AB': '\u00BB', // << -> >>
	'\u00BB': '\u00AB', // >> -> <<
	'\u2018': '\u2019', // left single quote -> right single quote
	'\u2019': '\u2018', // right single quote -> left single quote
	'\u201C': '\u201D', // left double quote -> right double quote
	'\u201D': '\u201C', // right double quote -> left double quote
	'\u2039': '\u203A', // single left angle quote -> single right angle quote
	'\u203A': '\u2039', // single right angle quote -> single left angle quote
}

// MirrorBrackets replaces brackets/parens with their mirrored equivalents for RTL.
// ( -> ), [ -> ], { -> }, < -> > and vice versa.
func MirrorBrackets(s string) string {
	runes := []rune(s)
	changed := false
	for i, r := range runes {
		if mirror, ok := bracketMirrors[r]; ok {
			runes[i] = mirror
			changed = true
		}
	}
	if !changed {
		return s
	}
	return string(runes)
}

// runType classifies a rune for the bidi algorithm.
type runType int

const (
	runTypeRTL     runType = iota // Strong RTL character
	runTypeLTR                    // Strong LTR character
	runTypeNeutral                // Whitespace, punctuation, numbers
)

// classifyRune returns the bidi type of a rune.
func classifyRune(r rune) runType {
	if IsRTLRune(r) {
		return runTypeRTL
	}
	if isStrongLTR(r) {
		return runTypeLTR
	}
	return runTypeNeutral
}

// bidiRun represents a contiguous run of characters with the same directionality.
type bidiRun struct {
	text      string
	direction Direction
}

// BidiReorder reorders a mixed LTR/RTL string for visual display.
// Uses a simplified Unicode Bidi Algorithm (UBA):
//   - RTL runs are reversed and brackets are mirrored
//   - LTR runs are kept as-is
//   - Numbers within RTL context stay LTR (not reversed)
//   - Neutral characters (spaces, punctuation) inherit direction from context
func BidiReorder(text string, baseDirection Direction) string {
	if text == "" {
		return ""
	}

	// Check if text is purely one direction -- fast path
	hasRTL := false
	hasLTR := false
	hasDigit := false
	for _, r := range text {
		rt := classifyRune(r)
		if rt == runTypeRTL {
			hasRTL = true
		} else if rt == runTypeLTR {
			hasLTR = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if hasRTL && hasLTR {
			break
		}
	}

	// Pure LTR text with LTR base: return as-is
	if !hasRTL && baseDirection == DirectionLTR {
		return text
	}

	// All-neutral text (no strong chars) with RTL base: reverse and mirror
	if !hasLTR && !hasRTL && baseDirection == DirectionRTL {
		return MirrorBrackets(ReverseString(text))
	}

	// Pure RTL text (no strong LTR) without digits: reverse and mirror
	// If digits are present, we need the full algorithm to keep them LTR
	if hasRTL && !hasLTR && !hasDigit {
		return MirrorBrackets(ReverseString(text))
	}

	// Mixed text: split into runs and reorder
	runs := splitBidiRuns(text, baseDirection)

	if baseDirection == DirectionRTL {
		// In RTL base direction, the overall order of runs is reversed
		return joinRunsRTLBase(runs)
	}

	// LTR base direction: runs stay in logical order, RTL runs are reversed
	return joinRunsLTRBase(runs)
}

// splitBidiRuns splits text into directional runs.
func splitBidiRuns(text string, baseDirection Direction) []bidiRun {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}

	// Classify each rune
	types := make([]runType, len(runes))
	for i, r := range runes {
		types[i] = classifyRune(r)
	}

	// Resolve neutral characters: they inherit direction from surrounding strong characters.
	// If between two characters of the same direction, inherit that direction.
	// Otherwise, inherit the base direction.
	resolved := make([]Direction, len(runes))
	for i, rt := range types {
		switch rt {
		case runTypeRTL:
			resolved[i] = DirectionRTL
		case runTypeLTR:
			resolved[i] = DirectionLTR
		case runTypeNeutral:
			// Look for surrounding strong characters
			prevDir := baseDirection
			for j := i - 1; j >= 0; j-- {
				if types[j] != runTypeNeutral {
					if types[j] == runTypeRTL {
						prevDir = DirectionRTL
					} else {
						prevDir = DirectionLTR
					}
					break
				}
			}
			nextDir := baseDirection
			for j := i + 1; j < len(runes); j++ {
				if types[j] != runTypeNeutral {
					if types[j] == runTypeRTL {
						nextDir = DirectionRTL
					} else {
						nextDir = DirectionLTR
					}
					break
				}
			}
			if prevDir == nextDir {
				resolved[i] = prevDir
			} else {
				resolved[i] = baseDirection
			}
		}
	}

	// Numbers within RTL context stay LTR
	for i, r := range runes {
		if unicode.IsDigit(r) && resolved[i] == DirectionRTL {
			resolved[i] = DirectionLTR
		}
	}

	// Group consecutive same-direction characters into runs
	var runs []bidiRun
	runStart := 0
	for i := 1; i <= len(runes); i++ {
		if i < len(runes) && resolved[i] == resolved[runStart] {
			continue
		}
		run := bidiRun{
			text:      string(runes[runStart:i]),
			direction: resolved[runStart],
		}
		runs = append(runs, run)
		runStart = i
	}

	return runs
}

// joinRunsLTRBase joins runs for LTR base direction.
// RTL runs are reversed and brackets mirrored; LTR runs are kept as-is.
func joinRunsLTRBase(runs []bidiRun) string {
	var b strings.Builder
	for _, run := range runs {
		if run.direction == DirectionRTL {
			b.WriteString(MirrorBrackets(ReverseString(run.text)))
		} else {
			b.WriteString(run.text)
		}
	}
	return b.String()
}

// joinRunsRTLBase joins runs for RTL base direction.
// The overall order of runs is reversed (last run first),
// RTL runs are reversed, and LTR runs are kept as-is.
func joinRunsRTLBase(runs []bidiRun) string {
	var b strings.Builder
	for i := len(runs) - 1; i >= 0; i-- {
		run := runs[i]
		if run.direction == DirectionRTL {
			b.WriteString(MirrorBrackets(ReverseString(run.text)))
		} else {
			b.WriteString(run.text)
		}
	}
	return b.String()
}
