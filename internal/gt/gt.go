package gt

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
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
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return string(out), fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
	}
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

// Checkout runs `gt checkout <name> --no-interactive`.
func (c *Client) Checkout(ctx context.Context, branchName string) error {
	_, err := c.executor.Execute(ctx, "gt", "checkout", branchName, "--no-interactive")
	return err
}

// StackSubmit runs `gt stack submit --no-interactive --branch <branchName>`.
func (c *Client) StackSubmit(ctx context.Context, branchName string) error {
	_, err := c.executor.Execute(ctx, "gt", "stack", "submit", "--no-interactive", "--branch", branchName)
	return err
}

// DownstackSubmit runs `gt downstack submit --no-interactive --branch <branchName>`.
func (c *Client) DownstackSubmit(ctx context.Context, branchName string) error {
	_, err := c.executor.Execute(ctx, "gt", "downstack", "submit", "--no-interactive", "--branch", branchName)
	return err
}

// StackRestack runs `gt stack restack --no-interactive --branch <branchName>`.
func (c *Client) StackRestack(ctx context.Context, branchName string) error {
	_, err := c.executor.Execute(ctx, "gt", "stack", "restack", "--no-interactive", "--branch", branchName)
	return err
}

// RepoSync runs `gt repo sync --no-interactive`.
func (c *Client) RepoSync(ctx context.Context) error {
	_, err := c.executor.Execute(ctx, "gt", "repo", "sync", "--no-interactive")
	return err
}

// Sync runs `gt sync -f --no-interactive`.
func (c *Client) Sync(ctx context.Context) error {
	_, err := c.executor.Execute(ctx, "gt", "sync", "-f", "--no-interactive")
	return err
}

// OpenPR runs `gt pr <branchName>` to open the branch's PR in the browser.
func (c *Client) OpenPR(ctx context.Context, branchName string) error {
	_, err := c.executor.Execute(ctx, "gt", "pr", branchName)
	return err
}

// BranchPRInfo runs `gt branch pr-info --branch <branchName> --no-interactive`
// and returns the raw JSON output.
func (c *Client) BranchPRInfo(ctx context.Context, branchName string) (string, error) {
	return c.executor.Execute(ctx, "gt", "branch", "pr-info", "--branch", branchName, "--no-interactive")
}
