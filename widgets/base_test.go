package widgets

import "testing"

func TestBase_HoverState(t *testing.T) {
	b := &Base{}
	if b.IsHovered() {
		t.Fatal("should not be hovered initially")
	}
	b.SetHovered(true)
	if !b.IsHovered() {
		t.Fatal("should be hovered after SetHovered(true)")
	}
	state := b.StyleState()
	if !state.Hovered {
		t.Fatal("StyleState should report Hovered=true")
	}
	b.SetHovered(false)
	if b.IsHovered() {
		t.Fatal("should not be hovered after SetHovered(false)")
	}
}

func TestBase_ActiveState(t *testing.T) {
	b := &Base{}
	if b.IsActive() {
		t.Fatal("should not be active initially")
	}
	b.SetActive(true)
	if !b.IsActive() {
		t.Fatal("should be active after SetActive(true)")
	}
	state := b.StyleState()
	if !state.Active {
		t.Fatal("StyleState should report Active=true")
	}
	b.SetActive(false)
	if b.IsActive() {
		t.Fatal("should not be active after SetActive(false)")
	}
}

func TestBase_HoverActiveNil(t *testing.T) {
	var b *Base
	b.SetHovered(true) // should not panic
	b.SetActive(true)  // should not panic
	if b.IsHovered() {
		t.Fatal("nil base should not be hovered")
	}
	if b.IsActive() {
		t.Fatal("nil base should not be active")
	}
}
