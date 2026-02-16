package ui

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// gitChangeMsg is sent when the file watcher detects a change in .git.
type gitChangeMsg struct{}

// watcherErrMsg is sent when the file watcher encounters an error.
type watcherErrMsg struct{ err error }

// createWatcher creates an fsnotify watcher on the .git directory and
// key subdirectories that change during git/graphite operations.
func createWatcher(gitDir string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch the .git directory itself (catches HEAD changes, index updates)
	if err := watcher.Add(gitDir); err != nil {
		watcher.Close()
		return nil, err
	}

	// Watch subdirectories that change during branch operations.
	// fsnotify doesn't recurse, so we add each explicitly.
	// Errors are non-fatal — the directory may not exist yet.
	subdirs := []string{
		"refs",
		"refs/heads",
		filepath.Join("refs", "branch-metadata"),
	}
	for _, sub := range subdirs {
		path := filepath.Join(gitDir, sub)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			_ = watcher.Add(path)
		}
	}

	return watcher, nil
}

// waitForChange returns a tea.Cmd that blocks until the watcher fires
// an event or error, then sends the appropriate message.
func waitForChange(watcher *fsnotify.Watcher) tea.Cmd {
	if watcher == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			return gitChangeMsg{}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return watcherErrMsg{err: err}
		}
	}
}
