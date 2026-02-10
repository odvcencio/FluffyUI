package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

// blockedImports are packages that playground user code must never import.
var blockedImports = map[string]bool{
	"os/exec":       true,
	"syscall":       true,
	"unsafe":        true,
	"plugin":        true,
	"net":           true,
	"net/http":      true,
	"C":             true,
	"runtime/cgo":   true,
	"debug/elf":     true,
	"os/signal":     true,
	"runtime/debug": true,
}

// blockedImportPrefixes are package prefixes that user code must never import.
var blockedImportPrefixes = []string{
	"net/",
}

// validateCode checks that the source is a valid main package with no
// dangerous imports. It returns a non-nil error describing the first
// violation found.
func validateCode(src []byte) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, parser.ImportsOnly)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	if f.Name.Name != "main" {
		return fmt.Errorf("package must be \"main\", got %q", f.Name.Name)
	}
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if blockedImports[path] {
			return fmt.Errorf("import %q is not allowed", path)
		}
		for _, prefix := range blockedImportPrefixes {
			if strings.HasPrefix(path, prefix) {
				return fmt.Errorf("import %q is not allowed", path)
			}
		}
	}
	return nil
}
