package gh

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/felixjung/trees/internal/runner"
)

const binaryName = "gh"

// Runner matches the subset of runner.Runner we need.
type Runner interface {
	Run(ctx context.Context, dir, name string, args ...string) error
}

var (
	ErrNotAvailable     = errors.New("gh CLI not found in PATH")
	ErrNotAuthenticated = errors.New("gh is not authenticated")
)

// GitHubCLI wraps interactions with the gh binary.
type GitHubCLI struct {
	Runner Runner
}

// New constructs a GitHubCLI instance.
func New(ctx context.Context, r Runner) (*GitHubCLI, error) {
	if r == nil {
		return nil, errors.New("runner is required")
	}
	client := &GitHubCLI{Runner: r}
	if err := client.EnsureAvailable(); err != nil {
		return nil, err
	}
	if err := client.EnsureAuthenticated(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

// EnsureAvailable checks that gh is available in PATH.
func (c *GitHubCLI) EnsureAvailable() error {
	if _, err := exec.LookPath(binaryName); err != nil {
		return ErrNotAvailable
	}
	return nil
}

// EnsureAuthenticated verifies gh authentication for github.com.
func (c *GitHubCLI) EnsureAuthenticated(ctx context.Context) error {
	if err := c.Runner.Run(
		ctx,
		"",
		binaryName,
		"auth",
		"status",
		"-h",
		"github.com",
	); err != nil {
		return fmt.Errorf("%w; run 'gh auth login'", ErrNotAuthenticated)
	}
	return nil
}

// Clone clones a repo into the destination directory.
func (c *GitHubCLI) Clone(ctx context.Context, repo, dest string) error {
	if err := c.Runner.Run(ctx, "", binaryName, "repo", "clone", repo, dest); err != nil {
		return c.HandleError(err, repo)
	}
	return nil
}

// HandleError maps known gh errors to user-friendly messages.
func (c *GitHubCLI) HandleError(err error, repo string) error {
	switch {
	case err == nil:
		return nil
	case c.isPermissionError(err):
		return handlePermissionError(repo)
	default:
		return handleUnknownError(err)
	}
}

func handlePermissionError(repo string) error {
	return fmt.Errorf("no permission to clone %s; ensure you have access", repo)
}

func handleUnknownError(err error) error {
	return err
}

// isPermissionError reports whether an error from gh likely indicates access issues.
func (c *GitHubCLI) isPermissionError(err error) bool {
	var cmdErr *runner.CommandError
	if errors.As(err, &cmdErr) {
		combined := strings.ToLower(cmdErr.Stdout + "\n" + cmdErr.Stderr)
		return strings.Contains(combined, "permission") ||
			strings.Contains(combined, "access denied") ||
			strings.Contains(combined, "not authorized") ||
			strings.Contains(combined, "forbidden") ||
			strings.Contains(combined, "not found")
	}
	return false
}
