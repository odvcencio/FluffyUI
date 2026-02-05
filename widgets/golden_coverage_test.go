package widgets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoldenCoverage_CoreWidgetCatalog(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".golden" {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".golden"))
	}

	required := map[string][]string{
		"layout.grid":                 {"grid_"},
		"layout.stack":                {"stack_"},
		"layout.splitter":             {"splitter_"},
		"layout.scrollview":           {"scrollview_"},
		"layout.panel":                {"panel_"},
		"layout.aspect_ratio":         {"aspect_ratio_"},
		"data.list":                   {"list_"},
		"data.table":                  {"table_"},
		"data.datagrid":               {"datagrid_", "data_grid"},
		"data.tree":                   {"tree_"},
		"data.directory_tree":         {"directory_tree_"},
		"data.log":                    {"log"},
		"data.rich_text":              {"rich_text"},
		"data.search_widget":          {"search_widget_"},
		"input.button":                {"button_"},
		"input.input":                 {"input_"},
		"input.masked_input":          {"masked_input"},
		"input.textarea":              {"textarea_"},
		"input.checkbox":              {"checkbox_"},
		"input.radio":                 {"radio_"},
		"input.select":                {"select_"},
		"input.autocomplete":          {"autocomplete"},
		"input.multiselect":           {"multiselect"},
		"input.slider":                {"slider_"},
		"input.range_slider":          {"range_slider_"},
		"input.datepicker":            {"datepicker_"},
		"input.daterangepicker":       {"daterangepicker_", "date_range_picker"},
		"input.timepicker":            {"timepicker_", "time_picker"},
		"navigation.tabs":             {"tabs_"},
		"navigation.menu":             {"menu_"},
		"navigation.breadcrumb":       {"breadcrumb_"},
		"navigation.stepper":          {"stepper_"},
		"navigation.palette":          {"palette"},
		"navigation.enhanced_palette": {"enhanced_palette_"},
		"navigation.accordion":        {"accordion_"},
		"navigation.section":          {"section_"},
		"feedback.dialog":             {"dialog_"},
		"feedback.spinner":            {"spinner_"},
		"feedback.progress":           {"progress_"},
		"feedback.alert":              {"alert_"},
		"feedback.toast_stack":        {"toast_stack_"},
		"feedback.sparkline":          {"sparkline_"},
		"feedback.barchart":           {"barchart_"},
		"feedback.linechart":          {"linechart_"},
		"helpers.label":               {"label_"},
		"helpers.text":                {"text_"},
		"helpers.simple":              {"simple_widget_"},
		"helpers.debug_overlay":       {"debug_overlay_"},
	}

	for key, prefixes := range required {
		if !hasAnyPrefix(names, prefixes) {
			t.Fatalf("missing golden coverage for %s (prefixes: %v)", key, prefixes)
		}
	}
}

func hasAnyPrefix(names []string, prefixes []string) bool {
	for _, name := range names {
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				return true
			}
		}
	}
	return false
}
