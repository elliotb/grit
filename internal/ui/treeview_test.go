package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/ejb/grit/internal/gt"
)

// renderTreeFromBranches is a test helper that flattens branches and renders
// with cursor at position 0 (default).
func renderTreeFromBranches(branches []*gt.Branch) string {
	entries := flattenForDisplay(branches)
	return renderTree(entries, 0)
}

func TestRenderTree_Empty(t *testing.T) {
	result := renderTree(nil, 0)
	if result != "(no stacks)" {
		t.Errorf("got %q, want %q", result, "(no stacks)")
	}
}

func TestRenderTree_SingleRoot(t *testing.T) {
	branches := []*gt.Branch{
		{Name: "main", IsCurrent: true},
	}
	result := ansi.Strip(renderTreeFromBranches(branches))
	if !strings.Contains(result, "◉ main") {
		t.Errorf("output should contain '◉ main', got:\n%s", result)
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

	result := ansi.Strip(renderTreeFromBranches(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), result)
	}

	if strings.HasPrefix(lines[0], "│") {
		t.Errorf("root should not have │ prefix, got: %q", lines[0])
	}
	if !strings.Contains(lines[0], "main") {
		t.Errorf("first line should contain 'main', got: %q", lines[0])
	}

	for _, i := range []int{1, 2} {
		if !strings.HasPrefix(lines[i], "│ ") {
			t.Errorf("line %d should start with '│ ', got: %q", i, lines[i])
		}
	}

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

	result := ansi.Strip(renderTreeFromBranches(branches))

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
	branches := []*gt.Branch{
		{
			Name: "main",
			Children: []*gt.Branch{
				{Name: "standalone-fix"},
			},
		},
	}

	result := ansi.Strip(renderTreeFromBranches(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), result)
	}

	if strings.HasPrefix(lines[1], "│") {
		t.Errorf("standalone branch should not have │ prefix, got: %q", lines[1])
	}
	if !strings.Contains(lines[1], "standalone-fix") {
		t.Errorf("line 1 should contain 'standalone-fix', got: %q", lines[1])
	}
}

func TestRenderTree_MultipleStacks(t *testing.T) {
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

	result := ansi.Strip(renderTreeFromBranches(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d:\n%s", len(lines), result)
	}

	if !strings.HasPrefix(lines[1], "│ ") {
		t.Errorf("stack-base should have │ prefix, got: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "│ ") {
		t.Errorf("stack-top should have │ prefix, got: %q", lines[2])
	}

	if strings.HasPrefix(lines[3], "│") {
		t.Errorf("standalone should not have │ prefix, got: %q", lines[3])
	}
}

func TestRenderTree_DeepChainFlattened(t *testing.T) {
	e := &gt.Branch{Name: "e", IsCurrent: true}
	d := &gt.Branch{Name: "d", Children: []*gt.Branch{e}}
	c := &gt.Branch{Name: "c", Children: []*gt.Branch{d}}
	b := &gt.Branch{Name: "b", Children: []*gt.Branch{c}}
	a := &gt.Branch{Name: "a", Children: []*gt.Branch{b}}
	branches := []*gt.Branch{
		{Name: "main", Children: []*gt.Branch{a}},
	}

	result := ansi.Strip(renderTreeFromBranches(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d:\n%s", len(lines), result)
	}

	if strings.HasPrefix(lines[0], "│") {
		t.Errorf("main should not have │ prefix")
	}

	for i := 1; i <= 5; i++ {
		if !strings.HasPrefix(lines[i], "│ ") {
			t.Errorf("line %d should start with '│ ', got: %q", i, lines[i])
		}
		if strings.HasPrefix(lines[i], "│ │") {
			t.Errorf("line %d should not be double-nested, got: %q", i, lines[i])
		}
	}
}

func TestRenderTree_CursorHighlight(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
		{branch: &gt.Branch{Name: "feature-a"}, depth: 1},
		{branch: &gt.Branch{Name: "feature-b", IsCurrent: true}, depth: 1},
	}

	// Cursor on middle entry
	result := ansi.Strip(renderTree(entries, 1))
	lines := strings.Split(result, "\n")

	// All entries should still be present
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "main") {
		t.Errorf("line 0 should contain 'main', got %q", lines[0])
	}
	if !strings.Contains(lines[1], "feature-a") {
		t.Errorf("line 1 should contain 'feature-a', got %q", lines[1])
	}
	if !strings.Contains(lines[2], "feature-b") {
		t.Errorf("line 2 should contain 'feature-b', got %q", lines[2])
	}
}

func TestRenderTree_CursorOutOfRange(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
	}

	// Cursor beyond range — no entry should get selected style
	result := ansi.Strip(renderTree(entries, 5))
	if !strings.Contains(result, "◯ main") {
		t.Errorf("out-of-range cursor should render normally, got:\n%s", result)
	}
}

func TestRenderTree_WithAnnotation(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
		{branch: &gt.Branch{Name: "feature-a", Annotation: "needs restack"}, depth: 1},
	}

	result := ansi.Strip(renderTree(entries, 0))
	if !strings.Contains(result, "(needs restack)") {
		t.Errorf("output should contain annotation, got:\n%s", result)
	}
	if !strings.Contains(result, "feature-a") {
		t.Errorf("output should contain branch name, got:\n%s", result)
	}
}

func TestRenderTree_AnnotationInSelectedBranch(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
		{branch: &gt.Branch{Name: "feature-a", Annotation: "merging"}, depth: 1},
	}

	// Cursor on annotated branch
	result := ansi.Strip(renderTree(entries, 1))
	if !strings.Contains(result, "(merging)") {
		t.Errorf("selected branch should include annotation, got:\n%s", result)
	}
}

func TestRenderTree_NoAnnotation(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
		{branch: &gt.Branch{Name: "feature-a"}, depth: 1},
	}

	result := ansi.Strip(renderTree(entries, 0))
	if strings.Contains(result, "(") {
		t.Errorf("output should not contain parentheses when no annotation, got:\n%s", result)
	}
}

func TestRenderTree_WithPRInfo(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
		{branch: &gt.Branch{Name: "feature-a", PR: gt.PRInfo{Number: 142, State: "OPEN"}}, depth: 1},
		{branch: &gt.Branch{Name: "feature-b", PR: gt.PRInfo{Number: 143, State: "DRAFT"}}, depth: 1},
	}

	result := ansi.Strip(renderTree(entries, 0))
	if !strings.Contains(result, "#142 open") {
		t.Errorf("output should contain '#142 open', got:\n%s", result)
	}
	if !strings.Contains(result, "#143 draft") {
		t.Errorf("output should contain '#143 draft', got:\n%s", result)
	}
}

func TestRenderTree_PRInfoInSelectedBranch(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
		{branch: &gt.Branch{Name: "feature-a", PR: gt.PRInfo{Number: 100, State: "MERGED"}}, depth: 1},
	}

	result := ansi.Strip(renderTree(entries, 1))
	if !strings.Contains(result, "#100 merged") {
		t.Errorf("selected branch should include PR info, got:\n%s", result)
	}
}

func TestRenderTree_NoPRInfo(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
		{branch: &gt.Branch{Name: "feature-a"}, depth: 1},
	}

	result := ansi.Strip(renderTree(entries, 0))
	if strings.Contains(result, "#") {
		t.Errorf("output should not contain '#' when no PR, got:\n%s", result)
	}
}

func TestPrLabel_States(t *testing.T) {
	tests := []struct {
		pr   gt.PRInfo
		want string
	}{
		{gt.PRInfo{Number: 42, State: "OPEN"}, "#42 open"},
		{gt.PRInfo{Number: 99, State: "DRAFT"}, "#99 draft"},
		{gt.PRInfo{Number: 10, State: "MERGED"}, "#10 merged"},
		{gt.PRInfo{Number: 5, State: "CLOSED"}, "#5 closed"},
		{gt.PRInfo{}, ""},
	}

	for _, tt := range tests {
		got := ansi.Strip(prLabel(tt.pr))
		got = strings.TrimSpace(got)
		if got != tt.want {
			t.Errorf("prLabel(%v) = %q, want %q", tt.pr, got, tt.want)
		}
	}
}

func TestRenderTree_AnnotationAndPR(t *testing.T) {
	entries := []displayEntry{
		{branch: &gt.Branch{Name: "main"}, depth: 0},
		{branch: &gt.Branch{
			Name:       "feature-a",
			Annotation: "needs restack",
			PR:         gt.PRInfo{Number: 142, State: "OPEN"},
		}, depth: 1},
	}

	result := ansi.Strip(renderTree(entries, 0))
	if !strings.Contains(result, "(needs restack)") {
		t.Errorf("output should contain annotation, got:\n%s", result)
	}
	if !strings.Contains(result, "#142 open") {
		t.Errorf("output should contain PR info, got:\n%s", result)
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
		{"b", 1},          // same depth as a (flattened chain)
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
