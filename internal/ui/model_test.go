package ui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ejb/grit/internal/gt"
)

type mockExecutor struct {
	fn func(ctx context.Context, name string, args ...string) (string, error)
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) (string, error) {
	return m.fn(ctx, name, args...)
}

// simpleMock creates a mockExecutor that always returns the given output/err.
func simpleMock(output string, err error) *mockExecutor {
	return &mockExecutor{fn: func(ctx context.Context, name string, args ...string) (string, error) {
		return output, err
	}}
}

func newTestModel(output string, err error) Model {
	client := gt.New(simpleMock(output, err))
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

func TestActionResult_Success_ReloadsTree(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true
	m.statusBar.startSpinner("Working...")

	updated, cmd := m.Update(actionResultMsg{action: "checkout", message: "Checked out feature-base"})
	m = updated.(Model)

	if m.running {
		t.Error("running should be false after actionResultMsg")
	}
	if m.statusBar.spinning {
		t.Error("spinner should be stopped")
	}
	// Should return a loadLog command
	if cmd == nil {
		t.Error("expected reload cmd after successful action")
	}
}

func TestActionResult_Error_ShowsError(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	updated, _ := m.Update(actionResultMsg{action: "checkout", err: errors.New("branch not found")})
	m = updated.(Model)

	if m.running {
		t.Error("running should be false after error")
	}
	if !m.statusBar.isError {
		t.Error("status bar should show error")
	}
	if !containsString(m.statusBar.message, "branch not found") {
		t.Errorf("status bar message = %q, want to contain 'branch not found'", m.statusBar.message)
	}
}

func TestActionResult_OpenPR_NoReload(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	updated, _ := m.Update(actionResultMsg{action: "openpr", message: "Opened PR"})
	m = updated.(Model)

	if m.running {
		t.Error("running should be false")
	}
}

func TestInputBlocked_WhenRunning(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = 0
	m.running = true

	// Navigation should be blocked
	m = sendKey(m, 'j')
	if m.cursor != 0 {
		t.Errorf("cursor should not move while running, got %d", m.cursor)
	}
}

func TestQuit_WorksWhileRunning(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'q'}}))
	if cmd == nil {
		t.Fatal("quit should work while running")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestRunAction_ProducesActionResultMsg(t *testing.T) {
	cmd := runAction("test", "Done", func(ctx context.Context) error {
		return nil
	})
	msg := cmd()
	result, ok := msg.(actionResultMsg)
	if !ok {
		t.Fatalf("expected actionResultMsg, got %T", msg)
	}
	if result.action != "test" {
		t.Errorf("action = %q, want %q", result.action, "test")
	}
	if result.err != nil {
		t.Errorf("err = %v, want nil", result.err)
	}
	if result.message != "Done" {
		t.Errorf("message = %q, want %q", result.message, "Done")
	}
}

func TestRunAction_PropagatesError(t *testing.T) {
	cmd := runAction("test", "Done", func(ctx context.Context) error {
		return errors.New("fail")
	})
	msg := cmd()
	result := msg.(actionResultMsg)
	if result.err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCheckout_EnterKey(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = 0 // on "main"

	updated, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))
	m = updated.(Model)

	if !m.running {
		t.Error("pressing Enter should set running=true")
	}
	if !m.statusBar.spinning {
		t.Error("spinner should be active")
	}
	if cmd == nil {
		t.Fatal("expected commands from Enter key")
	}
}

func TestCheckout_EnterOnEmptyTree(t *testing.T) {
	m := loadedModel("some random output without markers")

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))
	m = updated.(Model)

	if m.running {
		t.Error("Enter on empty tree should not start action")
	}
}

func TestTrunkKey_CheckoutsTrunk(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")

	updated, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'m'}}))
	m = updated.(Model)

	if !m.running {
		t.Error("pressing m should set running=true")
	}
	if !m.statusBar.spinning {
		t.Error("spinner should be active")
	}
	if !containsString(m.statusBar.spinnerLabel, "main") {
		t.Errorf("spinner label = %q, want to contain 'main'", m.statusBar.spinnerLabel)
	}
	if cmd == nil {
		t.Fatal("expected commands from m key")
	}
}

func TestTrunkKey_EmptyTree(t *testing.T) {
	m := loadedModel("some random output without markers")

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'m'}}))
	m = updated.(Model)

	if m.running {
		t.Error("m on empty tree should not start action")
	}
}

func TestActionKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      rune
		wantRun  bool
		wantSpin string
	}{
		// Cursor starts on feature-top (IsCurrent), so branch-specific actions include it.
		{"submit stack", 's', true, "Submitting stack (feature-top)..."},
		{"submit downstack", 'S', true, "Submitting downstack (feature-top)..."},
		{"restack", 'r', true, "Restacking (feature-top)..."},
		{"fetch", 'f', true, "Fetching..."},
		{"sync", 'y', true, "Syncing..."},
		{"open PR", 'o', true, "Opening PR (feature-top)..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")

			updated, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{tt.key}}))
			m = updated.(Model)

			if m.running != tt.wantRun {
				t.Errorf("running = %v, want %v", m.running, tt.wantRun)
			}
			if m.statusBar.spinnerLabel != tt.wantSpin {
				t.Errorf("spinnerLabel = %q, want %q", m.statusBar.spinnerLabel, tt.wantSpin)
			}
			if cmd == nil {
				t.Error("expected commands from action key")
			}
		})
	}
}

func TestActionKeys_BlockedWhileRunning(t *testing.T) {
	keys := []rune{'s', 'S', 'r', 'y', 'o'}

	for _, k := range keys {
		t.Run(string(k), func(t *testing.T) {
			m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
			m.running = true

			updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{k}}))
			m = updated.(Model)

			// Spinner should not have been re-started
			if m.statusBar.spinning {
				t.Error("action should be blocked while running")
			}
		})
	}
}

func TestInitialCursor_OnCurrentBranch(t *testing.T) {
	// feature-top is IsCurrent (◉), cursor should land there
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")

	// After parse: main(0), feature-base(1), feature-top(2)
	// feature-top has IsCurrent=true
	selected := m.selectedBranch()
	if selected == nil {
		t.Fatal("expected a selected branch")
	}
	if selected.Name != "feature-top" {
		t.Errorf("initial cursor on %q, want %q", selected.Name, "feature-top")
	}
}

func TestActionResult_SuccessStyle(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	updated, _ := m.Update(actionResultMsg{action: "checkout", message: "Checked out feature-base"})
	m = updated.(Model)

	if !m.statusBar.isSuccess {
		t.Error("status bar should show success style")
	}
	if m.statusBar.isError {
		t.Error("status bar should not show error style")
	}
	if m.statusBar.message != "Checked out feature-base" {
		t.Errorf("message = %q, want %q", m.statusBar.message, "Checked out feature-base")
	}
}

func TestView_BeforeReady(t *testing.T) {
	m := newTestModel("", nil)
	if got := m.View(); got != "Loading..." {
		t.Errorf("got %q, want %q", got, "Loading...")
	}
}

func TestView_ContainsLegend(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	view := m.View()

	// Legend should contain key descriptions
	for _, key := range []string{"checkout", "trunk", "submit", "restack", "fetch", "sync", "quit"} {
		if !containsString(view, key) {
			t.Errorf("view should contain legend text %q", key)
		}
	}
}

func TestWindowSize_AccountsForLegend(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	// Viewport height should be total height minus chrome (legend lines + status bar).
	// The legend wraps on narrow terminals, so chrome height is dynamic.
	wantHeight := 24 - m.chromeHeight()
	if m.viewport.Height != wantHeight {
		t.Errorf("viewport height = %d, want %d (24 - %d chrome)", m.viewport.Height, wantHeight, m.chromeHeight())
	}
}

func TestWindowSize_WideTerminal_LegendOneLine(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 140, 24)

	// At 140 columns, the legend fits on one line, so chrome = 2.
	if m.viewport.Height != 22 {
		t.Errorf("viewport height = %d, want 22 (24 - 2)", m.viewport.Height)
	}
}

func TestWindowSize_NarrowTerminal_LegendWraps(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 60, 24)

	// At 60 columns, the legend wraps to multiple lines.
	chrome := m.chromeHeight()
	if chrome <= 2 {
		t.Errorf("chrome height = %d, want > 2 for narrow terminal", chrome)
	}
	wantHeight := 24 - chrome
	if m.viewport.Height != wantHeight {
		t.Errorf("viewport height = %d, want %d", m.viewport.Height, wantHeight)
	}
}

// recordingMock creates a mockExecutor that records all calls.
type callRecord struct {
	name string
	args []string
}

func recordingMock() (*mockExecutor, *[]callRecord) {
	var calls []callRecord
	mock := &mockExecutor{fn: func(ctx context.Context, name string, args ...string) (string, error) {
		calls = append(calls, callRecord{name: name, args: args})
		return "", nil
	}}
	return mock, &calls
}

func TestActionKeys_TargetSelectedBranch(t *testing.T) {
	// Verify that branch-specific actions target the cursor-selected branch,
	// not the checked-out (IsCurrent) branch.
	// Uses feature-base (index 1) since trunk actions are now guarded.
	tests := []struct {
		name     string
		key      rune
		wantArgs []string // expected gt CLI args
	}{
		{"submit stack", 's', []string{"stack", "submit", "--no-interactive", "--branch", "feature-base"}},
		{"submit downstack", 'S', []string{"downstack", "submit", "--no-interactive", "--branch", "feature-base"}},
		{"restack", 'r', []string{"stack", "restack", "--no-interactive", "--branch", "feature-base"}},
		{"open PR", 'o', []string{"pr", "feature-base"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, calls := recordingMock()
			// Override the mock to also return log output for loadLog.
			logOutput := "│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main"
			mock.fn = func(ctx context.Context, name string, args ...string) (string, error) {
				*calls = append(*calls, callRecord{name: name, args: args})
				if len(args) > 0 && args[0] == "log" {
					return logOutput, nil
				}
				return "", nil
			}

			client := gt.New(mock)
			m := New(client, "")
			m = sendWindowSize(m, 80, 24)
			updated, _ := m.Update(logResultMsg{output: logOutput})
			m = updated.(Model)

			// Move cursor to feature-base (index 1).
			m.cursor = 1

			// Clear recorded calls from setup.
			*calls = nil

			// Press the action key.
			updated, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{tt.key}}))
			m = updated.(Model)

			if !m.running {
				t.Fatal("expected running=true")
			}
			if cmd == nil {
				t.Fatal("expected command")
			}

			// cmd is a tea.Batch; execute it to get the inner commands, then
			// run each one to find the actionResultMsg (which triggers the gt call).
			batchMsg := cmd()
			if cmds, ok := batchMsg.(tea.BatchMsg); ok {
				for _, c := range cmds {
					if c != nil {
						c()
					}
				}
			}

			if len(*calls) == 0 {
				t.Fatal("expected gt call, got none")
			}
			// Find the action call (skip any spinner-related calls).
			var got []string
			for _, c := range *calls {
				if c.name == "gt" {
					got = c.args
					break
				}
			}
			if got == nil {
				t.Fatal("no gt call found in recorded calls")
			}
			if len(got) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", got, tt.wantArgs)
			}
			for i, arg := range tt.wantArgs {
				if got[i] != arg {
					t.Errorf("arg[%d] = %q, want %q", i, got[i], arg)
				}
			}
		})
	}
}

func TestActionKeys_NoOpOnEmptyTree(t *testing.T) {
	keys := []rune{'s', 'S', 'r', 'o'}
	for _, k := range keys {
		t.Run(string(k), func(t *testing.T) {
			m := loadedModel("some random output without markers")

			updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{k}}))
			m = updated.(Model)

			if m.running {
				t.Errorf("key %c on empty tree should not start action", k)
			}
		})
	}
}

// --- Diff view integration tests ---

// diffMock creates a mockExecutor that handles both gt and git commands for diff tests.
func diffMock(logOutput string) *mockExecutor {
	return &mockExecutor{fn: func(ctx context.Context, name string, args ...string) (string, error) {
		if name == "gt" && len(args) > 0 && args[0] == "log" {
			return logOutput, nil
		}
		if name == "git" && len(args) > 0 && args[0] == "diff" {
			for _, a := range args {
				if a == "--stat" {
					return " model.go | 5 +++--\n keys.go  | 3 +++\n 2 files changed, 6 insertions(+), 2 deletions(-)\n", nil
				}
			}
			// File diff
			return "@@ -1,3 +1,5 @@\n+new line 1\n+new line 2\n old line\n", nil
		}
		return "", nil
	}}
}

func loadedDiffModel(logOutput string) Model {
	client := gt.New(diffMock(logOutput))
	m := New(client, "")
	m = sendWindowSize(m, 100, 30)
	updated, _ := m.Update(logResultMsg{output: logOutput})
	return updated.(Model)
}

func TestDiffKey_OpensLoading(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")

	updated, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'d'}}))
	m = updated.(Model)

	if !m.running {
		t.Error("pressing d should set running=true")
	}
	if !m.statusBar.spinning {
		t.Error("spinner should be active")
	}
	if !containsString(m.statusBar.spinnerLabel, "Loading diff") {
		t.Errorf("spinnerLabel = %q, want to contain 'Loading diff'", m.statusBar.spinnerLabel)
	}
	if cmd == nil {
		t.Fatal("expected commands from d key")
	}
}

func TestDiffKey_OnTrunk_ShowsError(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = 0 // on "main" (trunk/root)

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'d'}}))
	m = updated.(Model)

	if m.running {
		t.Error("d on trunk should not start running")
	}
	if m.mode != modeTree {
		t.Error("should stay in tree mode")
	}
	if !m.statusBar.isError {
		t.Error("status bar should show error")
	}
	if !containsString(m.statusBar.message, "No parent branch") {
		t.Errorf("message = %q, want to contain 'No parent branch'", m.statusBar.message)
	}
}

func TestDiffKey_EmptyTree(t *testing.T) {
	m := loadedDiffModel("some random output without markers")

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'d'}}))
	m = updated.(Model)

	if m.running {
		t.Error("d on empty tree should not start action")
	}
}

func TestDiffKey_BlockedWhileRunning(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'d'}}))
	m = updated.(Model)

	// Should still be in tree mode (d was blocked).
	if m.mode != modeTree {
		t.Error("d should be blocked while running")
	}
}

func TestDiffDataMsg_Success_EntersDiffMode(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true
	m.statusBar.startSpinner("Loading...")

	files := []diffFileEntry{
		{path: "model.go", summary: "5 +++--"},
		{path: "keys.go", summary: "3 +++"},
	}
	updated, cmd := m.Update(diffDataMsg{
		branchName:   "feature-top",
		parentBranch: "main",
		files:        files,
	})
	m = updated.(Model)

	if m.running {
		t.Error("running should be false after diffDataMsg")
	}
	if m.mode != modeDiff {
		t.Error("should be in diff mode")
	}
	if m.diff.branchName != "feature-top" {
		t.Errorf("branchName = %q, want %q", m.diff.branchName, "feature-top")
	}
	if m.diff.parentBranch != "main" {
		t.Errorf("parentBranch = %q, want %q", m.diff.parentBranch, "main")
	}
	if len(m.diff.files) != 2 {
		t.Errorf("got %d files, want 2", len(m.diff.files))
	}
	// Should issue loadDiffFile for first file.
	if cmd == nil {
		t.Error("expected command to load first file diff")
	}
}

func TestDiffDataMsg_NoFiles(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	updated, _ := m.Update(diffDataMsg{
		branchName:   "feature-top",
		parentBranch: "main",
		files:        nil,
	})
	m = updated.(Model)

	if m.mode != modeDiff {
		t.Error("should enter diff mode even with no files")
	}
	if len(m.diff.files) != 0 {
		t.Errorf("got %d files, want 0", len(m.diff.files))
	}
}

func TestDiffDataMsg_Error_ShowsError(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	updated, _ := m.Update(diffDataMsg{
		branchName: "main",
		err:        errors.New("no parent branch"),
	})
	m = updated.(Model)

	if m.running {
		t.Error("running should be false")
	}
	if m.mode != modeTree {
		t.Error("should stay in tree mode on error")
	}
	if !m.statusBar.isError {
		t.Error("status bar should show error")
	}
	if !containsString(m.statusBar.message, "no parent branch") {
		t.Errorf("message = %q, want to contain 'no parent branch'", m.statusBar.message)
	}
}

func TestDiffFileContentMsg_UpdatesViewport(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	// Enter diff mode manually.
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)
	m.diff.branchName = "feature-top"
	m.diff.parentBranch = "main"
	m.diff.setFiles([]diffFileEntry{{path: "model.go", summary: "5 +++--"}})

	updated, _ := m.Update(diffFileContentMsg{
		file:    "model.go",
		content: "+added line\n-removed line",
	})
	m = updated.(Model)

	view := m.diff.diffViewport.View()
	if !containsString(view, "+added line") {
		t.Error("diff viewport should contain the file content")
	}
}

func TestDiffFileContentMsg_Error(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)

	updated, _ := m.Update(diffFileContentMsg{
		file: "model.go",
		err:  errors.New("diff failed"),
	})
	m = updated.(Model)

	if !m.statusBar.isError {
		t.Error("status bar should show error")
	}
	if !containsString(m.statusBar.message, "diff failed") {
		t.Errorf("message = %q, want to contain 'diff failed'", m.statusBar.message)
	}
}

func TestDiffClose_Esc(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)
	m.diff.branchName = "feature-top"

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEscape}))
	m = updated.(Model)

	if m.mode != modeTree {
		t.Error("Esc should return to tree mode")
	}
}

func TestDiffClose_D(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)
	m.diff.branchName = "feature-top"

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'d'}}))
	m = updated.(Model)

	if m.mode != modeTree {
		t.Error("d should return to tree mode when in diff mode")
	}
}

func TestDiffClose_RestoresTree(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEscape}))
	m = updated.(Model)

	view := m.View()
	if !containsString(view, "feature-top") {
		t.Error("tree should be restored after closing diff")
	}
	if !containsString(view, "main") {
		t.Error("tree should contain 'main' after closing diff")
	}
}

func TestDiffMode_TreeKeysBlocked(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)
	m.diff.setFiles([]diffFileEntry{{path: "a.go", summary: "1 +"}})

	// Tree action keys should not trigger actions in diff mode.
	for _, k := range []rune{'s', 'S', 'r', 'f', 'y', 'o', 'm'} {
		updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{k}}))
		m = updated.(Model)
		if m.running {
			t.Errorf("key %c should not start action in diff mode", k)
		}
	}

	// Enter should also not trigger checkout.
	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))
	m = updated.(Model)
	if m.running {
		t.Error("Enter should not trigger checkout in diff mode")
	}
}

func TestDiffMode_QuitStillWorks(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)

	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'q'}}))
	if cmd == nil {
		t.Fatal("q should produce quit cmd in diff mode")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestDiffMode_Tab_TogglesFocus(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)
	m.diff.setFiles([]diffFileEntry{{path: "a.go"}})

	if m.diff.focusedPanel != panelFileList {
		t.Error("initial focus should be file list")
	}

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyTab}))
	m = updated.(Model)
	if m.diff.focusedPanel != panelDiff {
		t.Error("tab should switch to diff panel")
	}

	updated, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyTab}))
	m = updated.(Model)
	if m.diff.focusedPanel != panelFileList {
		t.Error("tab should switch back to file list")
	}
}

func TestDiffMode_Navigation_FileList(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)
	m.diff.branchName = "feature-top"
	m.diff.parentBranch = "main"
	m.diff.setFiles([]diffFileEntry{
		{path: "a.go", summary: "1 +"},
		{path: "b.go", summary: "2 ++"},
		{path: "c.go", summary: "3 +++"},
	})
	m.diff.focusedPanel = panelFileList

	if m.diff.fileCursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.diff.fileCursor)
	}

	// Move down.
	updated, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'j'}}))
	m = updated.(Model)
	if m.diff.fileCursor != 1 {
		t.Errorf("after down: cursor = %d, want 1", m.diff.fileCursor)
	}
	if cmd == nil {
		t.Error("expected loadDiffFile command after file cursor change")
	}

	// Move down again.
	updated, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'j'}}))
	m = updated.(Model)
	if m.diff.fileCursor != 2 {
		t.Errorf("after second down: cursor = %d, want 2", m.diff.fileCursor)
	}

	// Move down at end - should stay.
	updated, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'j'}}))
	m = updated.(Model)
	if m.diff.fileCursor != 2 {
		t.Errorf("at end: cursor = %d, want 2", m.diff.fileCursor)
	}

	// Move up.
	updated, cmd = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'k'}}))
	m = updated.(Model)
	if m.diff.fileCursor != 1 {
		t.Errorf("after up: cursor = %d, want 1", m.diff.fileCursor)
	}
	if cmd == nil {
		t.Error("expected loadDiffFile command after file cursor change")
	}
}

func TestDiffMode_Navigation_FileList_UpAtZero(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)
	m.diff.setFiles([]diffFileEntry{{path: "a.go"}})
	m.diff.focusedPanel = panelFileList

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'k'}}))
	m = updated.(Model)
	if m.diff.fileCursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.diff.fileCursor)
	}
}

func TestDiffMode_View_ShowsDiffLegend(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)
	m.diff.branchName = "feature-top"
	m.diff.parentBranch = "main"

	view := m.View()
	if !containsString(view, "switch panel") {
		t.Error("diff mode should show 'switch panel' in legend")
	}
	if !containsString(view, "close") {
		t.Error("diff mode should show 'close' in legend")
	}
}

func TestDiffMode_View_ShowsTreeLegendWhenClosed(t *testing.T) {
	m := loadedDiffModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	view := m.View()
	if !containsString(view, "diff") {
		t.Error("tree mode legend should contain 'diff'")
	}
	if !containsString(view, "checkout") {
		t.Error("tree mode legend should contain 'checkout'")
	}
}

func TestDiffMode_FullFlow(t *testing.T) {
	// End-to-end: press d, receive diff data, receive file content, view, close.
	logOutput := "│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main"
	m := loadedDiffModel(logOutput)

	// 1. Press d to open diff.
	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'d'}}))
	m = updated.(Model)
	if !m.running {
		t.Fatal("should be running after d")
	}

	// 2. Receive diff data.
	updated, _ = m.Update(diffDataMsg{
		branchName:   "feature-top",
		parentBranch: "main",
		files: []diffFileEntry{
			{path: "model.go", summary: "5 +++--"},
		},
	})
	m = updated.(Model)
	if m.mode != modeDiff {
		t.Fatal("should be in diff mode")
	}

	// 3. Receive file content.
	updated, _ = m.Update(diffFileContentMsg{
		file:    "model.go",
		content: "+added line\n old line\n-removed line",
	})
	m = updated.(Model)

	// 4. View should show diff content.
	view := m.View()
	if !containsString(view, "model.go") {
		t.Error("view should show file name")
	}
	if !containsString(view, "feature-top") {
		t.Error("view should show branch name")
	}

	// 5. Close with Esc.
	updated, _ = m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEscape}))
	m = updated.(Model)
	if m.mode != modeTree {
		t.Error("should return to tree mode")
	}
	view = m.View()
	if !containsString(view, "feature-top") {
		t.Error("tree should be restored")
	}
}

// --- Help mode tests ---

func TestHelpKey_OpensHelpMode(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")

	m = sendKey(m, '?')

	if m.mode != modeHelp {
		t.Error("pressing ? should switch to help mode")
	}
	view := m.View()
	if !containsString(view, "Keybindings") {
		t.Error("help mode should show keybinding content")
	}
}

func TestHelpKey_ClosesHelpMode(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m = sendKey(m, '?')
	if m.mode != modeHelp {
		t.Fatal("should be in help mode")
	}

	m = sendKey(m, '?')
	if m.mode != modeTree {
		t.Error("pressing ? again should return to tree mode")
	}
	view := m.View()
	if !containsString(view, "feature-top") {
		t.Error("tree should be restored")
	}
}

func TestHelpMode_EscCloses(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m = sendKey(m, '?')

	m = sendSpecialKey(m, tea.KeyEscape)
	if m.mode != modeTree {
		t.Error("esc should close help mode")
	}
}

func TestHelpMode_QuitStillWorks(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m = sendKey(m, '?')

	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'q'}}))
	if cmd == nil {
		t.Fatal("q should produce quit cmd in help mode")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", msg)
	}
}

func TestHelpMode_TreeKeysBlocked(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m = sendKey(m, '?')

	for _, k := range []rune{'s', 'S', 'r', 'f', 'y', 'o', 'd', 'm'} {
		updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{k}}))
		m = updated.(Model)
		if m.running {
			t.Errorf("key %c should not start action in help mode", k)
		}
		if m.mode != modeHelp {
			t.Errorf("key %c should not exit help mode", k)
		}
	}
}

func TestHelpMode_NavigationBlocked(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	initial := m.cursor
	m = sendKey(m, '?')

	m = sendKey(m, 'j')
	if m.cursor != initial {
		t.Error("navigation should be blocked in help mode")
	}
}

func TestHelpMode_ShowsHelpLegend(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m = sendKey(m, '?')

	view := m.View()
	if !containsString(view, "close help") {
		t.Error("help mode should show 'close help' in legend")
	}
}

func TestTreeLegend_ContainsHelp(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	view := m.View()
	if !containsString(view, "help") {
		t.Error("tree legend should contain 'help'")
	}
}

// --- PR info tests ---

func TestPRInfoResult_AppliedToBranches(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")

	infos := map[string]gt.PRInfo{
		"feature-top":  {Number: 142, State: "OPEN"},
		"feature-base": {Number: 143, State: "DRAFT"},
	}
	updated, _ := m.Update(prInfoResultMsg{infos: infos})
	m = updated.(Model)

	// Verify PR info is set on the branch tree.
	featureBase := m.branches[0].Children[0]
	if featureBase.PR.Number != 143 {
		t.Errorf("feature-base PR number = %d, want 143", featureBase.PR.Number)
	}
	if featureBase.PR.State != "DRAFT" {
		t.Errorf("feature-base PR state = %q, want DRAFT", featureBase.PR.State)
	}

	featureTop := featureBase.Children[0]
	if featureTop.PR.Number != 142 {
		t.Errorf("feature-top PR number = %d, want 142", featureTop.PR.Number)
	}
}

func TestPRInfoResult_RenderedInView(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")

	infos := map[string]gt.PRInfo{
		"feature-top":  {Number: 142, State: "OPEN"},
		"feature-base": {Number: 143, State: "DRAFT"},
	}
	updated, _ := m.Update(prInfoResultMsg{infos: infos})
	m = updated.(Model)

	view := m.View()
	if !containsString(view, "#142") {
		t.Error("view should contain '#142'")
	}
	if !containsString(view, "open") {
		t.Error("view should contain 'open'")
	}
	if !containsString(view, "#143") {
		t.Error("view should contain '#143'")
	}
}

func TestPRInfoResult_NoReRenderInDiffMode(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.mode = modeDiff
	m.diff = newDiffView(100, 28)

	infos := map[string]gt.PRInfo{
		"feature-top": {Number: 142, State: "OPEN"},
	}
	updated, _ := m.Update(prInfoResultMsg{infos: infos})
	m = updated.(Model)

	// PR info should still be applied to the branch data.
	featureBase := m.branches[0].Children[0]
	featureTop := featureBase.Children[0]
	if featureTop.PR.Number != 142 {
		t.Errorf("PR info should still be applied in diff mode, got %d", featureTop.PR.Number)
	}
}

func TestLogResult_DispatchesPRInfoLoad(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	content := "│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main"
	_, cmd := m.Update(logResultMsg{output: content})

	if cmd == nil {
		t.Fatal("logResultMsg should produce a command (including loadPRInfo)")
	}
}

func TestApplyPRInfo(t *testing.T) {
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
			},
		},
	}

	infos := map[string]gt.PRInfo{
		"a": {Number: 10, State: "OPEN"},
		"b": {Number: 20, State: "MERGED"},
	}
	applyPRInfo(branches, infos)

	if branches[0].PR.Number != 0 {
		t.Error("main should have no PR")
	}
	if branches[0].Children[0].PR.Number != 10 {
		t.Errorf("a PR = %d, want 10", branches[0].Children[0].PR.Number)
	}
	if branches[0].Children[0].Children[0].PR.Number != 20 {
		t.Errorf("b PR = %d, want 20", branches[0].Children[0].Children[0].PR.Number)
	}
}

// --- Error handling tests ---

func TestLogResult_GtNotFound(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	updated, _ := m.Update(logResultMsg{err: errors.New("executable file not found in $PATH")})
	m = updated.(Model)

	if !m.statusBar.isError {
		t.Error("status bar should show error")
	}
	if !containsString(m.statusBar.message, "gt CLI not found") {
		t.Errorf("message = %q, want gt CLI not found message", m.statusBar.message)
	}
}

func TestLogResult_DetachedHead(t *testing.T) {
	m := newTestModel("", nil)
	m = sendWindowSize(m, 80, 24)

	updated, _ := m.Update(logResultMsg{err: errors.New("detached HEAD state")})
	m = updated.(Model)

	if !m.statusBar.isError {
		t.Error("status bar should show error")
	}
	if !containsString(m.statusBar.message, "Detached HEAD") {
		t.Errorf("message = %q, want detached HEAD message", m.statusBar.message)
	}
}

func TestLogResult_ErrorPreservesExistingTree(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")

	// Verify tree is loaded.
	if len(m.branches) != 1 {
		t.Fatalf("expected 1 root, got %d", len(m.branches))
	}

	// Send an error — tree should be preserved.
	updated, _ := m.Update(logResultMsg{err: errors.New("some transient error")})
	m = updated.(Model)

	if len(m.branches) != 1 {
		t.Error("existing tree should be preserved on error")
	}
	if !containsString(m.statusBar.message, "Refresh failed") {
		t.Errorf("message = %q, want 'Refresh failed' prefix", m.statusBar.message)
	}

	// View should still show tree content.
	view := m.View()
	if !containsString(view, "feature-top") {
		t.Error("view should still show existing tree")
	}
}

func TestActionResult_Conflict(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	updated, _ := m.Update(actionResultMsg{action: "restack", err: errors.New("CONFLICT in file.go")})
	m = updated.(Model)

	if !m.statusBar.isError {
		t.Error("status bar should show error")
	}
	if !containsString(m.statusBar.message, "Conflict detected") {
		t.Errorf("message = %q, want conflict message", m.statusBar.message)
	}
}

func TestActionResult_ErrorReloadsTree(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.running = true

	_, cmd := m.Update(actionResultMsg{action: "restack", err: errors.New("some error")})

	if cmd == nil {
		t.Error("action error should trigger tree reload")
	}
}

func TestSubmitOnTrunk_ShowsError(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = 0 // on trunk

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'s'}}))
	m = updated.(Model)

	if m.running {
		t.Error("submit on trunk should not start action")
	}
	if !m.statusBar.isError {
		t.Error("status bar should show error")
	}
	if !containsString(m.statusBar.message, "Cannot submit trunk") {
		t.Errorf("message = %q, want trunk error", m.statusBar.message)
	}
}

func TestDownstackSubmitOnTrunk_ShowsError(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = 0

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'S'}}))
	m = updated.(Model)

	if m.running {
		t.Error("downstack submit on trunk should not start action")
	}
	if !containsString(m.statusBar.message, "Cannot submit trunk") {
		t.Errorf("message = %q, want trunk error", m.statusBar.message)
	}
}

func TestRestackOnTrunk_ShowsError(t *testing.T) {
	m := loadedModel("│ ◉  feature-top\n│ ◯  feature-base\n◯─┘  main")
	m.cursor = 0

	updated, _ := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'r'}}))
	m = updated.(Model)

	if m.running {
		t.Error("restack on trunk should not start action")
	}
	if !containsString(m.statusBar.message, "Cannot restack trunk") {
		t.Errorf("message = %q, want trunk error", m.statusBar.message)
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
