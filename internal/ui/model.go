package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"

	"github.com/ejb/grit/internal/gt"
)

// logResultMsg is sent when `gt log short` completes.
type logResultMsg struct {
	output string
	err    error
}

// Model is the root bubbletea model for grit.
type Model struct {
	gtClient  *gt.Client
	viewport  viewport.Model
	statusBar statusBar
	keys      keyMap
	ready     bool
	branches  []*gt.Branch
	rawOutput string
	err       error
	width     int
	height    int
	gitDir    string
	watcher   *fsnotify.Watcher
}

// New creates a new root model. If gitDir is non-empty, a file watcher is
// created for auto-refresh on .git changes.
func New(gtClient *gt.Client, gitDir string) Model {
	m := Model{
		gtClient:  gtClient,
		gitDir:    gitDir,
		keys:      defaultKeyMap(),
		statusBar: newStatusBar(),
	}

	if gitDir != "" {
		watcher, err := createWatcher(gitDir)
		if err == nil {
			m.watcher = watcher
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadLog(), waitForChange(m.watcher))
}

func (m Model) loadLog() tea.Cmd {
	client := m.gtClient
	return func() tea.Msg {
		output, err := client.LogShort(context.Background())
		return logResultMsg{output: output, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			if m.watcher != nil {
				m.watcher.Close()
			}
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		statusBarHeight := 1
		viewportHeight := msg.Height - statusBarHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.viewport.SetContent(renderTree(m.branches))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}

		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.setSize(msg.Width)

	case logResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.rawOutput = msg.output
			m.statusBar.setMessage("Error: "+msg.err.Error(), true)
		} else {
			m.err = nil
			m.rawOutput = msg.output
			m.statusBar.setMessage("", false)
			m.statusBar.setRefreshTime(time.Now())
		}

		// Parse and render the tree, falling back to raw output on parse error.
		content := m.rawOutput
		if m.err == nil {
			branches, parseErr := gt.ParseLogShort(m.rawOutput)
			if parseErr == nil {
				m.branches = branches
				content = renderTree(branches)
			}
		}

		if m.ready {
			m.viewport.SetContent(content)
		}

	case gitChangeMsg:
		cmds = append(cmds, m.loadLog(), waitForChange(m.watcher))

	case watcherErrMsg:
		m.statusBar.setMessage("Watch error: "+msg.err.Error(), true)
		cmds = append(cmds, waitForChange(m.watcher))
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewport.View(),
		m.statusBar.view(),
	)
}
