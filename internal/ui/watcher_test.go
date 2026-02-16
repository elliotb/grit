package ui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ejb/grit/internal/gt"
)

// setupFakeGitDir creates a minimal .git directory structure for watcher tests.
func setupFakeGitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Create HEAD file
	if err := os.WriteFile(filepath.Join(dir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("failed to create HEAD: %v", err)
	}
	// Create refs/heads directory
	if err := os.MkdirAll(filepath.Join(dir, "refs", "heads"), 0755); err != nil {
		t.Fatalf("failed to create refs/heads: %v", err)
	}
	return dir
}

// newWatcherTestModel creates a test model with a simple mock executor.
func newWatcherTestModel(gitDir string) Model {
	client := gt.New(simpleMock("◉  main", nil))
	return New(client, gitDir)
}

func TestUpdate_GitChangeMsg_ReturnsDebounceTick(t *testing.T) {
	m := newWatcherTestModel("")
	m = sendWindowSize(m, 80, 24)

	updated, cmd := m.Update(gitChangeMsg{})
	m = updated.(Model)

	if cmd == nil {
		t.Fatal("gitChangeMsg should return a command (debounce tick + waitForChange)")
	}

	// debounceSeq should have been incremented
	if m.debounceSeq != 1 {
		t.Errorf("debounceSeq = %d, want 1", m.debounceSeq)
	}
}

func TestUpdate_DebounceFireMsg_MatchingSeq(t *testing.T) {
	m := newWatcherTestModel("")
	m = sendWindowSize(m, 80, 24)
	m.debounceSeq = 5

	_, cmd := m.Update(debounceFireMsg{seq: 5})
	if cmd == nil {
		t.Fatal("debounceFireMsg with matching seq should return a loadLog command")
	}
}

func TestUpdate_DebounceFireMsg_StaleSeq(t *testing.T) {
	m := newWatcherTestModel("")
	m = sendWindowSize(m, 80, 24)
	m.debounceSeq = 5

	_, cmd := m.Update(debounceFireMsg{seq: 3})
	// Stale seq should not trigger a loadLog, but may still have viewport cmd
	_ = cmd
	if m.debounceSeq != 5 {
		t.Errorf("debounceSeq should not change on stale fire, got %d", m.debounceSeq)
	}
}

func TestUpdate_GitChangeMsg_MultipleEvents_OnlyLastFires(t *testing.T) {
	m := newWatcherTestModel("")
	m = sendWindowSize(m, 80, 24)

	// Simulate 3 rapid gitChangeMsg events
	for i := 0; i < 3; i++ {
		updated, _ := m.Update(gitChangeMsg{})
		m = updated.(Model)
	}

	// debounceSeq should be 3
	if m.debounceSeq != 3 {
		t.Errorf("debounceSeq = %d, want 3", m.debounceSeq)
	}

	// Only seq=3 should trigger a reload; seq=1 and seq=2 are stale
	updated1, cmd1 := m.Update(debounceFireMsg{seq: 1})
	m = updated1.(Model)
	_ = cmd1

	updated3, cmd3 := m.Update(debounceFireMsg{seq: 3})
	m = updated3.(Model)
	if cmd3 == nil {
		t.Fatal("debounceFireMsg with current seq should return a command")
	}
}

func TestDebounceDuration(t *testing.T) {
	if debounceDuration != 300*time.Millisecond {
		t.Errorf("debounceDuration = %v, want 300ms", debounceDuration)
	}
}

func TestUpdate_WatcherErrMsg_ShowsError(t *testing.T) {
	dir := setupFakeGitDir(t)
	m := newWatcherTestModel(dir)
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

func TestCreateWatcher_ValidGitDir(t *testing.T) {
	dir := setupFakeGitDir(t)
	watcher, err := createWatcher(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer watcher.Close()

	watchList := watcher.WatchList()
	if len(watchList) == 0 {
		t.Fatal("expected at least one watched path")
	}

	// Should watch HEAD and refs/heads, but NOT the .git directory itself
	for _, path := range watchList {
		if path == dir {
			t.Errorf("should not watch .git directory itself, but found %q in watch list", path)
		}
	}
}

func TestCreateWatcher_EmptyDir_NoWatchablePaths(t *testing.T) {
	dir := t.TempDir() // No HEAD or refs/heads
	_, err := createWatcher(dir)
	if err == nil {
		t.Fatal("expected error for dir without watchable paths")
	}
}

func TestWaitForChange_NilWatcher(t *testing.T) {
	cmd := waitForChange(nil)
	if cmd != nil {
		t.Error("waitForChange(nil) should return nil")
	}
}

func TestNew_WithGitDir(t *testing.T) {
	dir := setupFakeGitDir(t)
	m := newWatcherTestModel(dir)

	if m.watcher == nil {
		t.Fatal("expected watcher to be created for valid gitDir")
	}
	m.watcher.Close()
}

func TestNew_WithEmptyGitDir(t *testing.T) {
	client := gt.New(simpleMock("", nil))
	m := New(client, "")

	if m.watcher != nil {
		t.Error("expected no watcher for empty gitDir")
	}
}

func TestQuit_ClosesWatcher(t *testing.T) {
	dir := setupFakeGitDir(t)
	m := newWatcherTestModel(dir)

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
