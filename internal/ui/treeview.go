package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"

	"github.com/ejb/grit/internal/gt"
)

var (
	currentBranchStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	branchStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	treeEnumStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// renderTree converts parsed branches into a styled tree string using lipgloss/tree.
func renderTree(branches []*gt.Branch) string {
	if len(branches) == 0 {
		return "(no stacks)"
	}

	// Typically there's one root (the trunk). Render it directly.
	if len(branches) == 1 {
		return branchToTree(branches[0]).String()
	}

	// Multiple roots (unusual): render as children of a hidden-root tree.
	t := tree.New().
		EnumeratorStyle(treeEnumStyle)
	for _, b := range branches {
		t.Child(branchToTree(b))
	}
	return t.String()
}

// branchToTree recursively converts a Branch into a lipgloss/tree Tree node.
func branchToTree(b *gt.Branch) *tree.Tree {
	label := branchLabel(b)
	t := tree.Root(label).
		EnumeratorStyle(treeEnumStyle)

	for _, child := range b.Children {
		t.Child(branchToTree(child))
	}
	return t
}

// branchLabel returns a styled label for a branch.
func branchLabel(b *gt.Branch) string {
	if b.IsCurrent {
		return currentBranchStyle.Render("◉ " + b.Name)
	}
	return branchStyle.Render("◯ " + b.Name)
}
