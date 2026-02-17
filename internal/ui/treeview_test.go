package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/elliotb/grit/internal/gt"
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
	// gt log short order: feature-b (top), feature-a, main (trunk)
	branches := []*gt.Branch{
		{
			Name:  "main",
			Order: 2,
			Children: []*gt.Branch{
				{
					Name:  "feature-a",
					Depth: 1,
					Order: 1,
					Children: []*gt.Branch{
						{
							Name:      "feature-b",
							Depth:     1,
							Order:     0,
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

	// gt ls order: top-of-stack first, trunk last
	if !strings.Contains(lines[0], "feature-b") {
		t.Errorf("line 0 should contain 'feature-b' (top of stack), got: %q", lines[0])
	}
	if !strings.Contains(lines[1], "feature-a") {
		t.Errorf("line 1 should contain 'feature-a', got: %q", lines[1])
	}
	if !strings.Contains(lines[2], "main") {
		t.Errorf("line 2 should contain 'main' (trunk), got: %q", lines[2])
	}

	// feature-b and feature-a at depth 1, main at depth 0
	for _, i := range []int{0, 1} {
		if !strings.HasPrefix(lines[i], "│ ") {
			t.Errorf("line %d should start with '│ ', got: %q", i, lines[i])
		}
	}
	if strings.HasPrefix(lines[2], "│") {
		t.Errorf("trunk should not have │ prefix, got: %q", lines[2])
	}
}

func TestRenderTree_CurrentBranchMarker(t *testing.T) {
	branches := []*gt.Branch{
		{
			Name:  "main",
			Order: 2,
			Children: []*gt.Branch{
				{Name: "feature-a", Depth: 1, Order: 1},
				{Name: "feature-b", Depth: 1, Order: 0, IsCurrent: true},
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
	// A standalone branch at depth 0 (like in a multi-stack gt output)
	branches := []*gt.Branch{
		{
			Name:  "main",
			Order: 1,
			Children: []*gt.Branch{
				{Name: "standalone-fix", Order: 0},
			},
		},
	}

	result := ansi.Strip(renderTreeFromBranches(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), result)
	}

	// standalone-fix first (Order 0), main last (Order 1)
	if !strings.Contains(lines[0], "standalone-fix") {
		t.Errorf("line 0 should contain 'standalone-fix', got: %q", lines[0])
	}
	if strings.HasPrefix(lines[0], "│") {
		t.Errorf("standalone branch at depth 0 should not have │ prefix, got: %q", lines[0])
	}
}

func TestRenderTree_MultipleStacks(t *testing.T) {
	// gt log short order: standalone, stack-top, stack-base, main
	branches := []*gt.Branch{
		{
			Name:  "main",
			Order: 3,
			Children: []*gt.Branch{
				{
					Name:  "stack-base",
					Depth: 1,
					Order: 2,
					Children: []*gt.Branch{
						{Name: "stack-top", Depth: 1, Order: 1, IsCurrent: true},
					},
				},
				{Name: "standalone", Order: 0},
			},
		},
	}

	result := ansi.Strip(renderTreeFromBranches(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d:\n%s", len(lines), result)
	}

	// gt ls order: standalone (depth 0), stack-top (depth 1), stack-base (depth 1), main (depth 0)
	if strings.HasPrefix(lines[0], "│") {
		t.Errorf("standalone should not have │ prefix, got: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "│ ") {
		t.Errorf("stack-top should have │ prefix, got: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "│ ") {
		t.Errorf("stack-base should have │ prefix, got: %q", lines[2])
	}
	if strings.HasPrefix(lines[3], "│") {
		t.Errorf("main (trunk) should not have │ prefix, got: %q", lines[3])
	}
}

func TestRenderTree_DeepChainFlattened(t *testing.T) {
	// gt log short order: e (top), d, c, b, a, main (trunk)
	e := &gt.Branch{Name: "e", Depth: 1, Order: 0, IsCurrent: true}
	d := &gt.Branch{Name: "d", Depth: 1, Order: 1, Children: []*gt.Branch{e}}
	c := &gt.Branch{Name: "c", Depth: 1, Order: 2, Children: []*gt.Branch{d}}
	b := &gt.Branch{Name: "b", Depth: 1, Order: 3, Children: []*gt.Branch{c}}
	a := &gt.Branch{Name: "a", Depth: 1, Order: 4, Children: []*gt.Branch{b}}
	branches := []*gt.Branch{
		{Name: "main", Order: 5, Children: []*gt.Branch{a}},
	}

	result := ansi.Strip(renderTreeFromBranches(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d:\n%s", len(lines), result)
	}

	// Trunk (main) is last
	if strings.HasPrefix(lines[5], "│") {
		t.Errorf("main should not have │ prefix")
	}

	// All stack branches at depth 1
	for i := 0; i <= 4; i++ {
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
	// gt log short order: b (top), a, standalone, main (trunk)
	branches := []*gt.Branch{
		{
			Name:  "main",
			Order: 3,
			Children: []*gt.Branch{
				{
					Name:  "a",
					Depth: 1,
					Order: 1,
					Children: []*gt.Branch{
						{Name: "b", Depth: 1, Order: 0},
					},
				},
				{Name: "standalone", Order: 2},
			},
		},
	}

	entries := flattenForDisplay(branches)

	expected := []struct {
		name  string
		depth int
	}{
		{"b", 1},          // top of stack (Order 0)
		{"a", 1},          // base of stack (Order 1)
		{"standalone", 0}, // standalone at depth 0 (Order 2)
		{"main", 0},       // trunk (Order 3)
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

func TestRenderTree_BranchingStack(t *testing.T) {
	// gt log short output (and desired display order):
	//   ◯      02-04-upgrade_elixir          (depth 0, Order 0)
	//   │ ◯    add_kaffy_admin               (depth 1, Order 1)
	//   │ ◯    add_ks2_historical            (depth 1, Order 2)
	//   │ │ ◉  02-17-use_full-width          (depth 2, Order 3)
	//   ◯─┴─┘  master                        (depth 0, Order 4)
	branches := []*gt.Branch{
		{
			Name:  "master",
			Order: 4,
			Children: []*gt.Branch{
				{
					Name:  "add_ks2_historical",
					Depth: 1,
					Order: 2,
					Children: []*gt.Branch{
						{
							Name:  "add_kaffy_admin",
							Depth: 1,
							Order: 1,
							Children: []*gt.Branch{
								{Name: "02-17-use_full-width", Depth: 2, Order: 3, IsCurrent: true},
							},
						},
					},
				},
				{Name: "02-04-upgrade_elixir", Order: 0},
			},
		},
	}

	result := ansi.Strip(renderTreeFromBranches(branches))
	lines := strings.Split(result, "\n")

	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d:\n%s", len(lines), result)
	}

	// Display order matches gt ls: 02-04, add_kaffy, add_ks2, 02-17, master
	if !strings.Contains(lines[0], "02-04-upgrade_elixir") {
		t.Errorf("line 0 should be 02-04 (standalone), got: %q", lines[0])
	}
	if !strings.Contains(lines[1], "add_kaffy_admin") {
		t.Errorf("line 1 should be add_kaffy (top of depth-1 stack), got: %q", lines[1])
	}
	if !strings.Contains(lines[2], "add_ks2_historical") {
		t.Errorf("line 2 should be add_ks2 (base of depth-1 stack), got: %q", lines[2])
	}
	if !strings.Contains(lines[3], "02-17-use_full-width") {
		t.Errorf("line 3 should be 02-17 (depth-2 branch), got: %q", lines[3])
	}
	if !strings.Contains(lines[4], "master") {
		t.Errorf("line 4 should be master (trunk), got: %q", lines[4])
	}

	// Depth checks
	if strings.HasPrefix(lines[0], "│") {
		t.Errorf("02-04 (depth 0) should not have │ prefix, got: %q", lines[0])
	}
	for _, i := range []int{1, 2} {
		if !strings.HasPrefix(lines[i], "│ ") {
			t.Errorf("line %d should start with '│ ', got: %q", i, lines[i])
		}
	}
	if !strings.HasPrefix(lines[3], "│ │ ") {
		t.Errorf("02-17 (depth 2) should start with '│ │ ', got: %q", lines[3])
	}
	if strings.HasPrefix(lines[4], "│") {
		t.Errorf("master (trunk) should not have │ prefix, got: %q", lines[4])
	}
}

func TestFlattenForDisplay_BranchingStackDepths(t *testing.T) {
	// Verify the exact display depths and order for a branching stack scenario.
	branches := []*gt.Branch{
		{
			Name:  "master",
			Order: 4,
			Children: []*gt.Branch{
				{
					Name:  "stack-a",
					Depth: 1,
					Order: 2,
					Children: []*gt.Branch{
						{
							Name:  "stack-a-top",
							Depth: 1,
							Order: 1,
							Children: []*gt.Branch{
								{Name: "branch-off", Depth: 2, Order: 3},
							},
						},
					},
				},
				{Name: "standalone", Order: 0},
			},
		},
	}

	entries := flattenForDisplay(branches)

	expected := []struct {
		name  string
		depth int
	}{
		{"standalone", 0},
		{"stack-a-top", 1},
		{"stack-a", 1},
		{"branch-off", 2},
		{"master", 0},
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
