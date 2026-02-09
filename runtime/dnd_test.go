package runtime

import "testing"

func TestDragData_Fields(t *testing.T) {
	w := &testSimpleWidget{}
	data := DragData{
		Source: w,
		Kind:   "list-item",
		Value:  "hello",
		Label:  "Hello Item",
	}
	if data.Source != w {
		t.Fatal("expected Source to be the widget")
	}
	if data.Kind != "list-item" {
		t.Fatalf("expected Kind 'list-item', got %q", data.Kind)
	}
	if data.Value != "hello" {
		t.Fatalf("expected Value 'hello', got %v", data.Value)
	}
	if data.Label != "Hello Item" {
		t.Fatalf("expected Label 'Hello Item', got %q", data.Label)
	}
}

func TestDragStartMsg_IsCommand(t *testing.T) {
	msg := DragStartMsg{Data: DragData{Kind: "test"}}
	// Verify it implements the Command interface.
	var cmd Command = msg
	cmd.Command()
	if msg.Data.Kind != "test" {
		t.Fatalf("expected Kind 'test', got %q", msg.Data.Kind)
	}
}

func TestDragEndMsg_IsCommand(t *testing.T) {
	msg := DragEndMsg{
		Data:      DragData{Kind: "test", Label: "item"},
		Target:    nil,
		Cancelled: true,
	}
	var cmd Command = msg
	cmd.Command()
	if !msg.Cancelled {
		t.Fatal("expected Cancelled to be true")
	}
	if msg.Target != nil {
		t.Fatal("expected Target to be nil for cancelled drag")
	}
}

func TestDragEndMsg_WithTarget(t *testing.T) {
	w := &testSimpleWidget{}
	msg := DragEndMsg{
		Data:      DragData{Kind: "tree-node", Value: 42, Label: "Node 42"},
		Target:    w,
		Cancelled: false,
	}
	if msg.Cancelled {
		t.Fatal("expected Cancelled to be false")
	}
	if msg.Target != w {
		t.Fatal("expected Target to be the widget")
	}
	if msg.Data.Value != 42 {
		t.Fatalf("expected Value 42, got %v", msg.Data.Value)
	}
}

func TestDragData_EmptyFields(t *testing.T) {
	data := DragData{}
	if data.Source != nil {
		t.Fatal("expected nil Source")
	}
	if data.Kind != "" {
		t.Fatal("expected empty Kind")
	}
	if data.Value != nil {
		t.Fatal("expected nil Value")
	}
	if data.Label != "" {
		t.Fatal("expected empty Label")
	}
}

func TestDragStartMsg_InCommandsList(t *testing.T) {
	// Verify DragStartMsg can be used in HandleResult commands.
	data := DragData{Kind: "list-item", Label: "A"}
	result := WithCommand(DragStartMsg{Data: data})
	if !result.Handled {
		t.Fatal("expected handled")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	startMsg, ok := result.Commands[0].(DragStartMsg)
	if !ok {
		t.Fatalf("expected DragStartMsg, got %T", result.Commands[0])
	}
	if startMsg.Data.Kind != "list-item" {
		t.Fatalf("expected Kind 'list-item', got %q", startMsg.Data.Kind)
	}
}

func TestDragEndMsg_InCommandsList(t *testing.T) {
	data := DragData{Kind: "list-item", Label: "B"}
	result := WithCommand(DragEndMsg{Data: data, Cancelled: false, Target: &testSimpleWidget{}})
	if !result.Handled {
		t.Fatal("expected handled")
	}
	endMsg, ok := result.Commands[0].(DragEndMsg)
	if !ok {
		t.Fatalf("expected DragEndMsg, got %T", result.Commands[0])
	}
	if endMsg.Data.Label != "B" {
		t.Fatalf("expected Label 'B', got %q", endMsg.Data.Label)
	}
}

func TestApp_DragStateTracking(t *testing.T) {
	app := &App{}

	// Initially no drag active.
	if app.DragActive() {
		t.Fatal("expected no active drag initially")
	}
	if app.DragData() != nil {
		t.Fatal("expected nil drag data initially")
	}

	// Simulate DragStartMsg.
	w := &testSimpleWidget{}
	startData := DragData{Source: w, Kind: "list-item", Value: "test", Label: "Test"}
	app.handleCommand(DragStartMsg{Data: startData})

	if !app.DragActive() {
		t.Fatal("expected drag to be active after DragStartMsg")
	}
	if app.DragData() == nil {
		t.Fatal("expected non-nil drag data after DragStartMsg")
	}
	if app.DragData().Kind != "list-item" {
		t.Fatalf("expected Kind 'list-item', got %q", app.DragData().Kind)
	}

	// Simulate DragEndMsg (cancelled).
	app.handleCommand(DragEndMsg{Data: startData, Cancelled: true})

	if app.DragActive() {
		t.Fatal("expected no active drag after DragEndMsg")
	}
	if app.DragData() != nil {
		t.Fatal("expected nil drag data after DragEndMsg")
	}
}

func TestApp_DragStateTracking_Drop(t *testing.T) {
	app := &App{}

	w := &testSimpleWidget{}
	startData := DragData{Source: w, Kind: "tree-node", Value: 99, Label: "Node"}
	app.handleCommand(DragStartMsg{Data: startData})

	if !app.DragActive() {
		t.Fatal("expected drag active")
	}

	// Complete drop (not cancelled).
	app.handleCommand(DragEndMsg{Data: startData, Target: w, Cancelled: false})

	if app.DragActive() {
		t.Fatal("expected drag inactive after drop")
	}
}

func TestApp_DragActive_NilApp(t *testing.T) {
	var app *App
	if app.DragActive() {
		t.Fatal("expected false for nil app")
	}
	if app.DragData() != nil {
		t.Fatal("expected nil for nil app")
	}
}
