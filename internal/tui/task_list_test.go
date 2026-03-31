package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func TestTaskLineKeepsLabelAlignmentAcrossStates(t *testing.T) {
	theme := &Theme{}
	task := Task{ID: "repo", ParentID: "clone", Label: "repo"}

	pending := renderTaskLine(task, 1, taskStatus{State: TaskStatePending}, "⠋", theme)
	running := renderTaskLine(task, 1, taskStatus{State: TaskStateRunning}, "⠋", theme)
	done := renderTaskLine(task, 1, taskStatus{State: TaskStateDone}, "⠋", theme)
	failed := renderTaskLine(task, 1, taskStatus{State: TaskStateFailed}, "⠋", theme)

	expected := lipgloss.Width(strings.SplitN(pending, "repo", 2)[0])
	require.Equal(t, expected, lipgloss.Width(strings.SplitN(running, "repo", 2)[0]))
	require.Equal(t, expected, lipgloss.Width(strings.SplitN(done, "repo", 2)[0]))
	require.Equal(t, expected, lipgloss.Width(strings.SplitN(failed, "repo", 2)[0]))
}

func TestRenderTaskListContentRendersNestedRepositoryTasks(t *testing.T) {
	theme := &Theme{}
	tasks := []Task{
		{ID: "clone", Label: "Clone default branches"},
		{ID: "repo-one", ParentID: "clone", Label: "repo-one"},
		{ID: "repo-two", ParentID: "clone", Label: "repo-two"},
	}
	statuses := map[string]taskStatus{
		"clone":    {State: TaskStateRunning},
		"repo-one": {State: TaskStateDone},
		"repo-two": {State: TaskStatePending},
	}

	content := renderTaskListContent("Initializing worktrees", tasks, statuses, "⠋", theme)

	require.Contains(t, content, "Initializing worktrees")
	require.Contains(t, content, "[⠋] Clone default branches")
	require.Contains(t, content, "\n  [✓] repo-one")
	require.Contains(t, content, "\n  [ ] repo-two")
}

func TestTaskSymbolsUseThemeStyles(t *testing.T) {
	theme := &Theme{
		Help:          lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		PromptFocused: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		LabelFocused:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		Error:         lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
	}

	pending := renderTaskSymbol(TaskStatePending, "⠋", theme)
	running := renderTaskSymbol(TaskStateRunning, "⠋", theme)
	done := renderTaskSymbol(TaskStateDone, "⠋", theme)
	failed := renderTaskSymbol(TaskStateFailed, "⠋", theme)

	require.Contains(t, pending, "\x1b[")
	require.Contains(t, pending, "[ ]")
	require.Contains(t, running, "\x1b[")
	require.Contains(t, running, "[⠋]")
	require.Contains(t, done, "\x1b[")
	require.Contains(t, done, "[✓]")
	require.Contains(t, failed, "\x1b[")
	require.Contains(t, failed, "[✗]")
}
