package worktree

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Runner executes commands for worktree operations.
type Runner interface {
	Run(ctx context.Context, dir, name string, args ...string) error
}

// Add creates a worktree at the given path, invoked from the default branch worktree.
// extraArgs are forwarded to `git worktree add`.
func Add(ctx context.Context, r Runner, gitDir, worktreePath, startPoint string, extraArgs []string) (string, error) {
	if strings.TrimSpace(worktreePath) == "" {
		return "", errors.New("worktree path is required")
	}
	if err := validateRepoRoot(gitDir); err != nil {
		return "", err
	}
	opts, err := splitExtraArgs(extraArgs)
	if err != nil {
		return "", err
	}
	args := append([]string{"worktree", "add"}, opts...)
	args = append(args, worktreePath)
	if strings.TrimSpace(startPoint) != "" {
		args = append(args, startPoint)
	}
	if err := r.Run(ctx, gitDir, "git", args...); err != nil {
		return "", fmt.Errorf("add worktree: %w", err)
	}
	return worktreePath, nil
}

// Remove deletes a worktree at the given path.
// extraArgs are forwarded to `git worktree remove`.
func Remove(ctx context.Context, r Runner, gitDir, worktreePath string, extraArgs []string) (string, error) {
	if strings.TrimSpace(worktreePath) == "" {
		return "", errors.New("worktree path is required")
	}
	if err := validateRepoRoot(gitDir); err != nil {
		return "", err
	}
	opts, err := splitExtraArgs(extraArgs)
	if err != nil {
		return "", err
	}
	args := append([]string{"worktree", "remove"}, opts...)
	args = append(args, worktreePath)
	if err := r.Run(ctx, gitDir, "git", args...); err != nil {
		return "", fmt.Errorf("remove worktree: %w", err)
	}
	return worktreePath, nil
}

func validateRepoRoot(repoRoot string) error {
	if strings.TrimSpace(repoRoot) == "" {
		return errors.New("repo root is required")
	}
	info, err := os.Stat(repoRoot)
	if err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("repo root is not a directory: %s", repoRoot)
	}
	gitPath := filepath.Join(repoRoot, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		return fmt.Errorf("repo root is not a git repository: %s", repoRoot)
	}
	return nil
}

func splitExtraArgs(extra []string) ([]string, error) {
	if len(extra) == 0 {
		return nil, nil
	}
	if extra[0] != "--" {
		return nil, errors.New("extra args must start with --")
	}
	return extra[1:], nil
}
