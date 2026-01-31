package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type typeInfo struct {
	kind      string
	aliasBase string
}

type constructor struct {
	Name      string
	Signature string
	Args      []string
	HasError  bool
	Doc       string
}

type widgetEntry struct {
	Name         string
	Doc          string
	Constructors []constructor
}

func main() {
	root := flag.String("root", ".", "repository root")
	out := flag.String("out", "docs/api/widgets.md", "output file")
	flag.Parse()

	widgetsDir := filepath.Join(*root, "widgets")
	entries, err := scanWidgets(widgetsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan widgets: %v\n", err)
		os.Exit(1)
	}

	if err := writeDoc(*out, entries); err != nil {
		fmt.Fprintf(os.Stderr, "write doc: %v\n", err)
		os.Exit(1)
	}
}

func scanWidgets(dir string) ([]widgetEntry, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(info os.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	pkg, ok := pkgs["widgets"]
	if !ok {
		return nil, fmt.Errorf("widgets package not found")
	}

	typeInfos := map[string]typeInfo{}
	typeDocs := map[string]string{}
	funcDocs := map[string]string{}
	funcDecls := []*ast.FuncDecl{}

	for _, file := range pkg.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch decl := node.(type) {
			case *ast.GenDecl:
				if decl.Tok != token.TYPE {
					return true
				}
				for _, spec := range decl.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || !ts.Name.IsExported() {
						continue
					}
					info := typeInfo{}
					switch t := ts.Type.(type) {
					case *ast.StructType:
						info.kind = "struct"
					case *ast.InterfaceType:
						info.kind = "interface"
					case *ast.Ident:
						info.kind = "alias"
						info.aliasBase = t.Name
					case *ast.SelectorExpr:
						info.kind = "alias"
						info.aliasBase = exprString(fset, t)
					default:
						info.kind = "alias"
					}
					typeInfos[ts.Name.Name] = info
					doc := ""
					if ts.Doc != nil {
						doc = strings.TrimSpace(ts.Doc.Text())
					} else if decl.Doc != nil {
						doc = strings.TrimSpace(decl.Doc.Text())
					}
					if doc != "" {
						typeDocs[ts.Name.Name] = firstSentence(doc)
					}
				}
			case *ast.FuncDecl:
				if decl.Name.IsExported() && strings.HasPrefix(decl.Name.Name, "New") {
					funcDecls = append(funcDecls, decl)
					if decl.Doc != nil {
						if doc := strings.TrimSpace(decl.Doc.Text()); doc != "" {
							funcDocs[decl.Name.Name] = firstSentence(doc)
						}
					}
				}
			}
			return true
		})
	}

	entries := map[string]*widgetEntry{}
	for _, fn := range funcDecls {
		retName, hasError := returnTypeName(fn)
		if retName == "" {
			continue
		}
		if !ast.IsExported(retName) {
			continue
		}
		entry := entries[retName]
		if entry == nil {
			entry = &widgetEntry{Name: retName}
			entries[retName] = entry
		}
		if entry.Doc == "" {
			entry.Doc = typeDocs[retName]
		}

		args := exampleArgs(typeInfos, fn)
		signature := fnSignature(fset, fn)
		constructor := constructor{
			Name:      fn.Name.Name,
			Signature: signature,
			Args:      args,
			HasError:  hasError,
			Doc:       funcDocs[fn.Name.Name],
		}
		entry.Constructors = append(entry.Constructors, constructor)
	}

	list := make([]widgetEntry, 0, len(entries))
	for _, entry := range entries {
		sort.Slice(entry.Constructors, func(i, j int) bool {
			return entry.Constructors[i].Name < entry.Constructors[j].Name
		})
		list = append(list, *entry)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no constructors found")
	}

	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list, nil
}

func firstSentence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	return strings.TrimSpace(lines[0])
}

func returnTypeName(fn *ast.FuncDecl) (string, bool) {
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return "", false
	}
	result := fn.Type.Results.List[0].Type
	name := baseIdent(result)
	if name == "" {
		return "", false
	}

	hasError := false
	if len(fn.Type.Results.List) > 1 {
		last := fn.Type.Results.List[len(fn.Type.Results.List)-1]
		switch t := last.Type.(type) {
		case *ast.Ident:
			hasError = t.Name == "error"
		}
	}
	return name, hasError
}

func baseIdent(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return baseIdent(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

func fnSignature(fset *token.FileSet, fn *ast.FuncDecl) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, fn.Type)
	sig := strings.TrimPrefix(buf.String(), "func")
	return fn.Name.Name + sig
}

func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, expr)
	return buf.String()
}

func exampleArgs(typeInfos map[string]typeInfo, fn *ast.FuncDecl) []string {
	if fn.Type.Params == nil {
		return nil
	}
	args := []string{}
	for _, field := range fn.Type.Params.List {
		if _, ok := field.Type.(*ast.Ellipsis); ok {
			continue
		}
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			args = append(args, exampleValue(typeInfos, field.Type))
		}
	}
	return args
}

func exampleValue(typeInfos map[string]typeInfo, expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return zeroForIdent(typeInfos, t.Name)
	case *ast.StarExpr:
		return "nil"
	case *ast.ArrayType:
		if t.Len == nil {
			return "nil"
		}
		return exprString(token.NewFileSet(), expr) + "{}"
	case *ast.MapType:
		return "nil"
	case *ast.InterfaceType:
		return "nil"
	case *ast.FuncType:
		return "nil"
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			if pkg.Name == "time" && t.Sel.Name == "Duration" {
				return "0"
			}
			if pkg.Name == "runtime" && t.Sel.Name == "Rect" {
				return "runtime.Rect{}"
			}
		}
		return "nil"
	case *ast.IndexExpr, *ast.IndexListExpr:
		return "nil"
	default:
		return "nil"
	}
}

func zeroForIdent(typeInfos map[string]typeInfo, name string) string {
	switch name {
	case "string":
		return "\"\""
	case "bool":
		return "false"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return "0"
	case "float32", "float64":
		return "0"
	case "byte", "rune":
		return "0"
	case "error":
		return "nil"
	}
	info, ok := typeInfos[name]
	if !ok {
		return "nil"
	}
	switch info.kind {
	case "interface":
		return "nil"
	case "struct":
		return name + "{}"
	case "alias":
		base := info.aliasBase
		switch base {
		case "string":
			return name + "(\"\")"
		case "bool":
			return name + "(false)"
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
			return name + "(0)"
		case "float32", "float64":
			return name + "(0)"
		case "time.Duration":
			return name + "(0)"
		default:
			return name + "{}"
		}
	default:
		return "nil"
	}
}

func writeDoc(path string, entries []widgetEntry) error {
	var buf bytes.Buffer
	buf.WriteString("# Widgets API (Generated)\n\n")
	buf.WriteString("> This file is generated by `go run ./tools/gen_widgets_api`. Do not edit by hand.\n\n")
	buf.WriteString("FluffyUI widgets are constructed via `New*` functions and configured with option helpers.\n\n")
	buf.WriteString("## Widget Reference\n\n")

	for _, entry := range entries {
		buf.WriteString("### " + entry.Name + "\n\n")
		if entry.Doc != "" {
			buf.WriteString(entry.Doc + "\n\n")
		}
		buf.WriteString("Constructors:\n")
		for _, ctor := range entry.Constructors {
			buf.WriteString("- `" + ctor.Signature + "`\n")
		}
		buf.WriteString("\n")

		ctor := entry.Constructors[0]
		call := "widgets." + ctor.Name + "(" + strings.Join(ctor.Args, ", ") + ")"
		buf.WriteString("Example:\n\n")
		buf.WriteString("```go\n")
		varName := lowerFirst(entry.Name)
		if ctor.HasError {
			buf.WriteString(varName + ", err := " + call + "\n")
			buf.WriteString("if err != nil {\n")
			buf.WriteString("\t// handle error\n")
			buf.WriteString("}\n")
		} else {
			buf.WriteString(varName + " := " + call + "\n")
		}
		buf.WriteString("```")
		buf.WriteString("\n\n")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func lowerFirst(name string) string {
	if name == "" {
		return "widget"
	}
	runes := []rune(name)
	if unicode.IsUpper(runes[0]) {
		i := 0
		for i < len(runes) && unicode.IsUpper(runes[i]) {
			i++
		}
		if i > 1 && i < len(runes) && unicode.IsLower(runes[i]) {
			i--
		}
		for j := 0; j < i; j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		if i == 0 {
			runes[0] = unicode.ToLower(runes[0])
		}
		return string(runes)
	}
	return string(runes)
}
