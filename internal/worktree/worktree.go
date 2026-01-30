package worktree

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner executes commands for worktree operations.
type Runner interface {
	Run(ctx context.Context, dir string, name string, args ...string) error
}

// OSRunner runs commands via os/exec.
type OSRunner struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (r OSRunner) Run(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if r.Stdout != nil {
		cmd.Stdout = r.Stdout
	}
	if r.Stderr != nil {
		cmd.Stderr = r.Stderr
	}
	return cmd.Run()
}

// Add creates a worktree for the given branch name, branching from main.
func Add(ctx context.Context, r Runner, repoRoot, branch string) error {
	if stringsTrim(branch) == "" {
		return errors.New("branch name is required")
	}
	if branch == "main" {
		return errors.New("branch name 'main' is reserved")
	}
	if err := validateRepoRoot(repoRoot); err != nil {
		return err
	}

	mainPath := filepath.Join(repoRoot, "main")
	if !pathExists(mainPath) {
		if err := r.Run(ctx, repoRoot, "git", "worktree", "add", "-B", "main", mainPath, "main"); err != nil {
			return fmt.Errorf("ensure main worktree: %w", err)
		}
	}

	branchPath := filepath.Join(repoRoot, branch)
	if pathExists(branchPath) {
		return fmt.Errorf("worktree path already exists: %s", branchPath)
	}
	if err := r.Run(ctx, repoRoot, "git", "worktree", "add", "-b", branch, branchPath, "main"); err != nil {
		return fmt.Errorf("add worktree: %w", err)
	}
	return nil
}

// Remove deletes a worktree for the given branch name.
func Remove(ctx context.Context, r Runner, repoRoot, branch string) error {
	if stringsTrim(branch) == "" {
		return errors.New("branch name is required")
	}
	if branch == "main" {
		return errors.New("cannot remove the main worktree")
	}
	if err := validateRepoRoot(repoRoot); err != nil {
		return err
	}

	branchPath := filepath.Join(repoRoot, branch)
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
