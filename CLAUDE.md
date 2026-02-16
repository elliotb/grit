# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
go build                    # Build binary → ./grit
go test ./...               # Run all tests
go test -v ./...            # Verbose test output
go test ./internal/gt/      # Run tests for a single package
go test -run TestLogShort ./internal/gt/  # Run a single test
```

No Makefile, linter, or CI pipeline is configured yet. Requires Go 1.25.0+.

## Architecture

**grit** is a terminal UI that wraps the Graphite CLI (`gt`) to manage stacked PRs. It uses the bubbletea (Elm architecture) TUI framework. All git/graphite mutations delegate to the `gt` CLI via shell exec — grit never calls the GitHub API or runs git operations directly.

### Package structure

- **`main.go`** — Entry point. Creates a `gt.Client`, passes it to `ui.New()`, runs the bubbletea program with alt-screen.
- **`internal/gt/`** — Graphite CLI wrapper.
  - `gt.go` — `CommandExecutor` interface for testability + `Client` with typed methods for all gt/git commands (`LogShort`, `Checkout`, `StackSubmit`, `DownstackSubmit`, `StackRestack`, `RepoSync`, `Sync`, `OpenPR`, `BranchPRInfo`).
  - `parse.go` — Parses `gt log short` output into a `[]*Branch` tree. `FindParent` walks the tree to find a branch's parent.
  - `diff.go` — `DiffStat` and `DiffFile` methods that shell out to `git diff`.
  - `prinfo.go` — `ParsePRInfo` parses JSON from `gt branch pr-info` into `PRInfo` structs.
- **`internal/ui/`** — Bubbletea UI layer.
  - `model.go` — Root model: owns viewport, status bar, view mode, branches, cursor. `Update()` handles all message types. `loadLog()`, `loadPRInfo()`, `loadDiffData()`, `loadDiffFile()` return `tea.Cmd` for async operations.
  - `treeview.go` — `flattenForDisplay` converts branch tree → flat `[]displayEntry` list. `renderTree` renders with `│` connectors, cursor highlighting, PR labels, and annotations.
  - `diffview.go` — Split-panel diff view (file list + scrollable diff viewport). Also contains `parseDiffStat`.
  - `helpview.go` — Full-screen keybinding reference.
  - `keys.go` — `keyMap` struct with all keybindings.
  - `statusbar.go` — Bottom status bar with spinner, errors, and last-refresh time.
  - `watcher.go` — fsnotify file watcher on `.git/HEAD` and `refs/` subdirs, with debounced reload.

### Key patterns

- **Dependency injection for testing**: `gt.Client` accepts a `CommandExecutor` interface. Tests use a mock executor with canned output instead of shelling out to real `gt`/`git`.
- **Async commands via bubbletea messages**: Operations return `tea.Cmd` functions that produce typed messages (`logResultMsg`, `actionResultMsg`, `diffDataMsg`, `diffFileContentMsg`, `prInfoResultMsg`). The `Update` loop handles these messages to update state.
- **View modes**: The model has three modes — `modeTree` (default stack view), `modeDiff` (split-panel diff), `modeHelp` (keybinding reference). Key handling is mode-specific.
- **Debounced file watching**: `.git` changes trigger `gitChangeMsg` → increment `debounceSeq` → `tea.Tick` after 300ms → `debounceFireMsg` fires reload only if seq matches (stale ticks are ignored).
- **Cursor preservation**: After tree reloads, `preserveCursor` tries to keep the cursor on the same branch by name, falling back to the current branch, then index 0.
- **Action locking**: `m.running` flag blocks all input during async `gt` commands, preventing concurrent mutations.

### Development roadmap

See `PRODUCT.md` for the full specification and 5-phase build plan. Phases 1–5 are substantially complete: skeleton, parsed tree with auto-refresh, navigation + actions, diff view, and polish (PR status, help screen, error handling).

## Development Workflow

- **Write comprehensive tests**: All new code should have thorough test coverage. Run `go test ./...` as part of verification before considering a task done.
- **Commit frequently**: Commit after each individual task completes, not at the end of an entire development stage. This keeps changes atomic and reviewable.
