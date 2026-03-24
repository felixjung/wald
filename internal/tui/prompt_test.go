package tui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBubbleTeaEnvironmentSetsSSHTTYWhenMissing(t *testing.T) {
	t.Setenv("SSH_TTY", "")

	environ := bubbleTeaEnvironment()

	assert.Contains(t, environ, "SSH_TTY=wald")
}

func TestBubbleTeaEnvironmentOverridesSSHTTYWhenPresent(t *testing.T) {
	t.Setenv("SSH_TTY", "/dev/pts/1")

	environ := bubbleTeaEnvironment()

	assert.Contains(t, environ, "SSH_TTY=wald")
	for _, entry := range environ {
		if strings.HasPrefix(entry, "SSH_TTY=") {
			assert.Equal(t, "SSH_TTY=wald", entry)
		}
	}
}
