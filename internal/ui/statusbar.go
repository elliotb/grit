package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type statusBar struct {
	width       int
	message     string
	isError     bool
	lastRefresh time.Time
}

func newStatusBar() statusBar {
	return statusBar{}
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

func (s statusBar) view() string {
	style := lipgloss.NewStyle().
		Width(s.width).
		Padding(0, 1)

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
