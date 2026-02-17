package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/elliotb/grit/internal/gt"
)

var (
	currentBranchStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	branchStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	connectorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedBranchStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
	annotationStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	prOpenStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	prDraftStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	prMergedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	prClosedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// displayEntry represents a branch with its visual depth for flat rendering.
type displayEntry struct {
	branch *gt.Branch
	depth  int
}

// renderTree converts display entries into a styled flat display with │ connectors.
// The entry at the cursor index is highlighted with reverse video.
func renderTree(entries []displayEntry, cursor int) string {
	if len(entries) == 0 {
		return "(no stacks)"
	}

	var sb strings.Builder
	for i, e := range entries {
		if i > 0 {
			sb.WriteString("\n")
		}
		if e.depth > 0 {
			sb.WriteString(connectorStyle.Render(strings.Repeat("│ ", e.depth)))
		}
		if i == cursor {
			sb.WriteString(selectedBranchLabel(e.branch))
		} else {
			sb.WriteString(branchLabel(e.branch))
		}
	}
	return sb.String()
}

// flattenForDisplay collects all branches from the tree and sorts them by
// their original gt log short line order (top-of-stack first, trunk last).
func flattenForDisplay(branches []*gt.Branch) []displayEntry {
	var entries []displayEntry
	collectAll(branches, &entries)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].branch.Order < entries[j].branch.Order
	})
	return entries
}

// collectAll recursively collects all branches from the tree into a flat list.
func collectAll(branches []*gt.Branch, entries *[]displayEntry) {
	for _, b := range branches {
		*entries = append(*entries, displayEntry{b, b.Depth})
		collectAll(b.Children, entries)
	}
}

// annotationLabel returns a styled annotation suffix, or empty string if none.
func annotationLabel(b *gt.Branch) string {
	if b.Annotation == "" {
		return ""
	}
	return " " + annotationStyle.Render("("+b.Annotation+")")
}

// prLabel returns a styled PR status label, or empty string if no PR.
func prLabel(pr gt.PRInfo) string {
	if pr.Number == 0 {
		return ""
	}
	numStr := fmt.Sprintf("#%d", pr.Number)
	switch strings.ToUpper(pr.State) {
	case "OPEN":
		return " " + prOpenStyle.Render(numStr+" open")
	case "DRAFT":
		return " " + prDraftStyle.Render(numStr+" draft")
	case "MERGED":
		return " " + prMergedStyle.Render(numStr+" merged")
	case "CLOSED":
		return " " + prClosedStyle.Render(numStr+" closed")
	default:
		return " " + numStr
	}
}

// prLabelPlain returns an unstyled PR status string for use in reverse-video labels.
func prLabelPlain(pr gt.PRInfo) string {
	if pr.Number == 0 {
		return ""
	}
	numStr := fmt.Sprintf("#%d", pr.Number)
	switch strings.ToUpper(pr.State) {
	case "OPEN":
		return " " + numStr + " open"
	case "DRAFT":
		return " " + numStr + " draft"
	case "MERGED":
		return " " + numStr + " merged"
	case "CLOSED":
		return " " + numStr + " closed"
	default:
		return " " + numStr
	}
}

// branchLabel returns a styled label for a branch.
func branchLabel(b *gt.Branch) string {
	if b.IsCurrent {
		return currentBranchStyle.Render("◉ "+b.Name) + annotationLabel(b) + prLabel(b.PR)
	}
	return branchStyle.Render("◯ "+b.Name) + annotationLabel(b) + prLabel(b.PR)
}

// selectedBranchLabel returns a highlighted label for the cursor-selected branch.
func selectedBranchLabel(b *gt.Branch) string {
	marker := "◯ "
	if b.IsCurrent {
		marker = "◉ "
	}
	label := marker + b.Name
	if b.Annotation != "" {
		label += " (" + b.Annotation + ")"
	}
	label += prLabelPlain(b.PR)
	return selectedBranchStyle.Render(label)
}
