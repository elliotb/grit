package gt

import (
	"context"
	"strings"
)

// Parent runs `gt parent --branch <branchName> --no-interactive` and returns
// the parent branch name.
func (c *Client) Parent(ctx context.Context, branchName string) (string, error) {
	output, err := c.executor.Execute(ctx, "gt", "parent", "--branch", branchName, "--no-interactive")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// DiffStat runs `git diff --stat <parent>...<branch>` and returns the raw output.
func (c *Client) DiffStat(ctx context.Context, parent, branch string) (string, error) {
	return c.executor.Execute(ctx, "git", "diff", "--stat", parent+"..."+branch)
}

// DiffFile runs `git diff --color=always <parent>...<branch> -- <file>` and
// returns the colored diff output for a single file.
func (c *Client) DiffFile(ctx context.Context, parent, branch, file string) (string, error) {
	return c.executor.Execute(ctx, "git", "diff", "--color=always", parent+"..."+branch, "--", file)
}
