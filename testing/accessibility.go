package testing

import (
	"fmt"
	"strings"
	"testing"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/runtime"
)

// A11ySeverity describes audit severity.
type A11ySeverity string

const (
	A11yError   A11ySeverity = "error"
	A11yWarning A11ySeverity = "warning"
)

// A11yIssue describes an accessibility audit issue.
type A11yIssue struct {
	Path     string
	Message  string
	Severity A11ySeverity
}

// AuditAccessibility walks the widget tree and reports missing accessibility metadata.
func AuditAccessibility(root runtime.Widget) []A11yIssue {
	if root == nil {
		return nil
	}
	var issues []A11yIssue
	var walk func(node runtime.Widget, path []string)
	walk = func(node runtime.Widget, path []string) {
		if node == nil {
			return
		}
		name := widgetName(node)
		currentPath := append(path, name)

		if accessible, ok := node.(accessibility.Accessible); ok {
			role := strings.TrimSpace(string(accessible.AccessibleRole()))
			label := strings.TrimSpace(accessible.AccessibleLabel())
			if role == "" {
				issues = append(issues, A11yIssue{
					Path:     strings.Join(currentPath, "/"),
					Message:  "missing accessible role",
					Severity: A11yError,
				})
			}
			if label == "" {
				severity := A11yWarning
				if focusable, ok := node.(runtime.Focusable); ok && focusable.CanFocus() {
					severity = A11yError
				}
				issues = append(issues, A11yIssue{
					Path:     strings.Join(currentPath, "/"),
					Message:  "missing accessible label",
					Severity: severity,
				})
			}
		}

		provider, ok := node.(runtime.ChildProvider)
		if !ok {
			return
		}
		for _, child := range provider.ChildWidgets() {
			if child == nil {
				continue
			}
			segment := widgetName(child)
			if segmenter, ok := node.(runtime.PathSegmenter); ok {
				if seg := strings.TrimSpace(segmenter.PathSegment(child)); seg != "" {
					segment = seg
				}
			}
			walk(child, append(path, segment))
		}
	}

	walk(root, nil)
	return issues
}

// AssertAccessible fails the test if accessibility errors are found.
func AssertAccessible(t *testing.T, root runtime.Widget) {
	t.Helper()
	issues := AuditAccessibility(root)
	for _, issue := range issues {
		if issue.Severity == A11yWarning {
			t.Logf("a11y warning: %s: %s", issue.Path, issue.Message)
			continue
		}
		t.Errorf("a11y error: %s: %s", issue.Path, issue.Message)
	}
}

func widgetName(w runtime.Widget) string {
	if w == nil {
		return ""
	}
	name := fmt.Sprintf("%T", w)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return strings.TrimPrefix(name, "*")
}

// ---------------------------------------------------------------------------
// Role validation: required properties, invalid states, container children
// ---------------------------------------------------------------------------

// roleRequiredProperties maps roles to properties that must be present.
// "value" means AccessibleValue() must be non-nil.
// "selected" means StateSet.Selected or Checked must be set.
// "checked" means StateSet.Checked must be non-nil.
var roleRequiredProperties = map[accessibility.Role][]string{
	accessibility.RoleSlider:      {"value"},
	accessibility.RoleProgressBar: {"value"},
	accessibility.RoleMeter:       {"value"},
	accessibility.RoleSpinButton:  {"value"},
	accessibility.RoleScrollbar:   {"value"},
	accessibility.RoleTab:         {"selected"},
	accessibility.RoleOption:      {"selected"},
	accessibility.RoleCheckbox:    {"checked"},
	accessibility.RoleRadio:       {"checked"},
	accessibility.RoleSwitch:      {"checked"},
	accessibility.RoleMenuItemCheckbox: {"checked"},
	accessibility.RoleMenuItemRadio:    {"checked"},
	accessibility.RoleHeading:          {"level"},
}

// RequiredPropertiesForRole returns what a role needs to be valid.
func RequiredPropertiesForRole(role accessibility.Role) []string {
	props, ok := roleRequiredProperties[role]
	if !ok {
		return nil
	}
	out := make([]string, len(props))
	copy(out, props)
	return out
}

// invalidStateForRole lists states that should never be set on certain roles.
var invalidStateForRole = map[accessibility.Role][]string{
	accessibility.RoleButton:    {"multiline"},
	accessibility.RoleLink:      {"multiline"},
	accessibility.RoleImg:       {"multiline"},
	accessibility.RoleImage:     {"multiline"},
	accessibility.RoleSeparator: {"multiline", "multiselectable"},
	accessibility.RoleAlert:     {"multiline"},
	accessibility.RoleStatus:    {"multiline"},
	accessibility.RoleTimer:     {"multiline"},
	accessibility.RoleTooltip:   {"multiline"},
}

// containerChildRoles maps container roles to their expected child roles.
var containerChildRoles = map[accessibility.Role][]accessibility.Role{
	accessibility.RoleList:       {accessibility.RoleListItem, accessibility.RoleOption, accessibility.RoleGroup},
	accessibility.RoleListbox:    {accessibility.RoleOption, accessibility.RoleGroup},
	accessibility.RoleMenu:       {accessibility.RoleMenuItem, accessibility.RoleMenuItemCheckbox, accessibility.RoleMenuItemRadio, accessibility.RoleGroup, accessibility.RoleSeparator},
	accessibility.RoleMenuBar:    {accessibility.RoleMenuItem, accessibility.RoleMenuItemCheckbox, accessibility.RoleMenuItemRadio, accessibility.RoleGroup, accessibility.RoleSeparator},
	accessibility.RoleTabList:    {accessibility.RoleTab},
	accessibility.RoleTree:       {accessibility.RoleTreeItem, accessibility.RoleGroup},
	accessibility.RoleRadioGroup: {accessibility.RoleRadio},
	accessibility.RoleGrid:       {accessibility.RoleRow, accessibility.RoleRowGroup},
	accessibility.RoleTreeGrid:   {accessibility.RoleRow, accessibility.RoleRowGroup},
	accessibility.RoleRow:        {accessibility.RoleCell, accessibility.RoleGridCell, accessibility.RoleColumnHeader, accessibility.RoleRowHeader},
	accessibility.RoleRowGroup:   {accessibility.RoleRow},
	accessibility.RoleTable:      {accessibility.RoleRow, accessibility.RoleRowGroup},
}

// rolesWithExpandedState lists roles where aria-expanded is meaningful.
var rolesWithExpandedState = map[accessibility.Role]bool{
	accessibility.RoleTree:       true,
	accessibility.RoleTreeItem:   true,
	accessibility.RoleCombobox:   true,
	accessibility.RoleButton:     true,
	accessibility.RoleListbox:    true,
	accessibility.RoleMenuItem:   true,
	accessibility.RoleRow:        true,
	accessibility.RoleTab:        true,
	accessibility.RoleGroup:      true,
	accessibility.RoleGrid:       true,
	accessibility.RoleTreeGrid:   true,
	accessibility.RoleMenuItemCheckbox: true,
	accessibility.RoleMenuItemRadio:    true,
}

// rolesWithCheckedState lists roles where aria-checked is meaningful.
var rolesWithCheckedState = map[accessibility.Role]bool{
	accessibility.RoleCheckbox:         true,
	accessibility.RoleRadio:            true,
	accessibility.RoleSwitch:           true,
	accessibility.RoleMenuItemCheckbox: true,
	accessibility.RoleMenuItemRadio:    true,
	accessibility.RoleOption:           true,
	accessibility.RoleTreeItem:         true,
}

// rolesWithSelectedState lists roles where aria-selected is meaningful.
var rolesWithSelectedState = map[accessibility.Role]bool{
	accessibility.RoleTab:          true,
	accessibility.RoleOption:       true,
	accessibility.RoleGridCell:     true,
	accessibility.RoleRow:          true,
	accessibility.RoleColumnHeader: true,
	accessibility.RoleRowHeader:    true,
	accessibility.RoleCell:         true,
	accessibility.RoleTreeItem:     true,
}

// inputRoles lists roles where aria-required and aria-invalid are meaningful.
var inputRoles = map[accessibility.Role]bool{
	accessibility.RoleTextbox:    true,
	accessibility.RoleSearchbox:  true,
	accessibility.RoleCombobox:   true,
	accessibility.RoleSpinButton: true,
	accessibility.RoleSlider:     true,
	accessibility.RoleCheckbox:   true,
	accessibility.RoleRadio:      true,
	accessibility.RoleSwitch:     true,
	accessibility.RoleListbox:    true,
	accessibility.RoleGrid:       true,
	accessibility.RoleTreeGrid:   true,
}

// interactiveRoles lists roles that are inherently interactive and should be focusable.
var interactiveRoles = map[accessibility.Role]bool{
	accessibility.RoleButton:           true,
	accessibility.RoleCheckbox:         true,
	accessibility.RoleRadio:            true,
	accessibility.RoleTextbox:          true,
	accessibility.RoleSearchbox:        true,
	accessibility.RoleCombobox:         true,
	accessibility.RoleSlider:           true,
	accessibility.RoleSpinButton:       true,
	accessibility.RoleSwitch:           true,
	accessibility.RoleLink:             true,
	accessibility.RoleMenuItem:         true,
	accessibility.RoleMenuItemCheckbox: true,
	accessibility.RoleMenuItemRadio:    true,
	accessibility.RoleTab:              true,
	accessibility.RoleOption:           true,
	accessibility.RoleTreeItem:         true,
	accessibility.RoleScrollbar:        true,
	accessibility.RoleGridCell:         true,
}

// nonInteractiveRoles lists roles that should typically not be focusable.
var nonInteractiveRoles = map[accessibility.Role]bool{
	accessibility.RoleText:         true,
	accessibility.RoleSeparator:    true,
	accessibility.RoleImg:          true,
	accessibility.RoleImage:        true,
	accessibility.RolePresentation: true,
	accessibility.RoleNone:         true,
	accessibility.RoleNote:         true,
	accessibility.RoleMark:         true,
	accessibility.RoleCode:         true,
	accessibility.RoleTime:         true,
	accessibility.RoleMarquee:      true,
}

// AuditARIARoles validates ARIA role requirements, invalid states, and container children.
func AuditARIARoles(root runtime.Widget) []A11yIssue {
	if root == nil {
		return nil
	}
	var issues []A11yIssue
	landmarkMainCount := 0

	walkTree(root, nil, func(node runtime.Widget, path string) {
		acc, ok := node.(accessibility.Accessible)
		if !ok {
			return
		}
		role := acc.AccessibleRole()
		if role == "" {
			return
		}
		state := acc.AccessibleState()

		// Check required properties per role.
		if reqs, ok := roleRequiredProperties[role]; ok {
			for _, req := range reqs {
				switch req {
				case "value":
					if acc.AccessibleValue() == nil {
						issues = append(issues, A11yIssue{
							Path:     path,
							Message:  fmt.Sprintf("role %q requires a value (AccessibleValue must be non-nil)", role),
							Severity: A11yError,
						})
					}
				case "selected":
					if !state.Selected && state.Checked == nil {
						issues = append(issues, A11yIssue{
							Path:     path,
							Message:  fmt.Sprintf("role %q requires a selected or checked state", role),
							Severity: A11yWarning,
						})
					}
				case "checked":
					if state.Checked == nil {
						issues = append(issues, A11yIssue{
							Path:     path,
							Message:  fmt.Sprintf("role %q requires a checked state (Checked must be non-nil)", role),
							Severity: A11yError,
						})
					}
				case "level":
					if acc.AccessibleLevel() <= 0 {
						issues = append(issues, A11yIssue{
							Path:     path,
							Message:  fmt.Sprintf("role %q requires a positive level", role),
							Severity: A11yWarning,
						})
					}
				}
			}
		}

		// Check invalid state combinations.
		if badStates, ok := invalidStateForRole[role]; ok {
			for _, bad := range badStates {
				switch bad {
				case "multiline":
					if state.Multiline {
						issues = append(issues, A11yIssue{
							Path:     path,
							Message:  fmt.Sprintf("role %q should not have multiline state", role),
							Severity: A11yWarning,
						})
					}
				case "multiselectable":
					if state.Multiselectable {
						issues = append(issues, A11yIssue{
							Path:     path,
							Message:  fmt.Sprintf("role %q should not have multiselectable state", role),
							Severity: A11yWarning,
						})
					}
				}
			}
		}

		// Check container-child role relationships.
		if expectedChildren, ok := containerChildRoles[role]; ok {
			if cp, ok := node.(runtime.ChildProvider); ok {
				children := cp.ChildWidgets()
				for _, child := range children {
					if child == nil {
						continue
					}
					childAcc, ok := child.(accessibility.Accessible)
					if !ok {
						continue
					}
					childRole := childAcc.AccessibleRole()
					if childRole == "" {
						continue
					}
					found := false
					for _, expected := range expectedChildren {
						if childRole == expected {
							found = true
							break
						}
					}
					if !found {
						expectedStr := make([]string, len(expectedChildren))
						for i, e := range expectedChildren {
							expectedStr[i] = string(e)
						}
						issues = append(issues, A11yIssue{
							Path:     path,
							Message:  fmt.Sprintf("container role %q has child with role %q; expected one of [%s]", role, childRole, strings.Join(expectedStr, ", ")),
							Severity: A11yWarning,
						})
					}
				}
			}
		}

		// Landmark uniqueness: only one main landmark per tree.
		if acc.AccessibleLandmark() == accessibility.LandmarkMain {
			landmarkMainCount++
		}
	})

	if landmarkMainCount > 1 {
		issues = append(issues, A11yIssue{
			Path:     "(tree)",
			Message:  fmt.Sprintf("found %d main landmarks; there should be at most 1", landmarkMainCount),
			Severity: A11yError,
		})
	}

	return issues
}

// ---------------------------------------------------------------------------
// ARIA state consistency validation
// ---------------------------------------------------------------------------

// AuditARIAStates validates that ARIA states are used on appropriate roles.
func AuditARIAStates(root runtime.Widget) []A11yIssue {
	if root == nil {
		return nil
	}
	var issues []A11yIssue

	walkTree(root, nil, func(node runtime.Widget, path string) {
		acc, ok := node.(accessibility.Accessible)
		if !ok {
			return
		}
		role := acc.AccessibleRole()
		if role == "" {
			return
		}
		state := acc.AccessibleState()

		// aria-expanded only on expandable roles.
		if state.Expanded != nil && !rolesWithExpandedState[role] {
			issues = append(issues, A11yIssue{
				Path:     path,
				Message:  fmt.Sprintf("aria-expanded set on role %q which does not support it", role),
				Severity: A11yWarning,
			})
		}

		// aria-checked only on checkable roles.
		if state.Checked != nil && !rolesWithCheckedState[role] {
			issues = append(issues, A11yIssue{
				Path:     path,
				Message:  fmt.Sprintf("aria-checked set on role %q which does not support it", role),
				Severity: A11yWarning,
			})
		}

		// aria-selected only on selectable roles.
		if state.Selected && !rolesWithSelectedState[role] {
			issues = append(issues, A11yIssue{
				Path:     path,
				Message:  fmt.Sprintf("aria-selected set on role %q which does not support it", role),
				Severity: A11yWarning,
			})
		}

		// aria-required / aria-invalid only on input roles.
		if state.Required && !inputRoles[role] {
			issues = append(issues, A11yIssue{
				Path:     path,
				Message:  fmt.Sprintf("aria-required set on role %q which is not an input role", role),
				Severity: A11yWarning,
			})
		}
		if state.Invalid && !inputRoles[role] {
			issues = append(issues, A11yIssue{
				Path:     path,
				Message:  fmt.Sprintf("aria-invalid set on role %q which is not an input role", role),
				Severity: A11yWarning,
			})
		}

		// Value range validation: if ValueInfo has Min/Max, Current should be in range.
		if v := acc.AccessibleValue(); v != nil {
			if v.Max > v.Min && (v.Current < v.Min || v.Current > v.Max) {
				issues = append(issues, A11yIssue{
					Path:     path,
					Message:  fmt.Sprintf("value current %.2f out of range [%.2f, %.2f]", v.Current, v.Min, v.Max),
					Severity: A11yError,
				})
			}
		}
	})

	return issues
}

// ---------------------------------------------------------------------------
// Keyboard navigation audit
// ---------------------------------------------------------------------------

// AuditKeyboardNav validates keyboard navigation accessibility.
func AuditKeyboardNav(root runtime.Widget) []A11yIssue {
	if root == nil {
		return nil
	}
	var issues []A11yIssue

	// Collect all focusable widgets and interactive widgets.
	type widgetInfo struct {
		path   string
		widget runtime.Widget
		role   accessibility.Role
		bounds runtime.Rect
	}

	var focusableWidgets []widgetInfo
	var interactiveNotFocusable []widgetInfo
	var nonInteractiveFocusable []widgetInfo

	walkTree(root, nil, func(node runtime.Widget, path string) {
		var role accessibility.Role
		if acc, ok := node.(accessibility.Accessible); ok {
			role = acc.AccessibleRole()
		}

		focusable, isFocusable := node.(runtime.Focusable)
		canFocus := isFocusable && focusable.CanFocus()

		var bounds runtime.Rect
		if bp, ok := node.(runtime.BoundsProvider); ok {
			bounds = bp.Bounds()
		}

		info := widgetInfo{path: path, widget: node, role: role, bounds: bounds}

		if canFocus {
			focusableWidgets = append(focusableWidgets, info)
		}

		// Interactive roles must be focusable.
		if interactiveRoles[role] && !canFocus {
			// Only flag if the widget's state is not disabled and not hidden.
			if acc, ok := node.(accessibility.Accessible); ok {
				state := acc.AccessibleState()
				if !state.Disabled && !state.Hidden {
					interactiveNotFocusable = append(interactiveNotFocusable, info)
				}
			}
		}

		// Non-interactive roles should not be focusable.
		if nonInteractiveRoles[role] && canFocus {
			nonInteractiveFocusable = append(nonInteractiveFocusable, info)
		}
	})

	for _, w := range interactiveNotFocusable {
		issues = append(issues, A11yIssue{
			Path:     w.path,
			Message:  fmt.Sprintf("interactive role %q should be focusable", w.role),
			Severity: A11yError,
		})
	}

	for _, w := range nonInteractiveFocusable {
		issues = append(issues, A11yIssue{
			Path:     w.path,
			Message:  fmt.Sprintf("non-interactive role %q should not be focusable", w.role),
			Severity: A11yWarning,
		})
	}

	// Validate focus order is logical (top-to-bottom, left-to-right).
	for i := 1; i < len(focusableWidgets); i++ {
		prev := focusableWidgets[i-1]
		curr := focusableWidgets[i]
		// Check if current widget appears before previous in visual order.
		// Visual order: top-to-bottom (Y), then left-to-right (X).
		if curr.bounds.Y < prev.bounds.Y ||
			(curr.bounds.Y == prev.bounds.Y && curr.bounds.X < prev.bounds.X) {
			// Only flag if both have non-zero bounds (already laid out).
			if (prev.bounds.Width > 0 || prev.bounds.Height > 0) &&
				(curr.bounds.Width > 0 || curr.bounds.Height > 0) {
				issues = append(issues, A11yIssue{
					Path:     curr.path,
					Message:  fmt.Sprintf("focus order may be illogical: widget at (%d,%d) follows widget at (%d,%d)", curr.bounds.X, curr.bounds.Y, prev.bounds.X, prev.bounds.Y),
					Severity: A11yWarning,
				})
			}
		}
	}

	return issues
}

// ---------------------------------------------------------------------------
// Relationship validation
// ---------------------------------------------------------------------------

// AuditRelationships validates that ARIA relationship references exist in the tree.
func AuditRelationships(root runtime.Widget) []A11yIssue {
	if root == nil {
		return nil
	}
	var issues []A11yIssue

	// First pass: collect all widget IDs.
	idSet := map[string]bool{}
	type relRef struct {
		path     string
		attr     string
		targetID string
	}
	var refs []relRef

	// Also track per-widget references for circular detection.
	type widgetRefs struct {
		id          string
		labelledBy  string
		describedBy string
		controls    string
		owns        string
		flowTo      string
	}
	refsByID := map[string]widgetRefs{}

	walkTree(root, nil, func(node runtime.Widget, path string) {
		acc, ok := node.(accessibility.Accessible)
		if !ok {
			return
		}

		// Collect ID from widgets that have one.
		if idProvider, ok := node.(interface{ ID() string }); ok {
			id := strings.TrimSpace(idProvider.ID())
			if id != "" {
				idSet[id] = true
			}
		}

		labelledBy := strings.TrimSpace(acc.AccessibleLabelledBy())
		describedBy := strings.TrimSpace(acc.AccessibleDescribedBy())
		controls := strings.TrimSpace(acc.AccessibleControls())
		owns := strings.TrimSpace(acc.AccessibleOwns())
		flowTo := strings.TrimSpace(acc.AccessibleFlowTo())

		if labelledBy != "" {
			refs = append(refs, relRef{path: path, attr: "aria-labelledby", targetID: labelledBy})
		}
		if describedBy != "" {
			refs = append(refs, relRef{path: path, attr: "aria-describedby", targetID: describedBy})
		}
		if controls != "" {
			refs = append(refs, relRef{path: path, attr: "aria-controls", targetID: controls})
		}
		if owns != "" {
			refs = append(refs, relRef{path: path, attr: "aria-owns", targetID: owns})
		}
		if flowTo != "" {
			refs = append(refs, relRef{path: path, attr: "aria-flowto", targetID: flowTo})
		}

		// Store references by widget ID for circular detection.
		if idProvider, ok := node.(interface{ ID() string }); ok {
			id := strings.TrimSpace(idProvider.ID())
			if id != "" {
				refsByID[id] = widgetRefs{
					id:          id,
					labelledBy:  labelledBy,
					describedBy: describedBy,
					controls:    controls,
					owns:        owns,
					flowTo:      flowTo,
				}
			}
		}
	})

	// Validate that referenced IDs exist.
	for _, ref := range refs {
		if !idSet[ref.targetID] {
			issues = append(issues, A11yIssue{
				Path:     ref.path,
				Message:  fmt.Sprintf("%s references ID %q which does not exist in the tree", ref.attr, ref.targetID),
				Severity: A11yError,
			})
		}
	}

	// Check for circular references (A labelledBy B and B labelledBy A, etc).
	for id, wr := range refsByID {
		for _, targetID := range []string{wr.labelledBy, wr.describedBy, wr.controls, wr.owns, wr.flowTo} {
			if targetID == "" || targetID == id {
				if targetID == id {
					issues = append(issues, A11yIssue{
						Path:     "(id=" + id + ")",
						Message:  fmt.Sprintf("widget references itself via ARIA relationship (target ID %q)", targetID),
						Severity: A11yError,
					})
				}
				continue
			}
			// Check if target references back to this widget via any relationship.
			if targetRefs, ok := refsByID[targetID]; ok {
				for _, backRef := range []string{targetRefs.labelledBy, targetRefs.describedBy, targetRefs.controls, targetRefs.owns, targetRefs.flowTo} {
					if backRef == id {
						issues = append(issues, A11yIssue{
							Path:     "(id=" + id + ")",
							Message:  fmt.Sprintf("circular ARIA reference: %q and %q reference each other", id, targetID),
							Severity: A11yWarning,
						})
					}
				}
			}
		}
	}

	return issues
}

// ---------------------------------------------------------------------------
// Live region validation
// ---------------------------------------------------------------------------

// AuditLiveRegions validates live region ARIA attributes.
func AuditLiveRegions(root runtime.Widget) []A11yIssue {
	if root == nil {
		return nil
	}
	var issues []A11yIssue
	assertiveCount := 0

	walkTree(root, nil, func(node runtime.Widget, path string) {
		acc, ok := node.(accessibility.Accessible)
		if !ok {
			return
		}

		live := acc.AccessibleLive()
		relevant := acc.AccessibleRelevant()
		atomic := acc.AccessibleAtomic()

		if live == accessibility.LiveAssertive {
			assertiveCount++
		}

		// Relevant is only meaningful when Live != off.
		if relevant != "" && live == accessibility.LiveOff {
			issues = append(issues, A11yIssue{
				Path:     path,
				Message:  "aria-relevant is set but aria-live is off; relevant has no effect",
				Severity: A11yWarning,
			})
		}

		// Atomic is only meaningful when Live != off.
		if atomic && live == accessibility.LiveOff {
			issues = append(issues, A11yIssue{
				Path:     path,
				Message:  "aria-atomic is set but aria-live is off; atomic has no effect",
				Severity: A11yWarning,
			})
		}
	})

	if assertiveCount > 2 {
		issues = append(issues, A11yIssue{
			Path:     "(tree)",
			Message:  fmt.Sprintf("found %d assertive live regions; use sparingly (recommended max 2)", assertiveCount),
			Severity: A11yWarning,
		})
	}

	return issues
}

// ---------------------------------------------------------------------------
// Comprehensive audit: combines all checks
// ---------------------------------------------------------------------------

// AuditARIAAll runs all ARIA validation checks and returns combined issues.
func AuditARIAAll(root runtime.Widget) []A11yIssue {
	var all []A11yIssue
	all = append(all, AuditAccessibility(root)...)
	all = append(all, AuditARIARoles(root)...)
	all = append(all, AuditARIAStates(root)...)
	all = append(all, AuditKeyboardNav(root)...)
	all = append(all, AuditRelationships(root)...)
	all = append(all, AuditLiveRegions(root)...)
	return all
}

// AssertARIACompliant runs all ARIA checks and fails on errors.
func AssertARIACompliant(t *testing.T, root runtime.Widget) {
	t.Helper()
	issues := AuditARIAAll(root)
	for _, issue := range issues {
		if issue.Severity == A11yWarning {
			t.Logf("a11y warning: %s: %s", issue.Path, issue.Message)
			continue
		}
		t.Errorf("a11y error: %s: %s", issue.Path, issue.Message)
	}
}

// ValidateWidget validates a single widget's ARIA compliance.
func ValidateWidget(t *testing.T, w runtime.Widget) {
	t.Helper()
	if w == nil {
		return
	}
	acc, ok := w.(accessibility.Accessible)
	if !ok {
		t.Logf("a11y warning: widget %s does not implement Accessible", widgetName(w))
		return
	}

	path := widgetName(w)
	role := acc.AccessibleRole()
	state := acc.AccessibleState()
	label := strings.TrimSpace(acc.AccessibleLabel())

	if role == "" {
		t.Errorf("a11y error: %s: missing accessible role", path)
	}
	if label == "" {
		severity := "warning"
		if focusable, ok := w.(runtime.Focusable); ok && focusable.CanFocus() {
			severity = "error"
		}
		if severity == "error" {
			t.Errorf("a11y error: %s: missing accessible label", path)
		} else {
			t.Logf("a11y warning: %s: missing accessible label", path)
		}
	}

	// Required properties.
	if reqs, ok := roleRequiredProperties[role]; ok {
		for _, req := range reqs {
			switch req {
			case "value":
				if acc.AccessibleValue() == nil {
					t.Errorf("a11y error: %s: role %q requires a value", path, role)
				}
			case "checked":
				if state.Checked == nil {
					t.Errorf("a11y error: %s: role %q requires a checked state", path, role)
				}
			case "selected":
				if !state.Selected && state.Checked == nil {
					t.Logf("a11y warning: %s: role %q should have selected or checked state", path, role)
				}
			case "level":
				if acc.AccessibleLevel() <= 0 {
					t.Logf("a11y warning: %s: role %q should have a positive level", path, role)
				}
			}
		}
	}

	// Invalid states.
	if badStates, ok := invalidStateForRole[role]; ok {
		for _, bad := range badStates {
			switch bad {
			case "multiline":
				if state.Multiline {
					t.Logf("a11y warning: %s: role %q should not have multiline state", path, role)
				}
			case "multiselectable":
				if state.Multiselectable {
					t.Logf("a11y warning: %s: role %q should not have multiselectable state", path, role)
				}
			}
		}
	}

	// State consistency.
	if state.Expanded != nil && !rolesWithExpandedState[role] {
		t.Logf("a11y warning: %s: aria-expanded set on role %q which does not support it", path, role)
	}
	if state.Checked != nil && !rolesWithCheckedState[role] {
		t.Logf("a11y warning: %s: aria-checked set on role %q which does not support it", path, role)
	}
	if state.Selected && !rolesWithSelectedState[role] {
		t.Logf("a11y warning: %s: aria-selected set on role %q which does not support it", path, role)
	}
	if state.Required && !inputRoles[role] {
		t.Logf("a11y warning: %s: aria-required set on role %q which is not an input role", path, role)
	}
	if state.Invalid && !inputRoles[role] {
		t.Logf("a11y warning: %s: aria-invalid set on role %q which is not an input role", path, role)
	}

	// Value range.
	if v := acc.AccessibleValue(); v != nil {
		if v.Max > v.Min && (v.Current < v.Min || v.Current > v.Max) {
			t.Errorf("a11y error: %s: value current %.2f out of range [%.2f, %.2f]", path, v.Current, v.Min, v.Max)
		}
	}

	// Live region consistency.
	live := acc.AccessibleLive()
	if acc.AccessibleRelevant() != "" && live == accessibility.LiveOff {
		t.Logf("a11y warning: %s: aria-relevant set but aria-live is off", path)
	}
	if acc.AccessibleAtomic() && live == accessibility.LiveOff {
		t.Logf("a11y warning: %s: aria-atomic set but aria-live is off", path)
	}

	// Interactive role focusability.
	if interactiveRoles[role] {
		if focusable, ok := w.(runtime.Focusable); !ok || !focusable.CanFocus() {
			if !state.Disabled && !state.Hidden {
				t.Errorf("a11y error: %s: interactive role %q should be focusable", path, role)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// walkTree traverses the widget tree depth-first, calling fn for each node.
func walkTree(root runtime.Widget, path []string, fn func(node runtime.Widget, path string)) {
	if root == nil {
		return
	}
	var walk func(node runtime.Widget, path []string)
	walk = func(node runtime.Widget, path []string) {
		if node == nil {
			return
		}
		name := widgetName(node)
		currentPath := append(path, name)
		pathStr := strings.Join(currentPath, "/")

		fn(node, pathStr)

		provider, ok := node.(runtime.ChildProvider)
		if !ok {
			return
		}
		for _, child := range provider.ChildWidgets() {
			if child == nil {
				continue
			}
			segment := widgetName(child)
			if segmenter, ok := node.(runtime.PathSegmenter); ok {
				if seg := strings.TrimSpace(segmenter.PathSegment(child)); seg != "" {
					segment = seg
				}
			}
			walk(child, append(path, segment))
		}
	}

	walk(root, path)
}
