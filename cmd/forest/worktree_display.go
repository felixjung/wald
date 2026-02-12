package main

import (
	"path/filepath"
	"strings"
)

func displayWorktreeLabel(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}

func isDefaultWorktree(root, defaultBranch, path string) bool {
	if defaultBranch == "" {
		return false
	}
	defaultPath := filepath.Join(root, defaultBranch)
	return filepath.Clean(defaultPath) == filepath.Clean(path)
}
