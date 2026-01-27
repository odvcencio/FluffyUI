package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func writeFile(path string, data []byte, perm os.FileMode, force bool) error {
	if path == "" {
		return errors.New("missing file path")
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file already exists: %s", path)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func ensureDir(path string) error {
	if path == "" {
		return errors.New("missing directory path")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return nil
}

func dirEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, fmt.Errorf("path is not a directory: %s", path)
	}
	dir, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer dir.Close()
	_, err = dir.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return true, nil
	}
	return false, err
}

func renderTemplate(tmpl string, data any) (string, error) {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func toSnake(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	var out []rune
	var prevLower bool
	var prevDigit bool
	underscore := func() {
		if len(out) == 0 {
			return
		}
		if out[len(out)-1] != '_' {
			out = append(out, '_')
		}
	}
	for _, r := range name {
		switch {
		case r == '-' || r == '_' || r == ' ' || r == '.' || r == '/':
			underscore()
			prevLower = false
			prevDigit = false
			continue
		case r >= 'A' && r <= 'Z':
			if (prevLower || prevDigit) && len(out) > 0 {
				underscore()
			}
			r += 'a' - 'A'
		}
		if r >= '0' && r <= '9' {
			prevDigit = true
		} else {
			prevDigit = false
		}
		out = append(out, r)
		prevLower = r >= 'a' && r <= 'z'
	}
	result := strings.Trim(string(out), "_")
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	return result
}

func toPascal(name string) string {
	parts := strings.Split(toSnake(name), "_")
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		if len(part) == 1 {
			b.WriteString(strings.ToUpper(part))
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}
	return b.String()
}

func titleFromName(name string) string {
	parts := strings.Split(toSnake(name), "_")
	var titled []string
	for _, part := range parts {
		if part == "" {
			continue
		}
		if len(part) == 1 {
			titled = append(titled, strings.ToUpper(part))
			continue
		}
		titled = append(titled, strings.ToUpper(part[:1])+part[1:])
	}
	if len(titled) == 0 {
		return strings.TrimSpace(name)
	}
	return strings.Join(titled, " ")
}
