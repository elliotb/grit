package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	helpTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	helpKeyStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")).Width(12)
	helpDescStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	helpSectionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type helpEntry struct {
	key  string
	desc string
}

func renderHelp() string {
	sections := []struct {
		header  string
		entries []helpEntry
	}{
		{
			header: "Navigation",
			entries: []helpEntry{
				{"^/k", "Move cursor up"},
				{"v/j", "Move cursor down"},
				{"enter", "Check out selected branch"},
				{"m", "Check out trunk (main/master)"},
			},
		},
		{
			header: "Actions",
			entries: []helpEntry{
				{"s", "Submit stack"},
				{"S", "Submit downstack"},
				{"r", "Restack stack"},
				{"f", "Fetch (repo sync)"},
				{"y", "Sync"},
				{"o", "Open PR in browser"},
			},
		},
		{
			header: "Views",
			entries: []helpEntry{
				{"d", "Open diff view for selected branch"},
				{"?", "Toggle this help screen"},
				{"q", "Quit"},
			},
		},
		{
			header: "Diff View",
			entries: []helpEntry{
				{"^v", "Navigate files / scroll diff"},
				{"tab", "Switch panel focus"},
				{"esc/d", "Close diff view"},
			},
		},
	}

	var sb strings.Builder
	sb.WriteString(helpTitleStyle.Render("grit - Keybindings"))
	sb.WriteString("\n\n")

	for i, section := range sections {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(helpSectionStyle.Render("--- " + section.header + " ---"))
		sb.WriteString("\n")
		for _, e := range section.entries {
			sb.WriteString(helpKeyStyle.Render(e.key))
			sb.WriteString(helpDescStyle.Render(e.desc))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(helpSectionStyle.Render("Press ? or esc to close"))

	return sb.String()
}
