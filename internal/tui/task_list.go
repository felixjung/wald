package tui

import (
	"errors"
	"os"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const taskListIndent = "  "

var taskListSpinnerFrames = spinner.Spinner{
	Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	FPS:    spinner.Line.FPS,
}

// Task describes one row in a nested task list.
type Task struct {
	ID       string
	ParentID string
	Label    string
}

// TaskState describes the current state of a task row.
type TaskState string

const (
	// TaskStatePending indicates work has not started yet.
	TaskStatePending TaskState = "pending"
	// TaskStateRunning indicates work is in progress.
	TaskStateRunning TaskState = "running"
	// TaskStateDone indicates work completed successfully.
	TaskStateDone TaskState = "done"
	// TaskStateFailed indicates work failed.
	TaskStateFailed TaskState = "failed"
)

// TaskUpdate applies a new state to one row in the task list.
type TaskUpdate struct {
	ID     string
	State  TaskState
	Detail string
}

type taskStatus struct {
	State  TaskState
	Detail string
}

type taskListUpdateMsg struct {
	update TaskUpdate
	ok     bool
}

type taskListModel struct {
	title        string
	tasks        []Task
	statuses     map[string]taskStatus
	spinner      spinner.Model
	updates      <-chan TaskUpdate
	theme        *Theme
	themeProfile *ThemeProfile
	isDarkBG     bool
}

func newTaskListModel(
	title string,
	tasks []Task,
	updates <-chan TaskUpdate,
	theme *Theme,
	themeProfile *ThemeProfile,
	isDarkBG bool,
) *taskListModel {
	statuses := make(map[string]taskStatus, len(tasks))
	for _, task := range tasks {
		statuses[task.ID] = taskStatus{State: TaskStatePending}
	}

	return &taskListModel{
		title:        title,
		tasks:        append([]Task(nil), tasks...),
		statuses:     statuses,
		spinner:      spinner.New(spinner.WithSpinner(taskListSpinnerFrames), spinner.WithStyle(theme.PromptFocused)),
		updates:      updates,
		theme:        theme,
		themeProfile: themeProfile,
		isDarkBG:     isDarkBG,
	}
}

func (m *taskListModel) Init() tea.Cmd {
	cmds := []tea.Cmd{waitForTaskListUpdateCmd(m.updates)}
	if m.shouldTrackBackground() {
		cmds = append(cmds, requestBackgroundColorCmd(), scheduleBackgroundPollCmd())
	}
	return tea.Batch(cmds...)
}

func (m *taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case taskListUpdateMsg:
		if !msg.ok {
			return m, tea.Quit
		}
		wasRunning := m.hasRunningTasks()
		m.statuses[msg.update.ID] = taskStatus{
			State:  msg.update.State,
			Detail: msg.update.Detail,
		}

		cmds := []tea.Cmd{waitForTaskListUpdateCmd(m.updates)}
		if !wasRunning && m.hasRunningTasks() {
			cmds = append(cmds, m.spinner.Tick)
		}
		return m, tea.Batch(cmds...)
	case spinner.TickMsg:
		if !m.hasRunningTasks() {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.BackgroundColorMsg:
		m.updateThemeForBackground(msg.IsDark())
	case backgroundPollMsg:
		if m.shouldTrackBackground() {
			return m, tea.Batch(requestBackgroundColorCmd(), scheduleBackgroundPollCmd())
		}
	}

	return m, nil
}

func (m *taskListModel) View() tea.View {
	frame := lipgloss.NewStyle().Padding(1, 1)
	view := tea.NewView(frame.Render(renderTaskListContent(m.title, m.tasks, m.statuses, m.spinner.View(), m.theme)))
	return view
}

func (m *taskListModel) shouldTrackBackground() bool {
	return m.themeProfile != nil && m.themeProfile.IsAuto()
}

func (m *taskListModel) updateThemeForBackground(isDark bool) {
	if !m.shouldTrackBackground() || m.isDarkBG == isDark {
		return
	}
	m.isDarkBG = isDark
	m.theme = m.themeProfile.Theme(isDark)
	m.spinner.Style = m.theme.PromptFocused
}

func (m *taskListModel) hasRunningTasks() bool {
	for _, status := range m.statuses {
		if status.State == TaskStateRunning {
			return true
		}
	}
	return false
}

func waitForTaskListUpdateCmd(updates <-chan TaskUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-updates
		return taskListUpdateMsg{update: update, ok: ok}
	}
}

func renderTaskListContent(
	title string,
	tasks []Task,
	statuses map[string]taskStatus,
	spinnerFrame string,
	theme *Theme,
) string {
	lines := make([]string, 0, len(tasks)+1)
	if title != "" {
		lines = append(lines, theme.Title.Render(title), "")
	}

	parents := make(map[string]string, len(tasks))
	for _, task := range tasks {
		parents[task.ID] = task.ParentID
	}

	for _, task := range tasks {
		status, ok := statuses[task.ID]
		if !ok {
			status = taskStatus{State: TaskStatePending}
		}
		lines = append(lines, renderTaskLine(task, taskDepth(task, parents), status, spinnerFrame, theme))
	}

	return strings.Join(lines, "\n")
}

func renderTaskLine(task Task, depth int, status taskStatus, spinnerFrame string, theme *Theme) string {
	symbol := renderTaskSymbol(status.State, spinnerFrame, theme)
	labelStyle := theme.Text
	if depth == 0 {
		labelStyle = theme.LabelFocused
	}

	line := strings.Repeat(taskListIndent, depth) + symbol + " " + labelStyle.Render(task.Label)
	if strings.TrimSpace(status.Detail) == "" {
		return line
	}

	detailStyle := theme.Help
	if status.State == TaskStateFailed {
		detailStyle = theme.Error
	}
	return line + " " + detailStyle.Render("("+status.Detail+")")
}

func renderTaskSymbol(state TaskState, spinnerFrame string, theme *Theme) string {
	switch state {
	case TaskStateRunning:
		return theme.PromptFocused.Render("[" + spinnerFrame + "]")
	case TaskStateDone:
		return theme.LabelFocused.Render("[✓]")
	case TaskStateFailed:
		return theme.Error.Render("[✗]")
	default:
		return theme.Help.Render("[ ]")
	}
}

func taskDepth(task Task, parents map[string]string) int {
	depth := 0
	current := task.ParentID
	for current != "" {
		depth++
		next, ok := parents[current]
		if !ok {
			break
		}
		current = next
	}
	return depth
}

// TaskList renders and updates a themed Bubble Tea task list until the update
// channel is closed.
func TaskList(title string, tasks []Task, updates <-chan TaskUpdate, opts ...Option) error {
	defaultProfile := DefaultThemeProfile()
	config := options{input: os.Stdin, output: os.Stdout, themeProfile: &defaultProfile}
	for _, opt := range opts {
		opt(&config)
	}

	if len(tasks) == 0 {
		return errors.New("task list requires at least one task")
	}
	if updates == nil {
		return errors.New("task list updates channel is required")
	}

	initialDark := true
	if dark, ok := detectDarkBackground(config.input, config.output); ok {
		initialDark = dark
	}

	model := newTaskListModel(
		title,
		tasks,
		updates,
		resolveTheme(config, initialDark),
		config.themeProfile,
		initialDark,
	)
	_, err := tea.NewProgram(
		model,
		tea.WithInput(config.input),
		tea.WithOutput(config.output),
		tea.WithEnvironment(bubbleTeaEnvironment()),
	).Run()
	return err
}
