package ui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ejb/grit/internal/gt"
)

func TestUpdate_GitChangeMsg_TriggersReload(t *testing.T) {
	client := gt.New(&mockExecutor{output: "◉  main", err: nil})
	m := New(client, "")
	m = sendWindowSize(m, 80, 24)

	updated, cmd := m.Update(gitChangeMsg{})
	_ = updated.(Model)

	if cmd == nil {
		t.Fatal("gitChangeMsg should return a command (loadLog + waitForChange)")
	}
}

func TestUpdate_WatcherErrMsg_ShowsError(t *testing.T) {
	dir := t.TempDir()
	client := gt.New(&mockExecutor{output: "◉  main", err: nil})
	m := New(client, dir)
	defer m.watcher.Close()
	m = sendWindowSize(m, 80, 24)

	updated, cmd := m.Update(watcherErrMsg{err: errors.New("watch failed")})
	m = updated.(Model)

	view := m.View()
	if !containsString(view, "Watch error:") {
		t.Errorf("view should contain watch error, got:\n%s", view)
	}

	// Should still return a command to re-arm the watcher listener
	if cmd == nil {
		t.Fatal("watcherErrMsg should return a command to re-arm watcher")
	}
}

func TestCreateWatcher_InvalidDir(t *testing.T) {
	_, err := createWatcher("/nonexistent/path/.git")
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

func TestCreateWatcher_ValidDir(t *testing.T) {
	// Use a temp directory to test watcher creation
	dir := t.TempDir()
	watcher, err := createWatcher(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer watcher.Close()

	// Verify the directory is being watched
	watchList := watcher.WatchList()
	if len(watchList) == 0 {
		t.Fatal("expected at least one watched path")
	}
}

func TestWaitForChange_NilWatcher(t *testing.T) {
	cmd := waitForChange(nil)
	if cmd != nil {
		t.Error("waitForChange(nil) should return nil")
	}
}

func TestNew_WithGitDir(t *testing.T) {
	dir := t.TempDir()
	client := gt.New(&mockExecutor{output: "", err: nil})
	m := New(client, dir)

	if m.watcher == nil {
		t.Fatal("expected watcher to be created for valid gitDir")
	}
	m.watcher.Close()
}

func TestNew_WithEmptyGitDir(t *testing.T) {
	client := gt.New(&mockExecutor{output: "", err: nil})
	m := New(client, "")

	if m.watcher != nil {
		t.Error("expected no watcher for empty gitDir")
	}
}

func TestQuit_ClosesWatcher(t *testing.T) {
	dir := t.TempDir()
	client := gt.New(&mockExecutor{output: "", err: nil})
	m := New(client, dir)

	if m.watcher == nil {
		t.Fatal("expected watcher to be created")
	}

	// Quit should close the watcher
	_, cmd := m.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'q'}}))
	if cmd == nil {
		t.Fatal("expected quit command")
	}

	// Verify watcher is closed by checking WatchList returns empty
	watchList := m.watcher.WatchList()
	if len(watchList) != 0 {
		t.Error("expected watcher to be closed (empty watch list)")
	}
}

// mockExecutor is already defined in model_test.go, so we use it directly.
// This file relies on the test helpers from model_test.go since they're in
// the same package.
var _ gt.CommandExecutor = (*mockExecutor)(nil)

// Verify the mock satisfies the interface with the right context parameter.
func init() {
	m := &mockExecutor{output: "test"}
	_, _ = m.Execute(context.Background(), "test")
}
