package viewer

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/docs/site/content"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/widgets"
)

type navItem struct {
	Title     string
	Path      string
	DocID     string
	Depth     int
	Section   bool
	DocTitle  string
	Summary   string
	HeadingID string
}

// DocsView renders the documentation navigation and content pane.
type DocsView struct {
	widgets.Component
	site          *content.SiteContent
	header        *widgets.Label
	status        *widgets.Label
	searchInput   *widgets.Input
	navItems      *state.Signal[[]navItem]
	navList       *widgets.List[navItem]
	searchItems   *state.Signal[[]navItem]
	searchList    *widgets.List[navItem]
	docView       *MarkdownView
	activeDoc     *content.Doc
	showingSearch bool
}

// NewDocsView builds a docs viewer for site content.
func NewDocsView(site *content.SiteContent) *DocsView {
	view := &DocsView{
		site:        site,
		navItems:    state.NewSignal([]navItem{}),
		searchItems: state.NewSignal([]navItem{}),
	}
	view.header = widgets.NewLabel("FluffyUI Documentation").WithStyle(backend.DefaultStyle().Bold(true))
	view.status = widgets.NewLabel("Ready")
	view.searchInput = widgets.NewInput()
	view.searchInput.SetPlaceholder("Search docs...")
	view.searchInput.OnChange(view.onSearchChanged)
	view.searchInput.SetLabel("Search")

	view.navList = widgets.NewList(widgets.NewSignalAdapter(view.navItems, renderNavItem))
	view.navList.SetLabel("Documentation Navigation")
	view.navList.OnSelect(func(index int, item navItem) {
		view.selectDoc(item.DocID)
	})

	view.searchList = widgets.NewList(widgets.NewSignalAdapter(view.searchItems, renderSearchItem))
	view.searchList.SetLabel("Search Results")
	view.searchList.OnSelect(func(index int, item navItem) {
		view.selectDocAnchor(item.DocID, item.HeadingID)
	})

	view.docView = NewMarkdownView(nil)
	view.docView.SetLabel("Documentation Content")

	view.loadNav()
	view.selectInitialDoc()
	return view
}

func (d *DocsView) loadNav() {
	if d.site == nil || d.site.Root == nil {
		return
	}
	items := flattenNav(d.site.Root, 0)
	d.navItems.Set(items)
}

func (d *DocsView) selectInitialDoc() {
	if d.site == nil {
		return
	}
	for i, item := range d.navItems.Get() {
		if item.DocID != "" {
			d.selectDoc(item.DocID)
			d.navList.SetSelected(i)
			return
		}
	}
}

// Measure fills available space.
func (d *DocsView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

// Layout positions header, navigation, and content.
func (d *DocsView) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
	y := bounds.Y
	headerHeight := 1
	statusHeight := 1
	if d.header != nil {
		d.header.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: headerHeight})
		y += headerHeight
	}
	navWidth := max(24, bounds.Width/4)
	if navWidth > bounds.Width-20 {
		navWidth = max(20, bounds.Width/3)
	}
	contentWidth := bounds.Width - navWidth - 1
	if contentWidth < 10 {
		contentWidth = max(10, bounds.Width-navWidth)
	}
	availableHeight := bounds.Height - statusHeight - (y - bounds.Y)
	if availableHeight < 0 {
		availableHeight = 0
	}
	navHeight := availableHeight
	searchHeight := 1
	listHeight := navHeight - searchHeight
	if listHeight < 0 {
		listHeight = 0
	}
	if d.searchInput != nil {
		d.searchInput.Layout(runtime.Rect{X: bounds.X, Y: y, Width: navWidth, Height: searchHeight})
	}
	listBounds := runtime.Rect{X: bounds.X, Y: y + searchHeight, Width: navWidth, Height: listHeight}
	if d.showingSearch {
		d.searchList.Layout(listBounds)
	} else {
		d.navList.Layout(listBounds)
	}

	contentBounds := runtime.Rect{X: bounds.X + navWidth + 1, Y: y, Width: contentWidth, Height: navHeight}
	if d.docView != nil {
		d.docView.Layout(contentBounds)
	}

	if d.status != nil {
		d.status.Layout(runtime.Rect{X: bounds.X, Y: bounds.Y + bounds.Height - statusHeight, Width: bounds.Width, Height: statusHeight})
	}
}

// Render draws the docs UI.
func (d *DocsView) Render(ctx runtime.RenderContext) {
	if d.header != nil {
		d.header.Render(ctx)
	}
	if d.searchInput != nil {
		d.searchInput.Render(ctx)
	}
	if d.showingSearch {
		d.searchList.Render(ctx)
	} else {
		d.navList.Render(ctx)
	}
	if d.docView != nil {
		d.docView.Render(ctx)
	}
	if d.status != nil {
		d.status.Render(ctx)
	}
	d.drawDivider(ctx)
}

// HandleMessage forwards input to children.
func (d *DocsView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for _, child := range d.ChildWidgets() {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns focusable children.
func (d *DocsView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if d.searchInput != nil {
		children = append(children, d.searchInput)
	}
	if d.showingSearch {
		children = append(children, d.searchList)
	} else {
		children = append(children, d.navList)
	}
	if d.docView != nil {
		children = append(children, d.docView)
	}
	return children
}

func (d *DocsView) onSearchChanged(text string) {
	query := strings.TrimSpace(text)
	if query == "" {
		d.showingSearch = false
		d.searchItems.Set(nil)
		d.updateStatus("Navigation")
		d.Invalidate()
		return
	}
	d.showingSearch = true
	if d.site == nil || d.site.Index == nil {
		d.searchItems.Set(nil)
		d.updateStatus("No search index")
		d.Invalidate()
		return
	}
	hits := d.site.Index.Search(query, 50)
	results := make([]navItem, 0, len(hits))
	for _, hit := range hits {
		title := hit.Entry.Heading
		if title == "" {
			title = hit.Entry.DocTitle
		}
		results = append(results, navItem{
			Title:     title,
			DocID:     hit.Entry.DocID,
			HeadingID: hit.Entry.HeadingID,
			DocTitle:  hit.Entry.DocTitle,
			Summary:   hit.Entry.Text,
		})
	}
	d.searchItems.Set(results)
	d.updateStatus("Search results: " + itoa(len(results)))
	d.Invalidate()
}

func (d *DocsView) selectDoc(docID string) {
	if d.site == nil || docID == "" {
		return
	}
	doc, ok := d.site.Docs[docID]
	if !ok {
		return
	}
	d.activeDoc = doc
	if d.docView != nil {
		d.docView.SetLines(doc.Lines)
		d.docView.ScrollToStart()
	}
	if d.header != nil {
		d.header.SetText("FluffyUI Documentation - " + doc.Title)
	}
	d.updateStatus(doc.Path)
	d.Invalidate()
}

func (d *DocsView) selectDocAnchor(docID, anchor string) {
	d.selectDoc(docID)
	if anchor != "" && d.docView != nil {
		d.docView.ScrollToAnchor(anchor)
	}
}

func (d *DocsView) updateStatus(text string) {
	if d.status != nil {
		d.status.SetText(text)
	}
}

func (d *DocsView) drawDivider(ctx runtime.RenderContext) {
	if d == nil {
		return
	}
	bounds := d.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	navWidth := max(24, bounds.Width/4)
	if navWidth > bounds.Width-20 {
		navWidth = max(20, bounds.Width/3)
	}
	x := bounds.X + navWidth
	style := backend.DefaultStyle().Dim(true)
	for y := bounds.Y + 1; y < bounds.Y+bounds.Height-1; y++ {
		ctx.Buffer.Set(x, y, '|', style)
	}
}

func renderNavItem(item navItem, index int, selected bool, ctx runtime.RenderContext) {
	label := item.Title
	prefix := "- "
	if item.Section && item.DocID == "" {
		prefix = "+ "
	}
	if item.Depth > 0 {
		label = strings.Repeat("  ", item.Depth) + prefix + label
	} else {
		label = prefix + label
	}
	drawListLine(label, ctx)
}

func renderSearchItem(item navItem, index int, selected bool, ctx runtime.RenderContext) {
	label := item.DocTitle
	if item.HeadingID != "" && item.Title != "" {
		label = item.DocTitle + " - " + item.Title
	}
	drawListLine(label, ctx)
}

func drawListLine(text string, ctx runtime.RenderContext) {
	bounds := ctx.Bounds
	if bounds.Width <= 0 {
		return
	}
	truncated := truncateText(text, bounds.Width)
	ctx.Buffer.SetString(bounds.X, bounds.Y, truncated, backend.DefaultStyle())
}

func flattenNav(root *content.Node, depth int) []navItem {
	if root == nil {
		return nil
	}
	var items []navItem
	for _, child := range root.Children {
		if child == nil {
			continue
		}
		item := navItem{
			Title:   child.Title,
			Path:    child.Path,
			DocID:   "",
			Depth:   depth,
			Section: child.Doc == nil,
		}
		if child.Doc != nil {
			item.DocID = child.Doc.ID
			item.Summary = child.Doc.Summary
		}
		items = append(items, item)
		items = append(items, flattenNav(child, depth+1)...)
	}
	return items
}

func truncateText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(text) <= width {
		return text
	}
	return runewidth.Truncate(text, width, "")
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var out [20]byte
	pos := len(out)
	for v > 0 {
		pos--
		out[pos] = byte('0' + (v % 10))
		v /= 10
	}
	return string(out[pos:])
}

var _ runtime.Widget = (*DocsView)(nil)
var _ runtime.ChildProvider = (*DocsView)(nil)
