package ui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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

// actionResultMsg is sent when an async gt action completes.
type actionResultMsg struct {
	action  string
	err     error
	message string // success message to display
}

// debounceFireMsg is sent after a debounce delay to trigger a reload.
// The seq field must match Model.debounceSeq to fire; stale ticks are ignored.
type debounceFireMsg struct{ seq int }

// debounceDuration is the delay before reloading after a filesystem event.
const debounceDuration = 300 * time.Millisecond

// Model is the root bubbletea model for grit.
type Model struct {
	gtClient       *gt.Client
	viewport       viewport.Model
	statusBar      statusBar
	keys           keyMap
	ready          bool
	branches       []*gt.Branch
	displayEntries []displayEntry
	cursor         int
	rawOutput      string
	err            error
	width          int
	height         int
	gitDir         string
	watcher        *fsnotify.Watcher
	debounceSeq    int
	running        bool
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

// selectedBranch returns the branch at the current cursor position, or nil.
func (m Model) selectedBranch() *gt.Branch {
	if m.cursor >= 0 && m.cursor < len(m.displayEntries) {
		return m.displayEntries[m.cursor].branch
	}
	return nil
}

// preserveCursor tries to keep the cursor on the same branch after a tree
// reload. It searches by name first, falls back to the IsCurrent branch,
// then falls back to index 0.
func (m *Model) preserveCursor(oldBranchName string) {
	if oldBranchName != "" {
		for i, e := range m.displayEntries {
			if e.branch.Name == oldBranchName {
				m.cursor = i
				return
			}
		}
	}
	for i, e := range m.displayEntries {
		if e.branch.IsCurrent {
			m.cursor = i
			return
		}
	}
	m.cursor = 0
}

// ensureCursorVisible adjusts the viewport scroll so the cursor line is visible.
func (m *Model) ensureCursorVisible() {
	if m.cursor < m.viewport.YOffset {
		m.viewport.SetYOffset(m.cursor)
	} else if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.SetYOffset(m.cursor - m.viewport.Height + 1)
	}
}

// runAction returns a tea.Cmd that runs fn asynchronously and produces an
// actionResultMsg when it completes.
func runAction(action, successMsg string, fn func(ctx context.Context) error) tea.Cmd {
	return func() tea.Msg {
		err := fn(context.Background())
		return actionResultMsg{action: action, err: err, message: successMsg}
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

		// Block all other input while an action is running.
		if m.running {
			break
		}

		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.viewport.SetContent(renderTree(m.displayEntries, m.cursor))
				m.ensureCursorVisible()
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.displayEntries)-1 {
				m.cursor++
				m.viewport.SetContent(renderTree(m.displayEntries, m.cursor))
				m.ensureCursorVisible()
			}
		case key.Matches(msg, m.keys.Checkout):
			if branch := m.selectedBranch(); branch != nil {
				m.running = true
				name := branch.Name
				client := m.gtClient
				spinnerCmd := m.statusBar.startSpinner("Checking out " + name + "...")
				actionCmd := runAction("checkout", "Checked out "+name, func(ctx context.Context) error {
					return client.Checkout(ctx, name)
				})
				cmds = append(cmds, spinnerCmd, actionCmd)
			}
		case key.Matches(msg, m.keys.Trunk):
			if len(m.branches) > 0 {
				m.running = true
				name := m.branches[0].Name
				client := m.gtClient
				spinnerCmd := m.statusBar.startSpinner("Checking out " + name + "...")
				actionCmd := runAction("checkout", "Checked out "+name, func(ctx context.Context) error {
					return client.Checkout(ctx, name)
				})
				cmds = append(cmds, spinnerCmd, actionCmd)
			}
		case key.Matches(msg, m.keys.StackSubmit):
			if branch := m.selectedBranch(); branch != nil {
				m.running = true
				name := branch.Name
				client := m.gtClient
				spinnerCmd := m.statusBar.startSpinner("Submitting stack (" + name + ")...")
				actionCmd := runAction("submit", "Stack submitted", func(ctx context.Context) error {
					return client.StackSubmit(ctx, name)
				})
				cmds = append(cmds, spinnerCmd, actionCmd)
			}
		case key.Matches(msg, m.keys.DownstackSubmit):
			if branch := m.selectedBranch(); branch != nil {
				m.running = true
				name := branch.Name
				client := m.gtClient
				spinnerCmd := m.statusBar.startSpinner("Submitting downstack (" + name + ")...")
				actionCmd := runAction("downstack-submit", "Downstack submitted", func(ctx context.Context) error {
					return client.DownstackSubmit(ctx, name)
				})
				cmds = append(cmds, spinnerCmd, actionCmd)
			}
		case key.Matches(msg, m.keys.Restack):
			if branch := m.selectedBranch(); branch != nil {
				m.running = true
				name := branch.Name
				client := m.gtClient
				spinnerCmd := m.statusBar.startSpinner("Restacking (" + name + ")...")
				actionCmd := runAction("restack", "Restacked", func(ctx context.Context) error {
					return client.StackRestack(ctx, name)
				})
				cmds = append(cmds, spinnerCmd, actionCmd)
			}
		case key.Matches(msg, m.keys.Fetch):
			m.running = true
			client := m.gtClient
			spinnerCmd := m.statusBar.startSpinner("Fetching...")
			actionCmd := runAction("fetch", "Fetched", func(ctx context.Context) error {
				return client.RepoSync(ctx)
			})
			cmds = append(cmds, spinnerCmd, actionCmd)
		case key.Matches(msg, m.keys.Sync):
			m.running = true
			client := m.gtClient
			spinnerCmd := m.statusBar.startSpinner("Syncing...")
			actionCmd := runAction("sync", "Synced", func(ctx context.Context) error {
				return client.Sync(ctx)
			})
			cmds = append(cmds, spinnerCmd, actionCmd)
		case key.Matches(msg, m.keys.OpenPR):
			if branch := m.selectedBranch(); branch != nil {
				m.running = true
				name := branch.Name
				client := m.gtClient
				spinnerCmd := m.statusBar.startSpinner("Opening PR (" + name + ")...")
				actionCmd := runAction("openpr", "Opened PR for "+name, func(ctx context.Context) error {
					return client.OpenPR(ctx, name)
				})
				cmds = append(cmds, spinnerCmd, actionCmd)
			}
		}

	case tea.WindowSizeMsg:
		chromeHeight := 2 // legend + status bar
		viewportHeight := msg.Height - chromeHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.viewport.KeyMap = viewport.KeyMap{}
			m.viewport.SetContent(renderTree(m.displayEntries, m.cursor))
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
				oldName := ""
				if b := m.selectedBranch(); b != nil {
					oldName = b.Name
				}
				m.displayEntries = flattenForDisplay(branches)
				m.preserveCursor(oldName)
				content = renderTree(m.displayEntries, m.cursor)
			}
		}

		if m.ready {
			m.viewport.SetContent(content)
		}

	case actionResultMsg:
		m.running = false
		m.statusBar.stopSpinner()
		if msg.err != nil {
			m.statusBar.setMessage("Error: "+msg.err.Error(), true)
		} else {
			m.statusBar.setSuccessMessage(msg.message)
			// Reload tree after successful actions (except openpr which doesn't change git state).
			if msg.action != "openpr" {
				cmds = append(cmds, m.loadLog())
			}
		}

	case spinner.TickMsg:
		if m.running {
			var cmd tea.Cmd
			m.statusBar.spinner, cmd = m.statusBar.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case gitChangeMsg:
		m.debounceSeq++
		seq := m.debounceSeq
		cmds = append(cmds,
			tea.Tick(debounceDuration, func(time.Time) tea.Msg {
				return debounceFireMsg{seq: seq}
			}),
			waitForChange(m.watcher),
		)

	case debounceFireMsg:
		if msg.seq == m.debounceSeq {
			cmds = append(cmds, m.loadLog())
		}

	case watcherErrMsg:
		m.statusBar.setMessage("Watch error: "+msg.err.Error(), true)
		cmds = append(cmds, waitForChange(m.watcher))
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

var (
	legendKeyStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	legendDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func (m Model) legendView() string {
	pairs := []struct{ key, desc string }{
		{"↑↓", "navigate"},
		{"enter", "checkout"},
		{"m", "trunk"},
		{"s", "submit"},
		{"S", "downstack"},
		{"r", "restack"},
		{"f", "fetch"},
		{"y", "sync"},
		{"o", "open PR"},
		{"q", "quit"},
	}

	var sb strings.Builder
	for i, p := range pairs {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(legendKeyStyle.Render(p.key))
		sb.WriteString(" ")
		sb.WriteString(legendDescStyle.Render(p.desc))
	}

	style := lipgloss.NewStyle().Width(m.width).Padding(0, 1)
	return style.Render(sb.String())
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewport.View(),
		m.legendView(),
		m.statusBar.view(),
	)
}
