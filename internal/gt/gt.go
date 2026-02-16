package gt

import (
	"context"
	"os/exec"
)

// CommandExecutor abstracts the execution of shell commands.
// In production, this calls os/exec. In tests, it returns canned output.
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) (string, error)
}

// ExecCommandExecutor is the real implementation that shells out via os/exec.
type ExecCommandExecutor struct{}

func (e *ExecCommandExecutor) Execute(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	return string(out), err
}

// Client provides methods for running gt CLI commands.
type Client struct {
	executor CommandExecutor
}

// New creates a new Client with the given CommandExecutor.
func New(executor CommandExecutor) *Client {
	return &Client{executor: executor}
}

// NewDefault creates a new Client that executes real shell commands.
func NewDefault() *Client {
	return &Client{executor: &ExecCommandExecutor{}}
}

// LogShort runs `gt log short --no-interactive` and returns the raw output.
func (c *Client) LogShort(ctx context.Context) (string, error) {
	return c.executor.Execute(ctx, "gt", "log", "short", "--no-interactive")
}
