package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Runner executes commands for hook operations.
type Runner interface {
	Run(ctx context.Context, dir, name string, args ...string) error
}

// RunAll executes hook commands in order and fails fast on the first error.
func RunAll(ctx context.Context, r Runner, dir, hookName string, commands []string) error {
	if r == nil {
		return errors.New("runner is required")
	}
	if strings.TrimSpace(dir) == "" {
		return errors.New("hook directory is required")
	}
	for i, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			return fmt.Errorf("%s hook command %d is required", hookName, i+1)
		}
		if err := r.Run(ctx, dir, "sh", "-c", command); err != nil {
			return fmt.Errorf("%s hook command %d failed: %w", hookName, i+1, err)
		}
	}
	return nil
}
