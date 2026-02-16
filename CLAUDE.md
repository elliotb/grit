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

No Makefile, linter, or CI pipeline is configured yet.

## Architecture

**grit** is a terminal UI that wraps the Graphite CLI (`gt`) to manage stacked PRs. It uses the bubbletea (Elm architecture) TUI framework. All git/graphite mutations delegate to the `gt` CLI via shell exec — grit never calls the GitHub API or runs git operations directly.

### Package structure

- **`main.go`** — Entry point. Creates a `gt.Client`, passes it to `ui.New()`, runs the bubbletea program with alt-screen.
- **`internal/gt/`** — Graphite CLI wrapper. `CommandExecutor` interface abstracts shell execution for testability. `Client` provides typed methods (currently `LogShort`) that shell out to `gt` with `--no-interactive`.
- **`internal/ui/`** — Bubbletea UI layer. `Model` is the root model managing a viewport, status bar, and keybindings. Messages flow through `Update()` in the standard bubbletea pattern.

### Key patterns

- **Dependency injection for testing**: `gt.Client` accepts a `CommandExecutor` interface. Tests use a mock executor with canned output instead of shelling out to real `gt`.
- **Async commands via bubbletea messages**: Operations like `loadLog()` return `tea.Cmd` functions that produce typed messages (`logResultMsg`). The `Update` loop handles these messages to update state.
- **Status bar**: Bottom row showing errors or last-refresh time, sized to terminal width.

### Development roadmap

See `PRODUCT.md` for the full specification and 5-phase build plan. Phase 1 (skeleton with `gt log short` display) is complete. Next phases add parsed tree rendering, navigation, actions, diff view, and polish.

## Development Workflow

- **Write comprehensive tests**: All new code should have thorough test coverage. Run `go test ./...` as part of verification before considering a task done.
- **Commit frequently**: Commit after each individual task completes, not at the end of an entire development stage. This keeps changes atomic and reviewable.
