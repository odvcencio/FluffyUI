package content

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/odvcencio/fluffy-ui/markdown"
	"github.com/odvcencio/fluffy-ui/theme"
	"github.com/yuin/goldmark/ast"
	"gopkg.in/yaml.v3"
)

// LoadOptions configures how docs are loaded.
type LoadOptions struct {
	Renderer   *markdown.Renderer
	Source     string
	Ignore     []string
	Extensions []string
}

// SiteContent contains the docs tree, docs map, and search index.
type SiteContent struct {
	Root  *Node
	Docs  map[string]*Doc
	Index *SearchIndex
}

// Node represents a navigation tree node.
type Node struct {
	ID       string
	Title    string
	Path     string
	Order    int
	Doc      *Doc
	Children []*Node
}

// Doc represents a single markdown document.
type Doc struct {
	ID       string
	Path     string
	RelPath  string
	Title    string
	Summary  string
	Meta     DocMeta
	Content  string
	Headings []Heading
	Lines    []markdown.StyledLine
}

// DocMeta describes optional front matter fields.
type DocMeta struct {
	Title   string   `yaml:"title"`
	Summary string   `yaml:"summary"`
	Order   int      `yaml:"order"`
	Tags    []string `yaml:"tags"`
}

// Heading represents a markdown heading for TOC/anchors.
type Heading struct {
	Level int
	Text  string
	ID    string
}

// LoadDir loads docs from an OS directory path.
func LoadDir(root string, opts LoadOptions) (*SiteContent, error) {
	if root == "" {
		return nil, errors.New("missing docs root")
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("docs root is not a directory: %s", root)
	}
	return LoadFS(os.DirFS(root), ".", opts)
}

// LoadFS loads docs from an fs.FS rooted at base.
func LoadFS(fsys fs.FS, base string, opts LoadOptions) (*SiteContent, error) {
	options := opts.withDefaults()
	renderer := options.Renderer
	if renderer == nil {
		renderer = markdown.NewRenderer(theme.DefaultTheme())
	}
	parser := markdown.NewParser()

	root := &Node{ID: "", Title: "Docs", Path: "/"}
	docs := map[string]*Doc{}
	index := NewSearchIndex()

	err := fs.WalkDir(fsys, base, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path == base {
				return nil
			}
			if shouldIgnore(options, path, true) {
				return fs.SkipDir
			}
			return nil
		}
		if shouldIgnore(options, path, false) {
			return nil
		}
		if !hasAllowedExt(path, options.Extensions) {
			return nil
		}
		raw, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		meta, body, err := parseFrontMatter(string(raw))
		if err != nil {
			return fmt.Errorf("parse front matter (%s): %w", path, err)
		}
		relPath := strings.TrimPrefix(filepath.ToSlash(path), "./")
		docID, route := docIDFromPath(relPath)
		doc := &Doc{
			ID:      docID,
			Path:    route,
			RelPath: relPath,
			Meta:    meta,
			Content: body,
		}
		docAST := parser.ParseString(body)
		doc.Headings = extractHeadings(docAST, []byte(body))
		doc.Title = resolveDocTitle(doc, relPath)
		doc.Summary = resolveDocSummary(doc, docAST, []byte(body))
		doc.Lines = renderer.RenderAST(options.Source, docAST, body)

		docs[docID] = doc
		attachNode(root, doc, relPath)
		index.AddEntries(entriesForDoc(doc))
		return nil
	})
	if err != nil {
		return nil, err
	}

	sortNodes(root)
	return &SiteContent{Root: root, Docs: docs, Index: index}, nil
}

func (o LoadOptions) withDefaults() LoadOptions {
	out := o
	if out.Extensions == nil {
		out.Extensions = []string{".md", ".markdown"}
	}
	if out.Ignore == nil {
		out.Ignore = []string{"site", "demos"}
	}
	return out
}

func hasAllowedExt(path string, exts []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowed := range exts {
		if ext == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

func shouldIgnore(opts LoadOptions, path string, isDir bool) bool {
	rel := strings.TrimPrefix(filepath.ToSlash(path), "./")
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, ".") || strings.HasPrefix(part, "_") {
			return true
		}
	}
	if !isDir && len(parts) > 0 {
		for _, prefix := range opts.Ignore {
			prefix = strings.Trim(prefix, "/")
			if prefix == "" {
				continue
			}
			if strings.HasPrefix(rel, prefix+"/") || rel == prefix {
				return true
			}
		}
		return false
	}
	for _, prefix := range opts.Ignore {
		prefix = strings.Trim(prefix, "/")
		if prefix == "" {
			continue
		}
		if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
			return true
		}
	}
	return false
}

func parseFrontMatter(content string) (DocMeta, string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return DocMeta{}, content, nil
	}
	if strings.TrimSpace(lines[0]) != "---" {
		return DocMeta{}, content, nil
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return DocMeta{}, content, nil
	}
	metaText := strings.Join(lines[1:end], "\n")
	body := strings.Join(lines[end+1:], "\n")
	var meta DocMeta
	if err := yaml.Unmarshal([]byte(metaText), &meta); err != nil {
		return DocMeta{}, content, err
	}
	return meta, body, nil
}

func docIDFromPath(rel string) (string, string) {
	rel = strings.TrimPrefix(rel, "./")
	rel = filepath.ToSlash(rel)
	ext := filepath.Ext(rel)
	id := strings.TrimSuffix(rel, ext)
	if id == "index" {
		return id, "/"
	}
	if strings.HasSuffix(id, "/index") {
		id = strings.TrimSuffix(id, "/index")
		if id == "" {
			return "index", "/"
		}
		return id, "/" + id
	}
	return id, "/" + id
}

func resolveDocTitle(doc *Doc, relPath string) string {
	if doc != nil && doc.Meta.Title != "" {
		return doc.Meta.Title
	}
	if doc != nil && len(doc.Headings) > 0 && doc.Headings[0].Level == 1 {
		return doc.Headings[0].Text
	}
	base := strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath))
	return titleFromSegment(base)
}

func resolveDocSummary(doc *Doc, root ast.Node, source []byte) string {
	if doc != nil && doc.Meta.Summary != "" {
		return doc.Meta.Summary
	}
	if root != nil {
		if summary := firstParagraphText(root, source); summary != "" {
			return summary
		}
	}
	if doc != nil {
		text := plainTextFromLines(doc.Lines)
		return trimToLength(text, 160)
	}
	return ""
}

func attachNode(root *Node, doc *Doc, relPath string) {
	if root == nil || doc == nil {
		return
	}
	id := doc.ID
	segments := strings.Split(id, "/")
	current := root
	pathParts := []string{}
	for i, seg := range segments {
		if seg == "" {
			continue
		}
		pathParts = append(pathParts, seg)
		nodeID := strings.Join(pathParts, "/")
		child := findChild(current, nodeID)
		if child == nil {
			child = &Node{
				ID:    nodeID,
				Title: titleFromSegment(seg),
				Path:  "/" + nodeID,
				Order: 0,
			}
			current.Children = append(current.Children, child)
		}
		if i == len(segments)-1 {
			child.Doc = doc
			child.Title = doc.Title
			child.Order = doc.Meta.Order
		}
		current = child
	}
}

func findChild(node *Node, id string) *Node {
	for _, child := range node.Children {
		if child.ID == id {
			return child
		}
	}
	return nil
}

func sortNodes(node *Node) {
	if node == nil {
		return
	}
	sort.Slice(node.Children, func(i, j int) bool {
		left := node.Children[i]
		right := node.Children[j]
		if left.Order != right.Order {
			return left.Order < right.Order
		}
		return strings.ToLower(left.Title) < strings.ToLower(right.Title)
	})
	for _, child := range node.Children {
		sortNodes(child)
	}
}

func titleFromSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return ""
	}
	segment = strings.ReplaceAll(segment, "-", " ")
	segment = strings.ReplaceAll(segment, "_", " ")
	parts := strings.Fields(segment)
	for i, part := range parts {
		if len(part) == 1 {
			parts[i] = strings.ToUpper(part)
		} else {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}
