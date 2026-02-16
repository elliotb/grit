package gt

import (
	"context"
)

// DiffStat runs `git diff --stat <parent>...<branch>` and returns the raw output.
func (c *Client) DiffStat(ctx context.Context, parent, branch string) (string, error) {
	return c.executor.Execute(ctx, "git", "diff", "--stat", parent+"..."+branch)
}

// DiffFile runs `git diff --color=always <parent>...<branch> -- <file>` and
// returns the colored diff output for a single file.
func (c *Client) DiffFile(ctx context.Context, parent, branch, file string) (string, error) {
	return c.executor.Execute(ctx, "git", "diff", "--color=always", parent+"..."+branch, "--", file)
}
