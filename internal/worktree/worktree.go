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

// Add creates a worktree for the given branch name, branching from baseBranch.
// If branch is omitted, baseBranch is treated as the branch name and "main" is used as the base branch.
func Add(ctx context.Context, r Runner, repoRoot, baseBranch string, branch ...string) error {
	branchName := stringsTrim(baseBranch)
	if len(branch) > 0 {
		branchName = stringsTrim(branch[0])
	}
	if branchName == "" {
		return errors.New("branch name is required")
	}
	if len(branch) == 0 {
		baseBranch = "main"
	}
	baseBranch = stringsTrim(baseBranch)
	if baseBranch == "" {
		return errors.New("default branch is required")
	}
	if branchName == baseBranch {
		return fmt.Errorf("branch name %q is reserved", baseBranch)
	}
	if err := validateRepoRoot(repoRoot); err != nil {
		return err
	}

	basePath := filepath.Join(repoRoot, baseBranch)
	if !pathExists(basePath) {
		if err := r.Run(ctx, repoRoot, "git", "worktree", "add", "-B", baseBranch, basePath, baseBranch); err != nil {
			return fmt.Errorf("ensure default branch worktree: %w", err)
		}
	}

	branchPath := filepath.Join(repoRoot, branchName)
	if pathExists(branchPath) {
		return fmt.Errorf("worktree path already exists: %s", branchPath)
	}
	if err := r.Run(ctx, repoRoot, "git", "worktree", "add", "-b", branchName, branchPath, baseBranch); err != nil {
		return fmt.Errorf("add worktree: %w", err)
	}
	return nil
}

// Remove deletes a worktree for the given branch name.
// If branch is omitted, baseBranch is treated as the branch name and "main" is used as the base branch.
func Remove(ctx context.Context, r Runner, repoRoot, baseBranch string, branch ...string) error {
	branchName := stringsTrim(baseBranch)
	if len(branch) > 0 {
		branchName = stringsTrim(branch[0])
	}
	if branchName == "" {
		return errors.New("branch name is required")
	}
	if len(branch) == 0 {
		baseBranch = "main"
	}
	baseBranch = stringsTrim(baseBranch)
	if baseBranch == "" {
		return errors.New("default branch is required")
	}
	if branchName == baseBranch {
		return fmt.Errorf("cannot remove the default branch worktree %q", baseBranch)
	}
	if err := validateRepoRoot(repoRoot); err != nil {
		return err
	}

	branchPath := filepath.Join(repoRoot, branchName)
	if !pathExists(branchPath) {
		return fmt.Errorf("worktree path does not exist: %s", branchPath)
	}
	if err := r.Run(ctx, repoRoot, "git", "worktree", "remove", branchPath); err != nil {
		return fmt.Errorf("remove worktree: %w", err)
	}
	return nil
}

// WorkdirPath resolves the working directory for a project within a worktree.
func WorkdirPath(repoRoot, branch, workdir string) string {
	if workdir == "" {
		return filepath.Join(repoRoot, branch)
	}
	return filepath.Join(repoRoot, branch, filepath.FromSlash(workdir))
}

func validateRepoRoot(repoRoot string) error {
	if stringsTrim(repoRoot) == "" {
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

func stringsTrim(value string) string {
	return strings.TrimSpace(value)
}
