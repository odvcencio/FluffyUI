package fur

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Pretty formats any Go value for display.
func Pretty(v any) Renderable {
	return PrettyWith(v, PrettyOpts{})
}

// PrettyOpts configures formatting.
type PrettyOpts struct {
	Indent      int
	MaxDepth    int
	MaxItems    int
	MaxString   int
	ShowTypes   bool
	ShowPrivate bool
}

// PrettyWith formats a value using explicit options.
func PrettyWith(v any, opts PrettyOpts) Renderable {
	return prettyRenderable{value: v, opts: normalizePrettyOpts(opts)}
}

type prettyRenderable struct {
	value any
	opts  PrettyOpts
}

func (p prettyRenderable) Render(width int) []Line {
	state := &prettyState{opts: p.opts}
	state.formatValue(reflect.ValueOf(p.value), 0)
	state.flush()
	return wrapLines(state.lines, width)
}

func normalizePrettyOpts(opts PrettyOpts) PrettyOpts {
	if opts.Indent <= 0 {
		opts.Indent = 2
	}
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 6
	}
	if opts.MaxItems <= 0 {
		opts.MaxItems = 50
	}
	if opts.MaxString <= 0 {
		opts.MaxString = 200
	}
	return opts
}

type prettyState struct {
	opts        PrettyOpts
	lines       []Line
	current     Line
	indentLevel int
	visited     map[uintptr]int
}

func (s *prettyState) formatValue(v reflect.Value, depth int) {
	if !v.IsValid() {
		s.write("nil", dimStyle())
		return
	}
	if depth >= s.opts.MaxDepth {
		s.write("...", dimStyle())
		return
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			s.write("nil", dimStyle())
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			s.write("nil", dimStyle())
			return
		}
		cycle, exit := s.enter(v)
		if cycle {
			s.write("<cycle>", dimStyle())
			return
		}
		defer exit()
		s.write("&", punctStyle())
		s.formatValue(v.Elem(), depth+1)
	case reflect.Struct:
		cycle, exit := s.enter(v)
		if cycle {
			s.write("<cycle>", dimStyle())
			return
		}
		defer exit()
		s.formatStruct(v, depth)
	case reflect.Slice, reflect.Array:
		cycle, exit := s.enter(v)
		if cycle {
			s.write("<cycle>", dimStyle())
			return
		}
		defer exit()
		s.formatSlice(v, depth)
	case reflect.Map:
		cycle, exit := s.enter(v)
		if cycle {
			s.write("<cycle>", dimStyle())
			return
		}
		defer exit()
		s.formatMap(v, depth)
	case reflect.String:
		s.write(formatString(v.String(), s.opts.MaxString), stringStyle())
	case reflect.Bool:
		s.write(strconv.FormatBool(v.Bool()), boolStyle())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.write(fmt.Sprintf("%d", v.Int()), numberStyle())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		s.write(fmt.Sprintf("%d", v.Uint()), numberStyle())
	case reflect.Float32, reflect.Float64:
		s.write(formatFloat(v.Float()), numberStyle())
	case reflect.Complex64, reflect.Complex128:
		s.write(fmt.Sprintf("%v", v.Complex()), numberStyle())
	case reflect.Invalid:
		s.write("nil", dimStyle())
	default:
		if v.CanInterface() {
			s.write(fmt.Sprintf("%v", v.Interface()), plainStyle())
			return
		}
		s.write(v.Type().String(), typeStyle())
	}
}

func (s *prettyState) formatStruct(v reflect.Value, depth int) {
	if s.opts.ShowTypes {
		s.write(typeName(v.Type()), typeStyle())
	}
	s.write("{", punctStyle())
	fields := make([]structField, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if field.PkgPath != "" && !s.opts.ShowPrivate {
			continue
		}
		fields = append(fields, structField{name: field.Name, value: v.Field(i)})
	}
	if len(fields) == 0 {
		s.write("}", punctStyle())
		return
	}
	s.newline()
	s.indentLevel++
	for i, field := range fields {
		s.startLine()
		s.write(field.name, keyStyle())
		s.write(": ", punctStyle())
		s.formatValue(field.value, depth+1)
		if i < len(fields)-1 {
			s.write(",", punctStyle())
		}
		s.newline()
	}
	s.indentLevel--
	s.startLine()
	s.write("}", punctStyle())
}

func (s *prettyState) formatSlice(v reflect.Value, depth int) {
	if s.opts.ShowTypes {
		s.write(typeName(v.Type()), typeStyle())
	}
	s.write("{", punctStyle())
	length := v.Len()
	if length == 0 {
		s.write("}", punctStyle())
		return
	}
	limit := min(length, s.opts.MaxItems)
	s.newline()
	s.indentLevel++
	for i := 0; i < limit; i++ {
		s.startLine()
		s.formatValue(v.Index(i), depth+1)
		s.write(",", punctStyle())
		s.newline()
	}
	if length > limit {
		s.startLine()
		s.write(fmt.Sprintf("... (%d more)", length-limit), dimStyle())
		s.newline()
	}
	s.indentLevel--
	s.startLine()
	s.write("}", punctStyle())
}

func (s *prettyState) formatMap(v reflect.Value, depth int) {
	if s.opts.ShowTypes {
		s.write(typeName(v.Type()), typeStyle())
	}
	s.write("{", punctStyle())
	length := v.Len()
	if length == 0 {
		s.write("}", punctStyle())
		return
	}
	keys := v.MapKeys()
	keyInfos := make([]keyInfo, 0, len(keys))
	for _, key := range keys {
		repr := fmt.Sprintf("%v", key.Interface())
		keyInfos = append(keyInfos, keyInfo{key: key, repr: repr})
	}
	sort.Slice(keyInfos, func(i, j int) bool {
		return keyInfos[i].repr < keyInfos[j].repr
	})
	limit := min(len(keyInfos), s.opts.MaxItems)
	s.newline()
	s.indentLevel++
	for i := 0; i < limit; i++ {
		s.startLine()
		s.formatValue(keyInfos[i].key, depth+1)
		s.write(": ", punctStyle())
		s.formatValue(v.MapIndex(keyInfos[i].key), depth+1)
		s.write(",", punctStyle())
		s.newline()
	}
	if len(keyInfos) > limit {
		s.startLine()
		s.write(fmt.Sprintf("... (%d more)", len(keyInfos)-limit), dimStyle())
		s.newline()
	}
	s.indentLevel--
	s.startLine()
	s.write("}", punctStyle())
}

func (s *prettyState) write(text string, style Style) {
	appendSpan(&s.current, Span{Text: text, Style: style})
}

func (s *prettyState) startLine() {
	if s.indentLevel <= 0 || len(s.current) > 0 {
		return
	}
	s.write(strings.Repeat(" ", s.indentLevel*s.opts.Indent), DefaultStyle())
}

func (s *prettyState) newline() {
	s.lines = append(s.lines, s.current)
	s.current = nil
}

func (s *prettyState) flush() {
	if len(s.current) > 0 {
		s.lines = append(s.lines, s.current)
		s.current = nil
	}
	if len(s.lines) == 0 {
		s.lines = append(s.lines, Line{})
	}
}

func (s *prettyState) enter(v reflect.Value) (bool, func()) {
	ptr := pointerForValue(v)
	if ptr == 0 {
		return false, func() {}
	}
	if s.visited == nil {
		s.visited = make(map[uintptr]int)
	}
	if s.visited[ptr] > 0 {
		return true, func() {}
	}
	s.visited[ptr]++
	return false, func() {
		s.visited[ptr]--
		if s.visited[ptr] <= 0 {
			delete(s.visited, ptr)
		}
	}
}

type structField struct {
	name  string
	value reflect.Value
}

type keyInfo struct {
	key  reflect.Value
	repr string
}

func pointerForValue(v reflect.Value) uintptr {
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice:
		if v.IsNil() {
			return 0
		}
		return v.Pointer()
	}
	return 0
}

func typeName(t reflect.Type) string {
	if t == nil {
		return "<nil>"
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}

func formatString(value string, max int) string {
	if max <= 0 {
		return strconv.Quote(value)
	}
	if len([]rune(value)) <= max {
		return strconv.Quote(value)
	}
	runes := []rune(value)
	truncated := string(runes[:max]) + "..."
	return strconv.Quote(truncated)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func typeStyle() Style {
	return DefaultStyle().Foreground(ColorMagenta)
}

func keyStyle() Style {
	return DefaultStyle().Foreground(ColorCyan)
}

func stringStyle() Style {
	return DefaultStyle().Foreground(ColorGreen)
}

func numberStyle() Style {
	return DefaultStyle().Foreground(ColorYellow)
}

func boolStyle() Style {
	return DefaultStyle().Foreground(ColorYellow)
}

func plainStyle() Style {
	return DefaultStyle()
}

func punctStyle() Style {
	return DefaultStyle()
}

func dimStyle() Style {
	return Dim
}
