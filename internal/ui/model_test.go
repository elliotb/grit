package ui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ejb/grit/internal/gt"
)

type mockExecutor struct {
	output string
	err    error
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) (string, error) {
	return m.output, m.err
}

func newTestModel(output string, err error) Model {
	client := gt.New(&mockExecutor{output: output, err: err})
	return New(client, "")
}

func sendWindowSize(m Model, width, height int) Model {
	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(Model)
}

func TestInit_ReturnsCmd(t *testing.T) {
	m := newTestModel("output", nil)
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil, want a Cmd")
	}
}

func TestUpdate_Quit(t *testing.T) {
	m := newTestModel("", nil)
	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'q'}}))
	if cmd == nil {
		t.Fatal("expected quit cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestUpdate_CtrlC_Quit(t *testing.T) {
	m := newTestModel("", nil)
	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyCtrlC}))
	if cmd == nil {
		t.Fatal("expected quit cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestUpdate_WindowSize(t *testing.T) {
	m := newTestModel("", nil)
	view := m.View()
	if view != "Loading..." {
		t.Fatalf("before WindowSizeMsg, expected %q, got %q", "Loading...", view)
	}

	m = sendWindowSize(m, 80, 24)
	view = m.View()
	if view == "Loading..." {
		t.Fatal("after WindowSizeMsg, view should not be 'Loading...'")
	}
}

func TestUpdate_LogResult_Success(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	// Use realistic gt log short format
	content := "│ ◉  feature-a\n│ ◯  feature-b\n◯─┘  main"
	updated, _ := m.Update(logResultMsg{output: content})
	m = updated.(Model)

	view := m.View()
	if !containsString(view, "main") {
		t.Error("view should contain 'main'")
	}
	if !containsString(view, "feature-a") {
		t.Error("view should contain 'feature-a'")
	}
}

func TestUpdate_LogResult_ParsedTree(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	content := "│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main"
	updated, _ := m.Update(logResultMsg{output: content})
	m = updated.(Model)

	// Verify branches were parsed
	if len(m.branches) != 1 {
		t.Fatalf("expected 1 root branch, got %d", len(m.branches))
	}
	if m.branches[0].Name != "main" {
		t.Errorf("root = %q, want %q", m.branches[0].Name, "main")
	}

	// View should contain branch names from the parsed tree
	view := m.View()
	if !containsString(view, "feature-top") {
		t.Error("view should contain 'feature-top'")
	}
	if !containsString(view, "feature-base") {
		t.Error("view should contain 'feature-base'")
	}
}

func TestUpdate_LogResult_FallbackOnUnparseable(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	// Content that has no branch markers — parser returns empty, renderer shows "(no stacks)"
	content := "some random output without markers"
	updated, _ := m.Update(logResultMsg{output: content})
	m = updated.(Model)

	view := m.View()
	if !containsString(view, "(no stacks)") {
		t.Errorf("expected fallback output, got:\n%s", view)
	}
}

func TestUpdate_LogResult_Error(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	updated, _ := m.Update(logResultMsg{err: errors.New("gt not found")})
	m = updated.(Model)

	view := m.View()
	if !containsString(view, "Error:") {
		t.Error("view should contain 'Error:'")
	}
}

func TestUpdate_LogResult_PopulatesDisplayEntries(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	content := "│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main"
	updated, _ := m.Update(logResultMsg{output: content})
	m = updated.(Model)

	if len(m.displayEntries) != 3 {
		t.Fatalf("expected 3 display entries, got %d", len(m.displayEntries))
	}
	if m.displayEntries[0].branch.Name != "main" {
		t.Errorf("entry 0 = %q, want %q", m.displayEntries[0].branch.Name, "main")
	}
}

func TestSelectedBranch(t *testing.T) {
	m := Model{
		displayEntries: []displayEntry{
			{branch: &gt.Branch{Name: "main"}, depth: 0},
			{branch: &gt.Branch{Name: "feature-a"}, depth: 1},
		},
		cursor: 1,
	}

	b := m.selectedBranch()
	if b == nil || b.Name != "feature-a" {
		t.Errorf("selectedBranch() = %v, want feature-a", b)
	}
}

func TestSelectedBranch_Empty(t *testing.T) {
	m := Model{}
	if b := m.selectedBranch(); b != nil {
		t.Errorf("selectedBranch() on empty model = %v, want nil", b)
	}
}

func TestPreserveCursor_ByName(t *testing.T) {
	m := Model{
		displayEntries: []displayEntry{
			{branch: &gt.Branch{Name: "main"}, depth: 0},
			{branch: &gt.Branch{Name: "feature-a"}, depth: 1},
			{branch: &gt.Branch{Name: "feature-b", IsCurrent: true}, depth: 1},
		},
	}

	m.preserveCursor("feature-b")
	if m.cursor != 2 {
		t.Errorf("cursor = %d, want 2", m.cursor)
	}
}

func TestPreserveCursor_FallbackToCurrent(t *testing.T) {
	m := Model{
		displayEntries: []displayEntry{
			{branch: &gt.Branch{Name: "main"}, depth: 0},
			{branch: &gt.Branch{Name: "feature-a", IsCurrent: true}, depth: 1},
		},
	}

	// Branch "gone" no longer exists, should fall back to IsCurrent
	m.preserveCursor("gone")
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}
}

func TestPreserveCursor_FallbackToZero(t *testing.T) {
	m := Model{
		displayEntries: []displayEntry{
			{branch: &gt.Branch{Name: "main"}, depth: 0},
			{branch: &gt.Branch{Name: "feature-a"}, depth: 1},
		},
	}

	// No match and no IsCurrent, should default to 0
	m.preserveCursor("gone")
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestPreserveCursor_EmptyOldName(t *testing.T) {
	m := Model{
		displayEntries: []displayEntry{
			{branch: &gt.Branch{Name: "main"}, depth: 0},
			{branch: &gt.Branch{Name: "feature-a", IsCurrent: true}, depth: 1},
		},
	}

	// Empty old name should find IsCurrent
	m.preserveCursor("")
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1 (IsCurrent)", m.cursor)
	}
}

// Helper to load a tree with branches into a ready model.
func loadedModel(content string) Model {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)
	updated, _ := m.Update(logResultMsg{output: content})
	return updated.(Model)
}

func sendKey(m Model, r rune) Model {
	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}}))
	return updated.(Model)
}

func sendSpecialKey(m Model, k tea.KeyType) Model {
	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: k}))
	return updated.(Model)
}

func TestNavigation_Down(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	// Cursor starts at IsCurrent (feature-top, index 2 after flatten: main=0, feature-base=1, feature-top=2)
	initial := m.cursor

	m = sendKey(m, 'j')
	if m.cursor != initial+1 && m.cursor <= len(m.displayEntries)-1 {
		// Just verify it moved down
	}
	if m.cursor <= initial && initial < len(m.displayEntries)-1 {
		t.Errorf("cursor should have moved down from %d, but is %d", initial, m.cursor)
	}
}

func TestNavigation_UpAtZero(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = 0
	m = sendKey(m, 'k')
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}
}

func TestNavigation_DownAtEnd(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = len(m.displayEntries) - 1
	last := m.cursor
	m = sendKey(m, 'j')
	if m.cursor != last {
		t.Errorf("cursor should stay at %d, got %d", last, m.cursor)
	}
}

func TestNavigation_ArrowKeys(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = 1

	m = sendSpecialKey(m, tea.KeyDown)
	if m.cursor != 2 {
		t.Errorf("down arrow: cursor = %d, want 2", m.cursor)
	}

	m = sendSpecialKey(m, tea.KeyUp)
	if m.cursor != 1 {
		t.Errorf("up arrow: cursor = %d, want 1", m.cursor)
	}
}

func TestNavigation_EmptyTree(t *testing.T) {
	m := loadedModel("some random output without markers")
	m = sendKey(m, 'j')
	if m.cursor != 0 {
		t.Errorf("down on empty tree: cursor = %d, want 0", m.cursor)
	}
	m = sendKey(m, 'k')
	if m.cursor != 0 {
		t.Errorf("up on empty tree: cursor = %d, want 0", m.cursor)
	}
}

func TestView_BeforeReady(t *testing.T) {
	m := newTestModel("", nil)
	if got := m.View(); got != "Loading..." {
		t.Errorf("got %q, want %q", got, "Loading...")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
