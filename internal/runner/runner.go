package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// CommandError wraps an execution error with captured output.
type CommandError struct {
	Name   string
	Args   []string
	Err    error
	Stdout string
	Stderr string
}

func (e *CommandError) Error() string {
	output := strings.TrimSpace(e.Stdout + "\n" + e.Stderr)
	if output == "" {
		return fmt.Sprintf("%s %s: %v", e.Name, strings.Join(e.Args, " "), e.Err)
	}
	return fmt.Sprintf("%s %s: %v\n%s", e.Name, strings.Join(e.Args, " "), e.Err, output)
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

// OSRunner runs commands via os/exec.
type OSRunner struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (r OSRunner) Run(ctx context.Context, dir, name string, args ...string) error {
	_, err := r.RunOutput(ctx, dir, name, args...)
	return err
}

func (r OSRunner) RunOutput(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	if r.Stdout != nil {
		cmd.Stdout = io.MultiWriter(r.Stdout, &stdoutBuf)
	} else {
		cmd.Stdout = &stdoutBuf
	}
	if r.Stderr != nil {
		cmd.Stderr = io.MultiWriter(r.Stderr, &stderrBuf)
	} else {
		cmd.Stderr = &stderrBuf
	}

	if err := cmd.Run(); err != nil {
		return "", &CommandError{
			Name:   name,
			Args:   args,
			Err:    err,
			Stdout: stdoutBuf.String(),
			Stderr: stderrBuf.String(),
		}
	}
	return stdoutBuf.String(), nil
}
