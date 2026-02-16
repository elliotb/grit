package ui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// gitChangeMsg is sent when the file watcher detects a change in .git.
type gitChangeMsg struct{}

// watcherErrMsg is sent when the file watcher encounters an error.
type watcherErrMsg struct{ err error }

// createWatcher creates an fsnotify watcher on specific .git subdirectories
// that change during git/graphite operations. It deliberately does NOT watch
// the .git directory itself to avoid a process storm: running `gt` modifies
// transient files in .git (lock files, index), which would re-trigger the
// watcher in an infinite loop.
func createWatcher(gitDir string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch HEAD file for branch switches.
	headPath := filepath.Join(gitDir, "HEAD")
	if _, err := os.Stat(headPath); err == nil {
		if err := watcher.Add(headPath); err != nil {
			watcher.Close()
			return nil, err
		}
	}

	// Watch subdirectories that change during branch operations.
	// fsnotify doesn't recurse, so we add each explicitly.
	// Errors are non-fatal — the directory may not exist yet.
	subdirs := []string{
		"refs/heads",
		filepath.Join("refs", "branch-metadata"),
	}
	for _, sub := range subdirs {
		path := filepath.Join(gitDir, sub)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			_ = watcher.Add(path)
		}
	}

	if len(watcher.WatchList()) == 0 {
		watcher.Close()
		return nil, fmt.Errorf("no watchable paths found in %s", gitDir)
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
