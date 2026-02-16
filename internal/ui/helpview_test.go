package ui

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderHelp_ContainsSections(t *testing.T) {
	result := ansi.Strip(renderHelp())

	sections := []string{"Navigation", "Actions", "Views", "Diff View"}
	for _, section := range sections {
		if !containsString(result, section) {
			t.Errorf("help should contain section %q", section)
		}
	}
}

func TestRenderHelp_ContainsKeys(t *testing.T) {
	result := ansi.Strip(renderHelp())

	keys := []string{
		"enter", "Check out selected branch",
		"s", "Submit stack",
		"r", "Restack",
		"d", "Open diff",
		"tab", "Switch panel",
		"q", "Quit",
	}
	for _, k := range keys {
		if !containsString(result, k) {
			t.Errorf("help should contain %q", k)
		}
	}
}

func TestRenderHelp_ContainsCloseInstruction(t *testing.T) {
	result := ansi.Strip(renderHelp())

	if !containsString(result, "Press ? or esc to close") {
		t.Error("help should contain close instruction")
	}
}
