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

// Add creates a worktree for the given branch path. The branch name is derived from the path base.
func Add(ctx context.Context, r Runner, gitDir, worktreePath string) (string, error) {
	branchName := strings.TrimSpace(filepath.Base(worktreePath))
	if branchName == "" {
		return "", errors.New("branch name is required")
	}
	if err := validateRepoRoot(gitDir); err != nil {
		return "", err
	}
	if pathExists(worktreePath) {
		return "", fmt.Errorf("worktree path already exists: %s", worktreePath)
	}
	if err := r.Run(ctx, gitDir, "git", "worktree", "add", "-b", branchName, worktreePath); err != nil {
		return "", fmt.Errorf("add worktree: %w", err)
	}
	return worktreePath, nil
}

// Remove deletes a worktree at the given path.
func Remove(ctx context.Context, r Runner, gitDir, worktreePath string) (string, error) {
	if strings.TrimSpace(worktreePath) == "" {
		return "", errors.New("worktree path is required")
	}
	if err := validateRepoRoot(gitDir); err != nil {
		return "", err
	}
	if !pathExists(worktreePath) {
		return "", fmt.Errorf("worktree path does not exist: %s", worktreePath)
	}
	if err := r.Run(ctx, gitDir, "git", "worktree", "remove", worktreePath); err != nil {
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

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
