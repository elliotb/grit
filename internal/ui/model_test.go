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

	// View should contain tree connector characters from lipgloss/tree
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
