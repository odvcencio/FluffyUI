package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type themeFile struct {
	Name   string                       `yaml:"name"`
	Colors map[string]string            `yaml:"colors"`
	Styles map[string]map[string]string `yaml:"styles"`
}

type colorRGB struct {
	r float64
	g float64
	b float64
}

func runTheme(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: fluffy theme init|check|export|list|install [flags]")
	}
	switch args[0] {
	case "init":
		return runThemeInit(args[1:])
	case "check":
		return runThemeCheck(args[1:])
	case "export":
		return runThemeExport(args[1:])
	case "list":
		return runThemeList(args[1:])
	case "install":
		return runThemeInstall(args[1:])
	default:
		return fmt.Errorf("unknown theme command: %s", args[0])
	}
}

func runThemeInit(args []string) error {
	fs := flag.NewFlagSet("theme init", flag.ContinueOnError)
	path := fs.String("path", "themes/default.yaml", "theme file path")
	force := fs.Bool("force", false, "overwrite existing file")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return writeFile(*path, []byte(defaultThemeTemplate), 0o644, *force)
}

func runThemeCheck(args []string) error {
	fs := flag.NewFlagSet("theme check", flag.ContinueOnError)
	path := fs.String("path", "themes/default.yaml", "theme file path")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	theme, err := loadTheme(*path)
	if err != nil {
		return err
	}
	issues := validateTheme(theme)
	if len(issues) == 0 {
		fmt.Fprintln(os.Stdout, "theme check passed")
		return nil
	}
	for _, issue := range issues {
		fmt.Fprintln(os.Stderr, issue)
	}
	return fmt.Errorf("theme check failed: %d issue(s)", len(issues))
}

func runThemeExport(args []string) error {
	fs := flag.NewFlagSet("theme export", flag.ContinueOnError)
	path := fs.String("path", "themes/default.yaml", "theme file path")
	outPath := fs.String("output", "theme.css", "output css file")
	force := fs.Bool("force", false, "overwrite existing file")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	theme, err := loadTheme(*path)
	if err != nil {
		return err
	}
	css, err := exportThemeCSS(theme)
	if err != nil {
		return err
	}
	if err := ensureDir(filepath.Dir(*outPath)); err != nil {
		return err
	}
	return writeFile(*outPath, []byte(css), 0o644, *force)
}

type themeCatalog struct {
	Themes []themeCatalogEntry `yaml:"themes"`
}

type themeCatalogEntry struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
}

func runThemeList(args []string) error {
	fs := flag.NewFlagSet("theme list", flag.ContinueOnError)
	dir := fs.String("dir", "themes", "themes directory")
	catalogPath := fs.String("catalog", "", "optional catalog file")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *catalogPath != "" {
		raw, err := os.ReadFile(*catalogPath)
		if err != nil {
			return err
		}
		var catalog themeCatalog
		if err := yaml.Unmarshal(raw, &catalog); err != nil {
			return fmt.Errorf("parse catalog: %w", err)
		}
		for _, entry := range catalog.Themes {
			name := strings.TrimSpace(entry.Name)
			if name == "" {
				name = entry.Path
			}
			if entry.Description != "" {
				fmt.Fprintf(os.Stdout, "%s - %s\n", name, entry.Description)
			} else {
				fmt.Fprintln(os.Stdout, name)
			}
		}
		return nil
	}

	entries, err := os.ReadDir(*dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			fmt.Fprintln(os.Stdout, name)
		}
	}
	return nil
}

func runThemeInstall(args []string) error {
	fs := flag.NewFlagSet("theme install", flag.ContinueOnError)
	source := fs.String("source", "", "source theme file")
	dir := fs.String("dir", "themes", "destination directory")
	force := fs.Bool("force", false, "overwrite existing file")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*source) == "" {
		return errors.New("missing --source theme file")
	}
	if _, err := loadTheme(*source); err != nil {
		return err
	}
	if err := ensureDir(*dir); err != nil {
		return err
	}
	target := filepath.Join(*dir, filepath.Base(*source))
	content, err := os.ReadFile(*source)
	if err != nil {
		return err
	}
	return writeFile(target, content, 0o644, *force)
}

func loadTheme(path string) (themeFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return themeFile{}, err
	}
	var tf themeFile
	if err := yaml.Unmarshal(raw, &tf); err != nil {
		return themeFile{}, fmt.Errorf("parse theme: %w", err)
	}
	if tf.Colors == nil {
		tf.Colors = map[string]string{}
	}
	if tf.Styles == nil {
		tf.Styles = map[string]map[string]string{}
	}
	return tf, nil
}

func validateTheme(tf themeFile) []string {
	var issues []string
	if len(tf.Colors) == 0 {
		issues = append(issues, "theme: no colors defined")
	}
	for name, value := range tf.Colors {
		if _, _, err := resolveColor(value, tf.Colors); err != nil {
			issues = append(issues, fmt.Sprintf("color %q: %v", name, err))
		}
	}
	for selector, props := range tf.Styles {
		fgValue := pickProp(props, "foreground", "fg", "color")
		bgValue := pickProp(props, "background", "bg")
		if fgValue == "" || bgValue == "" {
			continue
		}
		fg, _, err := resolveColor(fgValue, tf.Colors)
		if err != nil {
			issues = append(issues, fmt.Sprintf("style %q foreground: %v", selector, err))
			continue
		}
		bg, _, err := resolveColor(bgValue, tf.Colors)
		if err != nil {
			issues = append(issues, fmt.Sprintf("style %q background: %v", selector, err))
			continue
		}
		contrast := contrastRatio(fg, bg)
		if contrast < 4.5 {
			issues = append(issues, fmt.Sprintf("style %q contrast %.2f below AA (4.5)", selector, contrast))
		}
	}
	return issues
}

func exportThemeCSS(tf themeFile) (string, error) {
	var selectors []string
	for selector := range tf.Styles {
		selectors = append(selectors, selector)
	}
	sort.Strings(selectors)

	var sb strings.Builder
	sb.WriteString("/* Generated by fluffy theme export */\n")
	for _, selector := range selectors {
		props := tf.Styles[selector]
		fg := pickProp(props, "foreground", "fg", "color")
		bg := pickProp(props, "background", "bg")
		fgHex := ""
		bgHex := ""
		if fg != "" {
			if _, hex, err := resolveColor(fg, tf.Colors); err == nil {
				fgHex = hex
			}
		}
		if bg != "" {
			if _, hex, err := resolveColor(bg, tf.Colors); err == nil {
				bgHex = hex
			}
		}
		if fgHex == "" && bgHex == "" {
			continue
		}
		sb.WriteString(selector)
		sb.WriteString(" {\n")
		if fgHex != "" {
			sb.WriteString("  color: ")
			sb.WriteString(fgHex)
			sb.WriteString(";\n")
		}
		if bgHex != "" {
			sb.WriteString("  background-color: ")
			sb.WriteString(bgHex)
			sb.WriteString(";\n")
		}
		sb.WriteString("}\n\n")
	}
	return sb.String(), nil
}

func pickProp(props map[string]string, keys ...string) string {
	if len(props) == 0 {
		return ""
	}
	for _, key := range keys {
		if value, ok := props[key]; ok {
			return strings.TrimSpace(value)
		}
		for prop, value := range props {
			if strings.EqualFold(prop, key) {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func resolveColor(value string, colors map[string]string) (colorRGB, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return colorRGB{}, "", errors.New("empty color value")
	}
	visited := map[string]bool{}
	for {
		if strings.HasPrefix(value, "#") {
			rgb, hex, err := parseHexColor(value)
			if err != nil {
				return colorRGB{}, "", err
			}
			return rgb, hex, nil
		}
		ref := strings.TrimSpace(value)
		if ref == "" {
			return colorRGB{}, "", errors.New("empty color reference")
		}
		if visited[ref] {
			return colorRGB{}, "", fmt.Errorf("cyclic color reference: %s", ref)
		}
		visited[ref] = true
		next, ok := colors[ref]
		if !ok {
			return colorRGB{}, "", fmt.Errorf("unknown color %q", ref)
		}
		value = strings.TrimSpace(next)
	}
}

func parseHexColor(value string) (colorRGB, string, error) {
	if len(value) != 7 || value[0] != '#' {
		return colorRGB{}, "", fmt.Errorf("invalid hex color %q", value)
	}
	r, err := strconv.ParseUint(value[1:3], 16, 8)
	if err != nil {
		return colorRGB{}, "", fmt.Errorf("invalid hex color %q", value)
	}
	g, err := strconv.ParseUint(value[3:5], 16, 8)
	if err != nil {
		return colorRGB{}, "", fmt.Errorf("invalid hex color %q", value)
	}
	b, err := strconv.ParseUint(value[5:7], 16, 8)
	if err != nil {
		return colorRGB{}, "", fmt.Errorf("invalid hex color %q", value)
	}
	hex := fmt.Sprintf("#%02x%02x%02x", r, g, b)
	return colorRGB{r: float64(r), g: float64(g), b: float64(b)}, hex, nil
}

func contrastRatio(a, b colorRGB) float64 {
	l1 := relativeLuminance(a)
	l2 := relativeLuminance(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func relativeLuminance(c colorRGB) float64 {
	r := channelToLinear(c.r)
	g := channelToLinear(c.g)
	b := channelToLinear(c.b)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func channelToLinear(value float64) float64 {
	srgb := value / 255.0
	if srgb <= 0.03928 {
		return srgb / 12.92
	}
	return math.Pow((srgb+0.055)/1.055, 2.4)
}
