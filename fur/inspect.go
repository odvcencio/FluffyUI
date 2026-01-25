package fur

import (
	"fmt"
	"reflect"
	"strings"
)

// Inspect generates a detailed report about a value.
func Inspect(v any) Renderable {
	return InspectWith(v, InspectOpts{})
}

// InspectOpts configures Inspect output.
type InspectOpts struct {
	Methods bool
	Fields  bool
	Private bool
}

// InspectWith returns Inspect output with options.
func InspectWith(v any, opts InspectOpts) Renderable {
	if !opts.Fields && !opts.Methods {
		opts.Fields = true
		opts.Methods = true
	}
	return inspectRenderable{value: v, opts: opts}
}

type inspectRenderable struct {
	value any
	opts  InspectOpts
}

func (i inspectRenderable) Render(width int) []Line {
	var content []Line
	value := i.value
	typeName := "<nil>"
	typeValue := reflect.TypeOf(value)
	if typeValue != nil {
		typeName = typeValue.String()
	}
	content = append(content, splitTextLines(fmt.Sprintf("Value: %v", value), DefaultStyle())...)

	if i.opts.Fields {
		content = append(content, Line{})
		content = append(content, splitTextLines("Fields:", DefaultStyle())...)
		content = append(content, inspectFields(value, i.opts.Private)...)
	}
	if i.opts.Methods {
		content = append(content, Line{})
		content = append(content, splitTextLines("Methods:", DefaultStyle())...)
		content = append(content, inspectMethods(value)...)
	}
	innerWidth := width
	if innerWidth > 0 {
		innerWidth -= 2
	}
	content = wrapLines(content, innerWidth)
	return renderBox(typeName, content, width, Dim)
}

func inspectFields(value any, includePrivate bool) []Line {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)
	if t == nil {
		return []Line{{}}
	}
	for t.Kind() == reflect.Pointer {
		if v.IsNil() {
			return []Line{{}}
		}
		v = v.Elem()
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return []Line{{}}
	}
	var lines []Line
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" && !includePrivate {
			continue
		}
		line := fmt.Sprintf("  %s %s", field.Name, field.Type.String())
		lines = append(lines, splitTextLines(line, DefaultStyle())...)
	}
	if len(lines) == 0 {
		lines = append(lines, Line{})
	}
	return lines
}

func inspectMethods(value any) []Line {
	t := reflect.TypeOf(value)
	if t == nil {
		return []Line{{}}
	}
	methods := make([]string, 0, t.NumMethod())
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methods = append(methods, "  "+formatMethodSignature(method))
	}
	if len(methods) == 0 {
		return []Line{{}}
	}
	var lines []Line
	for _, entry := range methods {
		lines = append(lines, splitTextLines(entry, DefaultStyle())...)
	}
	return lines
}

func formatMethodSignature(method reflect.Method) string {
	t := method.Type
	var args []string
	for i := 1; i < t.NumIn(); i++ {
		args = append(args, shortType(t.In(i)))
	}
	var outs []string
	for i := 0; i < t.NumOut(); i++ {
		outs = append(outs, shortType(t.Out(i)))
	}
	signature := method.Name + "(" + strings.Join(args, ", ") + ")"
	if len(outs) == 1 {
		signature += " " + outs[0]
	} else if len(outs) > 1 {
		signature += " (" + strings.Join(outs, ", ") + ")"
	}
	return signature
}

func shortType(t reflect.Type) string {
	if t == nil {
		return "<nil>"
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}
