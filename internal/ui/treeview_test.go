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
	if !strings.Contains(result, "◉ main") {
		t.Errorf("output should contain '◉ main', got:\n%s", result)
	}
}

func TestRenderTree_LinearStack(t *testing.T) {
	// Parser produces a chain: main → a → b → c
	// Renderer should flatten to all at same depth with │ prefix
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
	lines := strings.Split(result, "\n")

	// Should have 3 lines: main, feature-a, feature-b
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), result)
	}

	// main at depth 0 (no │ prefix)
	if strings.HasPrefix(lines[0], "│") {
		t.Errorf("root should not have │ prefix, got: %q", lines[0])
	}
	if !strings.Contains(lines[0], "main") {
		t.Errorf("first line should contain 'main', got: %q", lines[0])
	}

	// feature-a and feature-b should both have │ prefix at same indent
	for _, i := range []int{1, 2} {
		if !strings.HasPrefix(lines[i], "│ ") {
			t.Errorf("line %d should start with '│ ', got: %q", i, lines[i])
		}
	}

	// Both stack branches at same indent (same number of │ prefixes)
	if !strings.Contains(lines[1], "feature-a") {
		t.Errorf("line 1 should contain 'feature-a', got: %q", lines[1])
	}
	if !strings.Contains(lines[2], "feature-b") {
		t.Errorf("line 2 should contain 'feature-b', got: %q", lines[2])
	}
}

func TestRenderTree_CurrentBranchMarker(t *testing.T) {
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

	if !strings.Contains(result, "◉ feature-b") {
		t.Errorf("current branch should have ◉ prefix, got:\n%s", result)
	}
	if !strings.Contains(result, "◯ feature-a") {
		t.Errorf("non-current branch should have ◯ prefix, got:\n%s", result)
	}
	if !strings.Contains(result, "◯ main") {
		t.Errorf("root should have ◯ prefix, got:\n%s", result)
	}
}

func TestRenderTree_StandaloneBranch(t *testing.T) {
	// A standalone branch (leaf child at root level) should be at depth 0
	branches := []*gt.Branch{
		{
			Name: "main",
			Children: []*gt.Branch{
				{Name: "standalone-fix"},
			},
		},
	}

	result := ansi.Strip(renderTree(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), result)
	}

	// standalone-fix should be at depth 0 (no │ prefix)
	if strings.HasPrefix(lines[1], "│") {
		t.Errorf("standalone branch should not have │ prefix, got: %q", lines[1])
	}
	if !strings.Contains(lines[1], "standalone-fix") {
		t.Errorf("line 1 should contain 'standalone-fix', got: %q", lines[1])
	}
}

func TestRenderTree_MultipleStacks(t *testing.T) {
	// main has a stack (a→b→c) and a standalone branch
	branches := []*gt.Branch{
		{
			Name: "main",
			Children: []*gt.Branch{
				{
					Name: "stack-base",
					Children: []*gt.Branch{
						{Name: "stack-top", IsCurrent: true},
					},
				},
				{Name: "standalone"},
			},
		},
	}

	result := ansi.Strip(renderTree(branches))
	lines := strings.Split(result, "\n")

	// main (depth 0), stack-base (depth 1), stack-top (depth 1), standalone (depth 0)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d:\n%s", len(lines), result)
	}

	// Stack branches have │ prefix
	if !strings.HasPrefix(lines[1], "│ ") {
		t.Errorf("stack-base should have │ prefix, got: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "│ ") {
		t.Errorf("stack-top should have │ prefix, got: %q", lines[2])
	}

	// Standalone at depth 0
	if strings.HasPrefix(lines[3], "│") {
		t.Errorf("standalone should not have │ prefix, got: %q", lines[3])
	}
}

func TestRenderTree_DeepChainFlattened(t *testing.T) {
	// A long chain: main → a → b → c → d → e
	// All should render at the same depth (│ prefix)
	e := &gt.Branch{Name: "e", IsCurrent: true}
	d := &gt.Branch{Name: "d", Children: []*gt.Branch{e}}
	c := &gt.Branch{Name: "c", Children: []*gt.Branch{d}}
	b := &gt.Branch{Name: "b", Children: []*gt.Branch{c}}
	a := &gt.Branch{Name: "a", Children: []*gt.Branch{b}}
	branches := []*gt.Branch{
		{Name: "main", Children: []*gt.Branch{a}},
	}

	result := ansi.Strip(renderTree(branches))
	lines := strings.Split(result, "\n")

	// 6 lines: main + 5 stack branches
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d:\n%s", len(lines), result)
	}

	// main at depth 0
	if strings.HasPrefix(lines[0], "│") {
		t.Errorf("main should not have │ prefix")
	}

	// All stack branches at same depth (single │ prefix)
	for i := 1; i <= 5; i++ {
		if !strings.HasPrefix(lines[i], "│ ") {
			t.Errorf("line %d should start with '│ ', got: %q", i, lines[i])
		}
		// Should NOT have double │ (no nesting)
		if strings.HasPrefix(lines[i], "│ │") {
			t.Errorf("line %d should not be double-nested, got: %q", i, lines[i])
		}
	}
}

func TestFlattenForDisplay_Entries(t *testing.T) {
	branches := []*gt.Branch{
		{
			Name: "main",
			Children: []*gt.Branch{
				{
					Name: "a",
					Children: []*gt.Branch{
						{Name: "b"},
					},
				},
				{Name: "standalone"},
			},
		},
	}

	entries := flattenForDisplay(branches)

	expected := []struct {
		name  string
		depth int
	}{
		{"main", 0},
		{"a", 1},
		{"b", 1},         // same depth as a (flattened chain)
		{"standalone", 0}, // standalone at root level
	}

	if len(entries) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(entries))
	}

	for i, want := range expected {
		if entries[i].branch.Name != want.name {
			t.Errorf("entry[%d].name = %q, want %q", i, entries[i].branch.Name, want.name)
		}
		if entries[i].depth != want.depth {
			t.Errorf("entry[%d].depth = %d, want %d", i, entries[i].depth, want.depth)
		}
	}
}
