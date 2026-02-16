package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type statusBar struct {
	width        int
	message      string
	isError      bool
	lastRefresh  time.Time
	spinner      spinner.Model
	spinning     bool
	spinnerLabel string
}

func newStatusBar() statusBar {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	return statusBar{spinner: s}
}

func (s *statusBar) setSize(width int) {
	s.width = width
}

func (s *statusBar) setMessage(msg string, isError bool) {
	s.message = msg
	s.isError = isError
}

func (s *statusBar) setRefreshTime(t time.Time) {
	s.lastRefresh = t
}

// startSpinner begins the spinner animation with the given label.
// Returns a tea.Cmd that must be sent to start the spinner ticks.
func (s *statusBar) startSpinner(label string) tea.Cmd {
	s.spinning = true
	s.spinnerLabel = label
	return s.spinner.Tick
}

// stopSpinner ends the spinner animation.
func (s *statusBar) stopSpinner() {
	s.spinning = false
	s.spinnerLabel = ""
}

func (s statusBar) view() string {
	style := lipgloss.NewStyle().
		Width(s.width).
		Padding(0, 1)

	if s.spinning {
		style = style.Foreground(lipgloss.Color("6"))
		return style.Render(s.spinner.View() + " " + s.spinnerLabel)
	}

	if s.isError {
		style = style.
			Foreground(lipgloss.Color("1")).
			Bold(true)
	} else {
		style = style.
			Foreground(lipgloss.Color("8"))
	}

	text := s.message
	if text == "" && !s.lastRefresh.IsZero() {
		text = fmt.Sprintf("Last refreshed: %s", s.lastRefresh.Format("15:04:05"))
	}
	if text == "" {
		text = "grit"
	}

	return style.Render(text)
}
