package worktree

import (
	"context"
	"fmt"
	"strings"
)

// OutputRunner executes commands and returns stdout output.
type OutputRunner interface {
	RunOutput(ctx context.Context, dir, name string, args ...string) (string, error)
}

// Entry describes a git worktree.
type Entry struct {
	Path     string
	Branch   string
	Detached bool
}

// List returns all worktrees for the repository rooted at gitDir.
func List(ctx context.Context, r OutputRunner, gitDir string) ([]Entry, error) {
	if err := validateRepoRoot(gitDir); err != nil {
		return nil, err
	}
	output, err := r.RunOutput(ctx, gitDir, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list worktrees: %w", err)
	}
	return parseWorktreeList(output), nil
}

func parseWorktreeList(output string) []Entry {
	lines := strings.Split(output, "\n")
	entries := make([]Entry, 0)
	current := Entry{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current.Path != "" {
				entries = append(entries, current)
			}
			current = Entry{Path: strings.TrimSpace(strings.TrimPrefix(line, "worktree "))}
		case strings.HasPrefix(line, "branch "):
			branch := strings.TrimSpace(strings.TrimPrefix(line, "branch "))
			branch = strings.TrimPrefix(branch, "refs/heads/")
			current.Branch = branch
		case line == "detached":
			current.Detached = true
		}
	}
	if current.Path != "" {
		entries = append(entries, current)
	}
	return entries
}
