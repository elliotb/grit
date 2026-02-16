# grit — Product Description

## Context

Graphite's VS Code extension provides a convenient UI for managing stacked PRs, but there's no equivalent for terminal-centric workflows. **grit** fills this gap: a terminal UI that wraps the Graphite CLI (`gt`) to give full visibility and control over stacked PRs without leaving the terminal.

The name "grit" = **gr**aphite + g**it**.

---

## What it does

grit is a read-heavy, action-light TUI. Its primary job is to **show you the state of your stacks** at a glance, and let you **trigger common operations** with a few keystrokes. It delegates all data loading and mutations to the `gt` CLI.

### Main view: Stack tree

The default (and primary) screen shows all your stacks as an **indented tree** mirroring `gt log short`, with the current branch highlighted. Each branch shows:

- Branch name
- PR number and status (open / draft / merged / no PR)
- Sync indicator (needs sync, needs restack, up to date)

```
main
├── feature-auth
│   ├── feature-auth-middleware   #142 open ✓
│   └── feature-auth-tests       #143 draft ↻ needs restack
└── fix-pagination               #138 open ✓        ← you are here
```

The tree auto-refreshes when the `.git` directory changes (via filesystem watching), so it stays current as you commit, switch branches, rebase, or run `gt` commands in another terminal.

### Navigation

- **Arrow keys** — move through the tree
- **Enter** — check out the selected branch (`gt checkout`)
- **Tab** — switch focus between panels (when diff view is open)

### Actions (from the stack view)

| Key | Action |
|-----|--------|
| `s` | Submit current stack (`gt stack submit`) |
| `S` | Submit downstack (`gt downstack submit`) |
| `r` | Restack current stack (`gt stack restack`) |
| `y` | Sync repo (`gt repo sync`) |
| `o` | Open selected branch's PR in browser (`gt pr`) |
| `d` | Open diff view for selected branch |
| `?` | Show help / keybindings |
| `q` | Quit |

Actions that run `gt` commands show a spinner while executing and refresh the tree on completion. Errors are shown in a status bar at the bottom.

### Diff view

Pressing `d` on a branch opens a **split view**:

1. **Left panel**: list of changed files (from `git diff <parent>...<branch> --stat`)
2. **Right panel**: diff for the selected file (from `git diff --color=always <parent>...<branch> -- <file>`)

The diff is rendered by capturing `git diff --color=always` output and displaying it in a scrollable viewport — no custom diff renderer needed. Navigate files with arrow keys, scroll the diff with arrow keys when the right panel is focused, press `Esc` or `d` to return to the stack view.

---

## Data sources

grit reads data from two sources, both local:

1. **`gt` CLI commands** (with `--no-interactive`):
   - `gt log short` — stack tree structure and branch names
   - `gt branch pr-info` — PR number and status
   - `gt parent` / `gt children` — branch relationships
   - `gt submit`, `gt restack`, `gt sync` — mutations

2. **Git refs directly**:
   - `.git/refs/branch-metadata/<branch>` via `git cat-file -p` — parent branch, base commit (JSON)
   - `.git/HEAD` — current branch detection
   - `git diff` — file changes and diffs

Filesystem watching (fsnotify) on the `.git` directory triggers a refresh of the tree state.

---

## Tech stack

| Component | Choice | Why |
|-----------|--------|-----|
| Language | **Go** | bubbletea ecosystem, single binary distribution, good exec/process support |
| TUI framework | **bubbletea v1** (stable) | Elm architecture, battle-tested, active ecosystem |
| Styling | **lipgloss v1** | CSS-like terminal styling, tree rendering sub-package |
| Components | **bubbles v1** | viewport (diffs), help (keybindings), spinner (loading), key (bindings) |
| Tree rendering | **lipgloss/tree** | Purpose-built for hierarchical tree display |
| File watching | **fsnotify** | Cross-platform filesystem event notifications |
| CLI delegation | **os/exec** | Shell out to `gt` and `git` commands |

Requires **Go 1.24.0+** (bubbletea v1.3 requirement).

---

## Architecture (high-level)

```
┌─────────────────────────────────────┐
│           Root Model                │
│  ┌─────────┐  ┌──────────────────┐  │
│  │  Stack   │  │   Diff View      │  │
│  │  Tree    │  │  ┌────┐ ┌─────┐ │  │
│  │  Panel   │  │  │File│ │Diff │ │  │
│  │         │  │  │List│ │View │ │  │
│  │         │  │  │    │ │port │ │  │
│  └─────────┘  │  └────┘ └─────┘ │  │
│               └──────────────────┘  │
│  ┌─────────────────────────────────┐│
│  │         Status Bar              ││
│  └─────────────────────────────────┘│
└─────────────────────────────────────┘
```

- **Root model**: manages focus state, routes input, composes child views
- **Stack tree panel**: renders `lipgloss/tree`, handles branch selection and action keybindings
- **Diff view**: two sub-panels (file list + viewport), shown/hidden as an overlay or replacement for the tree
- **Status bar**: shows current operation, errors, and last refresh time
- **Data layer**: module that shells out to `gt`/`git`, parses output, and provides structured data to the UI

---

## What it does NOT do

- **No GitHub API calls** — all PR data comes via `gt` which handles auth
- **No git operations directly** — all mutations go through `gt`
- **No config file** (initially) — sensible defaults, zero setup
- **No mouse support** (initially) — keyboard only
- **No commit creation or editing** — use your normal git/gt workflow for that
- **No multi-repo support** — launched from a repo root, operates on that repo

---

## Development setup (first steps)

1. **Upgrade Go** to 1.24+ (current: 1.18.1)
2. **Initialize git repo** in project directory
3. **Initialize Go module** (`go mod init github.com/ejb/grit` or similar)
4. **Install dependencies**: bubbletea, lipgloss, bubbles, fsnotify
5. **Scaffold**: `main.go` with a minimal bubbletea program that runs `gt log short` and displays the output

---

## Incremental build plan

### Phase 1: Skeleton
- Minimal bubbletea app that runs `gt log short` and displays raw output in a viewport
- Can quit with `q`

### Phase 2: Parsed tree
- Parse `gt log short` output into a data structure
- Render using `lipgloss/tree` with current branch highlighted
- Auto-refresh via fsnotify on `.git` directory

### Phase 3: Navigation + actions
- Arrow key navigation through branches
- Enter to checkout
- `o` to open PR in browser
- `s` / `S` / `r` / `y` for submit/restack/sync
- Spinner during operations, status bar for results/errors

### Phase 4: Diff view
- `d` to open diff view for selected branch
- Left panel: changed file list from `git diff --stat`
- Right panel: scrollable colored diff in viewport
- Tab to switch focus, Esc to close

### Phase 5: Polish
- PR status indicators (parse `gt branch pr-info`)
- Sync/restack status indicators
- Help screen (`?`)
- Error handling and edge cases (no stacks, detached HEAD, conflicts)
- Graceful handling of `gt` command failures
