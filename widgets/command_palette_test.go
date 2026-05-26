package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func testPaletteCommands() []PaletteCommand {
	return []PaletteCommand{
		{ID: "save", Label: "Save", Description: "Save file", Shortcut: "Ctrl+S", Category: "File"},
		{ID: "open", Label: "Open", Description: "Open file", Shortcut: "Ctrl+O", Category: "File"},
		{ID: "quit", Label: "Quit", Description: "Quit application", Shortcut: "Ctrl+Q", Category: "App"},
		{ID: "search", Label: "Search", Description: "Find text", Shortcut: "Ctrl+F", Category: "Edit"},
		{ID: "saveas", Label: "Save As", Description: "Save file as", Shortcut: "Ctrl+Shift+S", Category: "File"},
	}
}

func TestCommandPalette_New(t *testing.T) {
	cmds := testPaletteCommands()
	cp := NewCommandPalette(cmds...)
	if cp == nil {
		t.Fatal("expected non-nil command palette")
	}
	if cp.AccessibleRole() != "combobox" {
		t.Fatalf("expected role combobox, got %q", cp.AccessibleRole())
	}
	if !cp.CanFocus() {
		t.Fatal("expected command palette to be focusable")
	}
	if cp.Open() {
		t.Fatal("expected palette to start closed")
	}
	if cp.maxResults != 10 {
		t.Fatalf("expected default maxResults 10, got %d", cp.maxResults)
	}
}

func TestCommandPalette_ShowHideToggle(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)

	// Show
	cp.Show()
	if !cp.Open() {
		t.Fatal("expected palette to be open after Show()")
	}
	if cp.Query() != "" {
		t.Fatal("expected query to be reset on Show()")
	}

	// Hide
	cp.Hide()
	if cp.Open() {
		t.Fatal("expected palette to be closed after Hide()")
	}

	// Toggle from closed -> open
	cp.Toggle()
	if !cp.Open() {
		t.Fatal("expected palette to be open after Toggle() from closed")
	}

	// Toggle from open -> closed
	cp.Toggle()
	if cp.Open() {
		t.Fatal("expected palette to be closed after Toggle() from open")
	}
}

func TestCommandPalette_SetOpenCompat(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.SetOpen(true)
	if !cp.Open() {
		t.Fatal("expected palette to be open via SetOpen(true)")
	}
	cp.SetOpen(false)
	if cp.Open() {
		t.Fatal("expected palette to be closed via SetOpen(false)")
	}
}

func TestCommandPalette_FilteredCount(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	if cp.FilteredCount() != 5 {
		t.Fatalf("expected 5 filtered commands, got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_EmptyQueryShowsAll(t *testing.T) {
	cmds := make([]PaletteCommand, 15)
	for i := range cmds {
		cmds[i] = PaletteCommand{ID: string(rune('a' + i)), Label: string(rune('A' + i))}
	}
	cp := NewCommandPalette(cmds...)
	cp.Show()

	// Empty query should show all commands (filtering returns all)
	if cp.FilteredCount() != 15 {
		t.Fatalf("expected 15 filtered commands with empty query, got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_EmptyQueryRespectedByMaxResults(t *testing.T) {
	cmds := make([]PaletteCommand, 15)
	for i := range cmds {
		cmds[i] = PaletteCommand{ID: string(rune('a' + i)), Label: string(rune('A' + i))}
	}
	cp := NewCommandPalette(cmds...)
	cp.SetMaxResults(5)
	cp.Show()

	// FilteredCount returns all matches, but rendering would cap at maxResults
	if cp.FilteredCount() != 15 {
		t.Fatalf("expected 15 filtered (rendering caps at maxResults), got %d", cp.FilteredCount())
	}
	// Verify maxResults was set
	if cp.maxResults != 5 {
		t.Fatalf("expected maxResults 5, got %d", cp.maxResults)
	}
}

func TestCommandPalette_FuzzyFilteringOnLabel(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	// Type "sa" to filter - should match "Save" and "Save As" by label
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 's'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})
	if cp.Query() != "sa" {
		t.Fatalf("expected query 'sa', got %q", cp.Query())
	}

	// Should match "Save" and "Save As" (label contains "sa")
	// "Search" label does not contain "sa"
	if cp.FilteredCount() != 2 {
		t.Fatalf("expected 2 filtered commands for 'sa', got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_FuzzyFilteringOnDescription(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	// Type "find" - should match "Search" because its Description is "Find text"
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'f'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'i'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'n'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'd'})

	if cp.FilteredCount() != 1 {
		t.Fatalf("expected 1 filtered command for 'find' (description match), got %d", cp.FilteredCount())
	}
	if cp.filtered[0].ID != "search" {
		t.Fatalf("expected 'search' command, got %q", cp.filtered[0].ID)
	}
}

func TestCommandPalette_FuzzyFilteringPrefixFirst(t *testing.T) {
	cmds := []PaletteCommand{
		{ID: "close", Label: "Close Tab", Description: "Close current tab"},
		{ID: "open", Label: "Open File", Description: "Open a file"},
		{ID: "openrecent", Label: "Open Recent", Description: "Open recently used"},
	}
	cp := NewCommandPalette(cmds...)
	cp.Show()
	cp.Focus()

	// Type "open" - should match "Open File" and "Open Recent" (prefix) and
	// "Close Tab" should NOT match. Prefix matches should come first.
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'o'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'p'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'e'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'n'})

	if cp.FilteredCount() != 2 {
		t.Fatalf("expected 2 commands matching 'open', got %d", cp.FilteredCount())
	}
	// Both are prefix matches so order should be preserved (Open File, Open Recent)
	if cp.filtered[0].ID != "open" {
		t.Fatalf("expected first match to be 'open', got %q", cp.filtered[0].ID)
	}
	if cp.filtered[1].ID != "openrecent" {
		t.Fatalf("expected second match to be 'openrecent', got %q", cp.filtered[1].ID)
	}
}

func TestCommandPalette_FuzzyFilteringContainsAfterPrefix(t *testing.T) {
	cmds := []PaletteCommand{
		{ID: "refile", Label: "Refile Note", Description: "Move note to another file"},
		{ID: "file-new", Label: "File New", Description: "Create a new file"},
		{ID: "profile", Label: "Edit Profile", Description: "Edit user profile"},
	}
	cp := NewCommandPalette(cmds...)
	cp.Show()
	cp.Focus()

	// Type "file" - "File New" is prefix match (score 3), "Refile Note" contains in label (score 2),
	// "Edit Profile" contains "file" in label via "Profile" (score 2)
	for _, r := range "file" {
		cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: r})
	}

	if cp.FilteredCount() != 3 {
		t.Fatalf("expected 3 commands matching 'file', got %d", cp.FilteredCount())
	}
	// Prefix match should come first
	if cp.filtered[0].ID != "file-new" {
		t.Fatalf("expected first match to be prefix 'file-new', got %q", cp.filtered[0].ID)
	}
}

func TestCommandPalette_BackspaceDeletesQuery(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 's'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})
	if cp.Query() != "sa" {
		t.Fatalf("expected query 'sa', got %q", cp.Query())
	}

	// Backspace
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if cp.Query() != "s" {
		t.Fatalf("expected query 's' after backspace, got %q", cp.Query())
	}
	// "s" matches Save, Search, Save As
	if cp.FilteredCount() != 3 {
		t.Fatalf("expected 3 filtered commands for 's', got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_BackspaceOnEmptyDoesNothing(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	// Backspace on empty query should still be handled but no crash
	result := cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if !result.Handled {
		t.Fatal("expected backspace to be handled even on empty query")
	}
	if cp.Query() != "" {
		t.Fatalf("expected empty query after backspace on empty, got %q", cp.Query())
	}
}

func TestCommandPalette_KeyNavigation(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected initial selected 0, got %d", cp.SelectedIndex())
	}

	// Down
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if cp.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1 after down, got %d", cp.SelectedIndex())
	}

	// Down again
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if cp.SelectedIndex() != 2 {
		t.Fatalf("expected selected 2 after second down, got %d", cp.SelectedIndex())
	}

	// Up
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if cp.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1 after up, got %d", cp.SelectedIndex())
	}

	// Up
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0 after up, got %d", cp.SelectedIndex())
	}

	// Up at 0 should stay at 0
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selected still 0, got %d", cp.SelectedIndex())
	}
}

func TestCommandPalette_KeyNavigationBoundsBottom(t *testing.T) {
	cmds := []PaletteCommand{
		{ID: "a", Label: "A"},
		{ID: "b", Label: "B"},
	}
	cp := NewCommandPalette(cmds...)
	cp.Show()
	cp.Focus()

	// Go to last item
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if cp.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1, got %d", cp.SelectedIndex())
	}

	// Try to go past last item
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if cp.SelectedIndex() != 1 {
		t.Fatalf("expected selected clamped at 1, got %d", cp.SelectedIndex())
	}
}

func TestCommandPalette_ExecuteCommand(t *testing.T) {
	executed := false
	cmds := []PaletteCommand{
		{ID: "test", Label: "Test", OnExecute: func() { executed = true }},
	}
	cp := NewCommandPalette(cmds...)
	cp.Show()
	cp.Focus()

	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !executed {
		t.Fatal("expected command OnExecute to fire")
	}
	if cp.Open() {
		t.Fatal("expected palette to close after execution")
	}
}

func TestCommandPalette_ExecuteCallsOnExecuteCallback(t *testing.T) {
	var executedCmd PaletteCommand
	cmds := []PaletteCommand{
		{ID: "cmd1", Label: "Command One"},
		{ID: "cmd2", Label: "Command Two"},
	}
	cp := NewCommandPalette(cmds...)
	cp.SetOnExecute(func(cmd PaletteCommand) {
		executedCmd = cmd
	})
	cp.Show()
	cp.Focus()

	// Navigate to second command
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if executedCmd.ID != "cmd2" {
		t.Fatalf("expected onExecute callback with cmd2, got %q", executedCmd.ID)
	}
}

func TestCommandPalette_EscapeCloses(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if cp.Open() {
		t.Fatal("expected palette to close on escape")
	}
}

func TestCommandPalette_SetCommands(t *testing.T) {
	cp := NewCommandPalette()
	cp.Show()

	if cp.FilteredCount() != 0 {
		t.Fatalf("expected 0 commands initially, got %d", cp.FilteredCount())
	}

	cp.SetCommands([]PaletteCommand{
		{ID: "a", Label: "Alpha"},
		{ID: "b", Label: "Beta"},
	})

	if cp.FilteredCount() != 2 {
		t.Fatalf("expected 2 commands after SetCommands, got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_AddCommand(t *testing.T) {
	cp := NewCommandPalette(
		PaletteCommand{ID: "a", Label: "Alpha"},
	)
	cp.Show()

	if cp.FilteredCount() != 1 {
		t.Fatalf("expected 1 command, got %d", cp.FilteredCount())
	}

	cp.AddCommand(PaletteCommand{ID: "b", Label: "Beta"})

	if cp.FilteredCount() != 2 {
		t.Fatalf("expected 2 commands after AddCommand, got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_SetMaxResults(t *testing.T) {
	cp := NewCommandPalette()
	if cp.maxResults != 10 {
		t.Fatalf("expected default maxResults 10, got %d", cp.maxResults)
	}

	cp.SetMaxResults(5)
	if cp.maxResults != 5 {
		t.Fatalf("expected maxResults 5, got %d", cp.maxResults)
	}

	// Minimum should be 1
	cp.SetMaxResults(0)
	if cp.maxResults != 1 {
		t.Fatalf("expected maxResults clamped to 1, got %d", cp.maxResults)
	}
}

func TestCommandPalette_Measure(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	size := cp.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 40})
	if size.Width < 10 {
		t.Fatalf("expected width >= 10, got %d", size.Width)
	}
	if size.Height < 4 {
		t.Fatalf("expected height >= 4, got %d", size.Height)
	}
}

func TestCommandPalette_MeasureClosed(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	size := cp.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 40})
	if size.Width != 0 || size.Height != 0 {
		t.Fatalf("expected zero size when closed, got %dx%d", size.Width, size.Height)
	}
}

func TestCommandPalette_StyleType(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	if cp.StyleType() != "CommandPalette" {
		t.Fatalf("expected StyleType 'CommandPalette', got %q", cp.StyleType())
	}
}

func TestCommandPalette_UnfocusedNoHandle(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	result := cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestCommandPalette_ClosedNoHandle(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Focus()
	// Palette is closed, messages should not be handled
	result := cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when palette is closed")
	}
}

func TestCommandPalette_A11yRoleCombobox(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	if cp.AccessibleRole() != accessibility.RoleCombobox {
		t.Fatalf("expected ARIA role %q, got %q", accessibility.RoleCombobox, cp.AccessibleRole())
	}
}

func TestCommandPalette_A11yHasPopupListbox(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	if cp.Base.HasPopup != "listbox" {
		t.Fatalf("expected HasPopup 'listbox', got %q", cp.Base.HasPopup)
	}
}

func TestCommandPalette_A11yExpandedState(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)

	// Initially closed
	if cp.Base.State.Expanded == nil || *cp.Base.State.Expanded != false {
		t.Fatal("expected expanded=false when closed")
	}

	cp.Show()
	if cp.Base.State.Expanded == nil || *cp.Base.State.Expanded != true {
		t.Fatal("expected expanded=true when open")
	}

	cp.Hide()
	if cp.Base.State.Expanded == nil || *cp.Base.State.Expanded != false {
		t.Fatal("expected expanded=false after close")
	}
}

func TestCommandPalette_A11yAutocomplete(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	if cp.Base.Autocomplete != "list" {
		t.Fatalf("expected Autocomplete 'list', got %q", cp.Base.Autocomplete)
	}
}

func TestCommandPalette_A11yPosInSetAndSetSize(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	if cp.Base.SetSize != 5 {
		t.Fatalf("expected SetSize 5, got %d", cp.Base.SetSize)
	}
	if cp.Base.PosInSet != 1 {
		t.Fatalf("expected PosInSet 1, got %d", cp.Base.PosInSet)
	}

	// Navigate down
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if cp.Base.PosInSet != 2 {
		t.Fatalf("expected PosInSet 2 after down, got %d", cp.Base.PosInSet)
	}
}

func TestCommandPalette_NilSafety(t *testing.T) {
	var cp *CommandPalette

	// None of these should panic
	cp.Show()
	cp.Hide()
	cp.Toggle()
	cp.SetOpen(true)
	cp.SetCommands(nil)
	cp.AddCommand(PaletteCommand{})
	cp.SetMaxResults(5)
	cp.SetOnExecute(nil)
	cp.SetStyle(backend.DefaultStyle())
	cp.Bind(runtime.Services{})
	cp.Unbind()

	if cp.Open() {
		t.Fatal("nil palette should not be open")
	}
	if cp.Query() != "" {
		t.Fatal("nil palette query should be empty")
	}
	if cp.SelectedIndex() != 0 {
		t.Fatal("nil palette selected index should be 0")
	}
	if cp.FilteredCount() != 0 {
		t.Fatal("nil palette filtered count should be 0")
	}
}

func TestCommandPalette_ShowResetsQueryAndSelection(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	// Type and navigate
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 's'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	if cp.Query() != "s" {
		t.Fatal("expected query to be 's'")
	}

	// Close and reopen - query and selection should reset
	cp.Hide()
	cp.Show()

	if cp.Query() != "" {
		t.Fatalf("expected query to be reset after Show(), got %q", cp.Query())
	}
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selection to be reset after Show(), got %d", cp.SelectedIndex())
	}
}

func TestCommandPalette_BindUnbind(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Bind(runtime.Services{})
	cp.Unbind()
	// Just verify no panic
}

func TestCommandPalette_CaseInsensitiveFiltering(t *testing.T) {
	cmds := []PaletteCommand{
		{ID: "upper", Label: "SAVE FILE"},
		{ID: "lower", Label: "save file"},
		{ID: "mixed", Label: "Save File"},
	}
	cp := NewCommandPalette(cmds...)
	cp.Show()
	cp.Focus()

	// Type lowercase "save"
	for _, r := range "save" {
		cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: r})
	}

	// All three should match (case-insensitive)
	if cp.FilteredCount() != 3 {
		t.Fatalf("expected 3 case-insensitive matches, got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_SelectionClampedOnFilter(t *testing.T) {
	cp := NewCommandPalette(testPaletteCommands()...)
	cp.Show()
	cp.Focus()

	// Navigate to last item (index 4)
	for i := 0; i < 4; i++ {
		cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	}
	if cp.SelectedIndex() != 4 {
		t.Fatalf("expected selected 4, got %d", cp.SelectedIndex())
	}

	// Type a query that reduces results - selection should clamp
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'q'})
	// "q" matches "Quit" only
	if cp.FilteredCount() != 1 {
		t.Fatalf("expected 1 match for 'q', got %d", cp.FilteredCount())
	}
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selection clamped to 0, got %d", cp.SelectedIndex())
	}
}
