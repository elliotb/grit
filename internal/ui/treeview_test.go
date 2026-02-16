package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/ejb/grit/internal/gt"
)

func TestRenderTree_Empty(t *testing.T) {
	result := renderTree(nil)
	if result != "(no stacks)" {
		t.Errorf("got %q, want %q", result, "(no stacks)")
	}
}

func TestRenderTree_SingleRoot(t *testing.T) {
	branches := []*gt.Branch{
		{Name: "main", IsCurrent: true},
	}
	result := ansi.Strip(renderTree(branches))
	if !strings.Contains(result, "main") {
		t.Errorf("output should contain 'main', got:\n%s", result)
	}
	if !strings.Contains(result, "◉") {
		t.Errorf("output should contain ◉ marker, got:\n%s", result)
	}
}

func TestRenderTree_LinearStack(t *testing.T) {
	branches := []*gt.Branch{
		{
			Name: "main",
			Children: []*gt.Branch{
				{
					Name: "feature-a",
					Children: []*gt.Branch{
						{
							Name:      "feature-b",
							IsCurrent: true,
						},
					},
				},
			},
		},
	}

	result := ansi.Strip(renderTree(branches))

	for _, name := range []string{"main", "feature-a", "feature-b"} {
		if !strings.Contains(result, name) {
			t.Errorf("output should contain %q, got:\n%s", name, result)
		}
	}

	// Should have tree connectors
	if !strings.Contains(result, "└──") && !strings.Contains(result, "├──") {
		t.Errorf("output should contain tree connectors, got:\n%s", result)
	}
}

func TestRenderTree_CurrentBranchHighlighted(t *testing.T) {
	branches := []*gt.Branch{
		{
			Name: "main",
			Children: []*gt.Branch{
				{Name: "feature-a"},
				{Name: "feature-b", IsCurrent: true},
			},
		},
	}

	result := ansi.Strip(renderTree(branches))

	// Current branch should have ◉ prefix
	if !strings.Contains(result, "◉ feature-b") {
		t.Errorf("current branch should have ◉ prefix, got:\n%s", result)
	}

	// Non-current branches should have ◯ prefix
	if !strings.Contains(result, "◯ feature-a") {
		t.Errorf("non-current branch should have ◯ prefix, got:\n%s", result)
	}
	if !strings.Contains(result, "◯ main") {
		t.Errorf("root should have ◯ prefix, got:\n%s", result)
	}
}

func TestRenderTree_MultipleStacks(t *testing.T) {
	branches := []*gt.Branch{
		{
			Name: "main",
			Children: []*gt.Branch{
				{
					Name: "stack-1-base",
					Children: []*gt.Branch{
						{Name: "stack-1-top", IsCurrent: true},
					},
				},
				{Name: "standalone-branch"},
			},
		},
	}

	result := ansi.Strip(renderTree(branches))

	for _, name := range []string{"main", "stack-1-base", "stack-1-top", "standalone-branch"} {
		if !strings.Contains(result, name) {
			t.Errorf("output should contain %q, got:\n%s", name, result)
		}
	}
}

func TestRenderTree_TreeStructure(t *testing.T) {
	// Verify the tree structure renders correctly with proper nesting
	branches := []*gt.Branch{
		{
			Name: "main",
			Children: []*gt.Branch{
				{
					Name: "feature-a",
					Children: []*gt.Branch{
						{Name: "feature-a1"},
						{Name: "feature-a2"},
					},
				},
				{Name: "feature-b"},
			},
		},
	}

	result := ansi.Strip(renderTree(branches))

	// The tree should show feature-a1 and feature-a2 as children of feature-a
	// and feature-b as a sibling of feature-a (both children of main)
	lines := strings.Split(result, "\n")
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines of output, got %d:\n%s", len(lines), result)
	}
}
