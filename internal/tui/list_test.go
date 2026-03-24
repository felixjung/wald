package tui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListWritesRenderedContent(t *testing.T) {
	t.Parallel()

	output := new(bytes.Buffer)
	theme := &Theme{}
	projects := []ListProject{
		{
			Name: "identity-infrastructure",
			Worktrees: []ListWorktree{
				{
					Label:   "main",
					Branch:  "main",
					Default: true,
				},
			},
		},
	}

	err := List(
		"Worktrees",
		"/tmp/worktrees",
		projects,
		WithOutput(output),
		WithTheme(theme),
	)
	require.NoError(t, err)

	require.Equal(t, renderListContent("Worktrees", "/tmp/worktrees", projects, theme), output.String())
	require.NotContains(t, output.String(), "\x1b[?")
}
