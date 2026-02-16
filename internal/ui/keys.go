package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit            key.Binding
	Up              key.Binding
	Down            key.Binding
	Checkout        key.Binding
	Trunk           key.Binding
	StackSubmit     key.Binding
	DownstackSubmit key.Binding
	Restack         key.Binding
	Fetch           key.Binding
	Sync            key.Binding
	OpenPR          key.Binding
	Diff            key.Binding
	DiffClose       key.Binding
	Tab             key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Checkout: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "checkout"),
		),
		Trunk: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "trunk"),
		),
		StackSubmit: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "submit stack"),
		),
		DownstackSubmit: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "submit downstack"),
		),
		Restack: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restack"),
		),
		Fetch: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "fetch"),
		),
		Sync: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "sync"),
		),
		OpenPR: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open PR"),
		),
		Diff: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "diff"),
		),
		DiffClose: key.NewBinding(
			key.WithKeys("d", "esc"),
			key.WithHelp("esc/d", "close"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
	}
}
