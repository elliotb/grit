package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	content   string
	err       error
	width     int
	height    int
}

// New creates a new root model.
func New(gtClient *gt.Client) Model {
	return Model{
		gtClient:  gtClient,
		keys:      defaultKeyMap(),
		statusBar: newStatusBar(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadLog()
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
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		statusBarHeight := 1
		viewportHeight := msg.Height - statusBarHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.viewport.SetContent(m.content)
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
			m.content = msg.output
			m.statusBar.setMessage("Error: "+msg.err.Error(), true)
		} else {
			m.err = nil
			m.content = msg.output
			m.statusBar.setMessage("", false)
			m.statusBar.setRefreshTime(time.Now())
		}
		if m.ready {
			m.viewport.SetContent(m.content)
		}
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
