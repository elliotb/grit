package gt

import (
	"testing"
)

func TestParseLogShort_Empty(t *testing.T) {
	branches, err := ParseLogShort("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 0 {
		t.Fatalf("expected 0 roots, got %d", len(branches))
	}
}

func TestParseLogShort_SingleBranch(t *testing.T) {
	input := "◉  master"
	branches, err := ParseLogShort(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 root, got %d", len(branches))
	}
	root := branches[0]
	if root.Name != "master" {
		t.Errorf("root name = %q, want %q", root.Name, "master")
	}
	if !root.IsCurrent {
		t.Error("root should be current branch")
	}
	if len(root.Children) != 0 {
		t.Errorf("expected 0 children, got %d", len(root.Children))
	}
}

func TestParseLogShort_SingleBranchNotCurrent(t *testing.T) {
	input := "◯  main"
	branches, err := ParseLogShort(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 root, got %d", len(branches))
	}
	if branches[0].IsCurrent {
		t.Error("branch should not be current")
	}
}

func TestParseLogShort_LinearStack(t *testing.T) {
	// A simple stack: main → feature-a → feature-b → feature-c (current)
	input := `│ ◉  feature-c
│ ◯  feature-b
│ ◯  feature-a
◯─┘  main`

	branches, err := ParseLogShort(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 root, got %d", len(branches))
	}

	// root = main
	root := branches[0]
	if root.Name != "main" {
		t.Errorf("root = %q, want %q", root.Name, "main")
	}
	if root.IsCurrent {
		t.Error("main should not be current")
	}

	// main → feature-a
	if len(root.Children) != 1 {
		t.Fatalf("main children = %d, want 1", len(root.Children))
	}
	a := root.Children[0]
	if a.Name != "feature-a" {
		t.Errorf("child 0 = %q, want %q", a.Name, "feature-a")
	}
	if a.IsCurrent {
		t.Error("feature-a should not be current")
	}

	// feature-a → feature-b
	if len(a.Children) != 1 {
		t.Fatalf("feature-a children = %d, want 1", len(a.Children))
	}
	b := a.Children[0]
	if b.Name != "feature-b" {
		t.Errorf("child = %q, want %q", b.Name, "feature-b")
	}

	// feature-b → feature-c
	if len(b.Children) != 1 {
		t.Fatalf("feature-b children = %d, want 1", len(b.Children))
	}
	c := b.Children[0]
	if c.Name != "feature-c" {
		t.Errorf("child = %q, want %q", c.Name, "feature-c")
	}
	if !c.IsCurrent {
		t.Error("feature-c should be current")
	}
	if len(c.Children) != 0 {
		t.Errorf("feature-c children = %d, want 0", len(c.Children))
	}
}

func TestParseLogShort_MultipleStacks(t *testing.T) {
	// Two stacks off master:
	// master → upgrade_elixir (standalone)
	// master → usage_rules → add_deps → credo (chain)
	input := `◯    upgrade_elixir
│ ◉  credo
│ ◯  add_deps
│ ◯  usage_rules
◯─┘  master`

	branches, err := ParseLogShort(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 root, got %d", len(branches))
	}

	root := branches[0]
	if root.Name != "master" {
		t.Errorf("root = %q, want %q", root.Name, "master")
	}

	// master has 2 children: usage_rules (stack) and upgrade_elixir (standalone)
	if len(root.Children) != 2 {
		t.Fatalf("master children = %d, want 2", len(root.Children))
	}

	// First child (from bottom-to-top after reversal): usage_rules
	usageRules := root.Children[0]
	if usageRules.Name != "usage_rules" {
		t.Errorf("first child = %q, want %q", usageRules.Name, "usage_rules")
	}

	// usage_rules → add_deps
	if len(usageRules.Children) != 1 {
		t.Fatalf("usage_rules children = %d, want 1", len(usageRules.Children))
	}
	addDeps := usageRules.Children[0]
	if addDeps.Name != "add_deps" {
		t.Errorf("child = %q, want %q", addDeps.Name, "add_deps")
	}

	// add_deps → credo
	if len(addDeps.Children) != 1 {
		t.Fatalf("add_deps children = %d, want 1", len(addDeps.Children))
	}
	credo := addDeps.Children[0]
	if credo.Name != "credo" {
		t.Errorf("child = %q, want %q", credo.Name, "credo")
	}
	if !credo.IsCurrent {
		t.Error("credo should be current")
	}

	// Second child of master: upgrade_elixir
	upgradeElixir := root.Children[1]
	if upgradeElixir.Name != "upgrade_elixir" {
		t.Errorf("second child = %q, want %q", upgradeElixir.Name, "upgrade_elixir")
	}
	if len(upgradeElixir.Children) != 0 {
		t.Errorf("upgrade_elixir children = %d, want 0", len(upgradeElixir.Children))
	}
}

func TestParseLogShort_RealOutput(t *testing.T) {
	// Real output from ogat_app
	input := `◯    02-04-upgrade_elixir_to_1.20.0-rc.1
│ ◉  02-16-update_credo_to_latest_on_master_branch
│ ◯  02-16-update_tidewave_from_0.5.4_to_0.5.5
│ ◯  02-16-update_oban_web_from_2.11.7_to_2.11.8
│ ◯  02-16-update_live_debugger_0.6.0_phoenix_live_view_1.1.24_plug_cowboy_2.8.0
│ ◯  02-16-update_lazy_html_from_0.1.8_to_0.1.10
│ ◯  02-16-update_langchain_from_0.5.1_to_0.5.2
│ ◯  02-16-update_ex_cldr_from_2.46.0_to_2.47.0
│ ◯  02-16-add_update-deps_claude_code_skill
│ ◯  02-16-update_usage_rules_to_v1.1.0_and_migrate_to_project_config
◯─┘  master`

	branches, err := ParseLogShort(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 root, got %d", len(branches))
	}

	root := branches[0]
	if root.Name != "master" {
		t.Errorf("root = %q, want %q", root.Name, "master")
	}

	// master has 2 children: the main stack + upgrade_elixir
	if len(root.Children) != 2 {
		t.Fatalf("master children = %d, want 2", len(root.Children))
	}

	// Verify the stack chain depth: usage_rules → ... → credo (9 branches deep)
	branch := root.Children[0] // usage_rules
	if branch.Name != "02-16-update_usage_rules_to_v1.1.0_and_migrate_to_project_config" {
		t.Errorf("first child = %q, want usage_rules branch", branch.Name)
	}

	// Walk the chain to the top
	expectedChain := []string{
		"02-16-add_update-deps_claude_code_skill",
		"02-16-update_ex_cldr_from_2.46.0_to_2.47.0",
		"02-16-update_langchain_from_0.5.1_to_0.5.2",
		"02-16-update_lazy_html_from_0.1.8_to_0.1.10",
		"02-16-update_live_debugger_0.6.0_phoenix_live_view_1.1.24_plug_cowboy_2.8.0",
		"02-16-update_oban_web_from_2.11.7_to_2.11.8",
		"02-16-update_tidewave_from_0.5.4_to_0.5.5",
		"02-16-update_credo_to_latest_on_master_branch",
	}

	for _, expected := range expectedChain {
		if len(branch.Children) != 1 {
			t.Fatalf("branch %q children = %d, want 1", branch.Name, len(branch.Children))
		}
		branch = branch.Children[0]
		if branch.Name != expected {
			t.Errorf("got %q, want %q", branch.Name, expected)
		}
	}

	// Credo should be current and have no children
	if !branch.IsCurrent {
		t.Error("credo branch should be current")
	}
	if len(branch.Children) != 0 {
		t.Errorf("credo children = %d, want 0", len(branch.Children))
	}

	// Second child of master: upgrade_elixir
	upgrade := root.Children[1]
	if upgrade.Name != "02-04-upgrade_elixir_to_1.20.0-rc.1" {
		t.Errorf("second child = %q, want upgrade_elixir branch", upgrade.Name)
	}
}

func TestParseLogShort_CurrentBranchDetection(t *testing.T) {
	input := `│ ◯  feature-c
│ ◉  feature-b
│ ◯  feature-a
◯─┘  main`

	branches, err := ParseLogShort(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count current branches
	currentCount := countCurrent(branches[0])
	if currentCount != 1 {
		t.Errorf("expected exactly 1 current branch, got %d", currentCount)
	}

	// Verify it's feature-b
	a := branches[0].Children[0] // feature-a
	b := a.Children[0]           // feature-b
	if !b.IsCurrent {
		t.Error("feature-b should be current")
	}
	if b.Name != "feature-b" {
		t.Errorf("current branch name = %q, want %q", b.Name, "feature-b")
	}
}

func TestParseLogShort_ConnectorOnlyLines(t *testing.T) {
	// Lines with only connectors (no branch marker) should be skipped
	input := `│ ◉  feature-b
│
│ ◯  feature-a
│
◯─┘  main`

	branches, err := ParseLogShort(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 root, got %d", len(branches))
	}
	if branches[0].Name != "main" {
		t.Errorf("root = %q, want %q", branches[0].Name, "main")
	}
	if len(branches[0].Children) != 1 {
		t.Fatalf("main children = %d, want 1", len(branches[0].Children))
	}
}

func TestParseLogShort_WhitespaceOnlyInput(t *testing.T) {
	branches, err := ParseLogShort("   \n  \n\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 0 {
		t.Fatalf("expected 0 roots, got %d", len(branches))
	}
}

func TestParseLine_BranchWithConnectors(t *testing.T) {
	// The trunk line has ─┘ characters that should be stripped
	pl, ok := parseLine("◯─┘  master")
	if !ok {
		t.Fatal("expected line to parse")
	}
	if pl.name != "master" {
		t.Errorf("name = %q, want %q", pl.name, "master")
	}
	if pl.depth != 0 {
		t.Errorf("depth = %d, want 0", pl.depth)
	}
	if pl.isCurrent {
		t.Error("should not be current")
	}
}

func TestParseLine_IndentedCurrent(t *testing.T) {
	pl, ok := parseLine("│ ◉  feature-branch")
	if !ok {
		t.Fatal("expected line to parse")
	}
	if pl.name != "feature-branch" {
		t.Errorf("name = %q, want %q", pl.name, "feature-branch")
	}
	if pl.depth != 1 {
		t.Errorf("depth = %d, want 1", pl.depth)
	}
	if !pl.isCurrent {
		t.Error("should be current")
	}
}

func TestParseLine_NoMarker(t *testing.T) {
	_, ok := parseLine("│")
	if ok {
		t.Error("expected line without marker to not parse")
	}

	_, ok = parseLine("")
	if ok {
		t.Error("expected empty line to not parse")
	}

	_, ok = parseLine("   │   ")
	if ok {
		t.Error("expected connector-only line to not parse")
	}
}

func TestParseLine_DepthCalculation(t *testing.T) {
	tests := []struct {
		line  string
		depth int
	}{
		{"◯  branch-d0", 0},
		{"│ ◯  branch-d1", 1},
		{"│ │ ◯  branch-d2", 2},
	}

	for _, tt := range tests {
		pl, ok := parseLine(tt.line)
		if !ok {
			t.Errorf("parseLine(%q): expected ok", tt.line)
			continue
		}
		if pl.depth != tt.depth {
			t.Errorf("parseLine(%q): depth = %d, want %d", tt.line, pl.depth, tt.depth)
		}
	}
}

func TestExtractAnnotation(t *testing.T) {
	tests := []struct {
		input          string
		wantName       string
		wantAnnotation string
	}{
		{"my-branch", "my-branch", ""},
		{"my-branch (merging)", "my-branch", "merging"},
		{"my-branch (needs restack)", "my-branch", "needs restack"},
		{"my-branch (rebasing)", "my-branch", "rebasing"},
		{"branch-with-(parens)-in-name", "branch-with-(parens)-in-name", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		gotName, gotAnnotation := extractAnnotation(tt.input)
		if gotName != tt.wantName {
			t.Errorf("extractAnnotation(%q) name = %q, want %q", tt.input, gotName, tt.wantName)
		}
		if gotAnnotation != tt.wantAnnotation {
			t.Errorf("extractAnnotation(%q) annotation = %q, want %q", tt.input, gotAnnotation, tt.wantAnnotation)
		}
	}
}

func TestParseLine_WithAnnotation(t *testing.T) {
	line := "│ ◯  my-branch (merging)"
	pl, ok := parseLine(line)
	if !ok {
		t.Fatal("expected valid parse")
	}
	if pl.name != "my-branch" {
		t.Errorf("name = %q, want %q", pl.name, "my-branch")
	}
	if pl.annotation != "merging" {
		t.Errorf("annotation = %q, want %q", pl.annotation, "merging")
	}
}

func TestParseLogShort_BranchWithAnnotation(t *testing.T) {
	input := "│ ◉  feature-top (needs restack)\n│ ◯  feature-base\n◯─┘  main"
	branches, err := ParseLogShort(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("expected 1 root, got %d", len(branches))
	}
	// Walk to find feature-top
	featureBase := branches[0].Children[0]
	if featureBase.Name != "feature-base" {
		t.Errorf("expected feature-base, got %q", featureBase.Name)
	}
	if featureBase.Annotation != "" {
		t.Errorf("feature-base annotation = %q, want empty", featureBase.Annotation)
	}
	featureTop := featureBase.Children[0]
	if featureTop.Name != "feature-top" {
		t.Errorf("expected feature-top (annotation stripped), got %q", featureTop.Name)
	}
	if featureTop.Annotation != "needs restack" {
		t.Errorf("feature-top annotation = %q, want %q", featureTop.Annotation, "needs restack")
	}
}

// countCurrent recursively counts branches with IsCurrent == true.
func countCurrent(b *Branch) int {
	count := 0
	if b.IsCurrent {
		count++
	}
	for _, child := range b.Children {
		count += countCurrent(child)
	}
	return count
}
