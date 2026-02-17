# grit

A terminal UI for managing [Graphite](https://graphite.dev/) stacked PRs. The name is **gr**aphite + g**it**.

> **Note:** I vibe coded this for my own use. It works for my workflow but comes with no guarantees, no support, and probably some rough edges. Use at your own risk.

## What it does

grit gives you a read-heavy, action-light TUI for your Graphite stacks. It shows your branches as an indented tree, lets you navigate and check out branches, trigger common operations, and inspect diffs — all without leaving the terminal.

```
main
├─ ◯ feature-auth
│  ├─ ◯ feature-auth-middleware   #142 open
│  └─ ◯ feature-auth-tests       #143 draft
└─ ◉ fix-pagination               #138 open        ← you are here
```

The tree auto-refreshes when your `.git` directory changes, so it stays current as you work in another terminal.

### Views

- **Stack tree** (default) — your branches as a tree with PR status labels
- **Diff view** — split panel with file list + scrollable colored diff
- **Help screen** — keybinding reference

## Keybindings

### Stack tree

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `enter` | Check out selected branch |
| `m` | Check out trunk (main/master) |
| `d` | Open diff view |
| `s` | Submit stack |
| `S` | Submit downstack |
| `r` | Restack stack |
| `f` | Fetch (repo sync) |
| `y` | Sync |
| `o` | Open PR in browser |
| `?` | Toggle help |
| `q` | Quit |

### Diff view

| Key | Action |
|-----|--------|
| `j` / `↓` | Next file / scroll down |
| `k` / `↑` | Previous file / scroll up |
| `tab` | Switch focus between file list and diff |
| `d` / `esc` | Close diff view |

## Requirements

- **Go 1.25.0+**
- **[Graphite CLI](https://graphite.dev/docs/graphite-cli)** (`gt`) installed and authenticated

## Installation

```bash
go install github.com/elliotb/grit@latest
```

Or build from source:

```bash
git clone https://github.com/elliotb/grit.git
cd grit
go build
./grit
```

## How it works

grit delegates everything to the `gt` CLI — it never calls the GitHub API or runs git mutations directly. It parses output from commands like `gt log short` and `gt branch pr-info` to build the tree, and shells out to `gt` for actions like submit, restack, and sync.

The UI is built with [bubbletea](https://github.com/charmbracelet/bubbletea) (Elm architecture). File watching via [fsnotify](https://github.com/fsnotify/fsnotify) triggers debounced tree reloads so the display stays in sync with your repo state.
