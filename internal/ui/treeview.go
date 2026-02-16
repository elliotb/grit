package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ejb/grit/internal/gt"
)

var (
	currentBranchStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	branchStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	connectorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedBranchStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
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

// flattenForDisplay walks the branch tree and produces a flat list of entries
// with visual depth. Linear chains (single-child branches) stay at the same
// depth rather than increasing.
func flattenForDisplay(branches []*gt.Branch) []displayEntry {
	var entries []displayEntry
	for _, root := range branches {
		entries = append(entries, displayEntry{root, 0})
		collectStack(root.Children, 1, true, &entries)
	}
	return entries
}

// collectStack recursively collects branches into a flat display list.
// Linear chains (single child) keep the same depth. Standalone branches
// (leaf children directly under a root) are shown at depth 0.
func collectStack(children []*gt.Branch, depth int, rootLevel bool, entries *[]displayEntry) {
	for _, child := range children {
		if len(child.Children) == 0 && rootLevel {
			// Standalone leaf branch directly under root: show at depth 0
			*entries = append(*entries, displayEntry{child, 0})
		} else {
			*entries = append(*entries, displayEntry{child, depth})
			// Pass same depth for linear chains; no longer at root level
			collectStack(child.Children, depth, false, entries)
		}
	}
}

// branchLabel returns a styled label for a branch.
func branchLabel(b *gt.Branch) string {
	if b.IsCurrent {
		return currentBranchStyle.Render("◉ " + b.Name)
	}
	return branchStyle.Render("◯ " + b.Name)
}

// selectedBranchLabel returns a highlighted label for the cursor-selected branch.
func selectedBranchLabel(b *gt.Branch) string {
	marker := "◯ "
	if b.IsCurrent {
		marker = "◉ "
	}
	return selectedBranchStyle.Render(marker + b.Name)
}
