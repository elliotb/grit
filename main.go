package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ejb/grit/internal/gt"
	"github.com/ejb/grit/internal/ui"
)

func main() {
	gtClient := gt.NewDefault()
	model := ui.New(gtClient)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
