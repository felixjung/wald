package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/felixjung/wald/internal/app"
	"github.com/felixjung/wald/internal/tui"
	"github.com/urfave/cli/v3"
)

var runInitTaskList = tui.TaskList

func newInitCommand(application appAPI, themeProfile tui.ThemeProfile) *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize worktrees for configured projects",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if shouldUseInitProgressTUI(cmd) {
				return runInitWithTaskList(ctx, application, themeProfile)
			}

			return application.Init(ctx, &initPlainTextReporter{output: os.Stderr})
		},
	}
}

func shouldUseInitProgressTUI(cmd *cli.Command) bool {
	if cmd.Bool("verbose") {
		return false
	}
	return isTerminal(os.Stdin) && isTerminal(os.Stderr)
}

func runInitWithTaskList(ctx context.Context, application appAPI, themeProfile tui.ThemeProfile) error {
	events := make(chan app.InitProgressEvent)
	planReady := make(chan app.InitProgressPlan, 1)
	updates := make(chan tui.TaskUpdate, 32)
	bridgeDone := make(chan struct{})

	go func() {
		defer close(bridgeDone)
		defer close(updates)

		planSent := false
		for event := range events {
			switch event.Type {
			case app.InitProgressEventPlan:
				if !planSent {
					planReady <- event.Plan
					close(planReady)
					planSent = true
				}
			case app.InitProgressEventTask:
				updates <- tui.TaskUpdate{
					ID:     event.TaskID,
					State:  mapInitTaskState(event.State),
					Detail: event.Detail,
				}
			}
		}

		if !planSent {
			close(planReady)
		}
	}()

	var runErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		runErr = application.Init(ctx, &initEventReporter{events: events})
		close(events)
	}()

	plan, ok := <-planReady
	if !ok {
		<-done
		<-bridgeDone
		return runErr
	}

	tasks := make([]tui.Task, 0, len(plan.Tasks))
	for _, task := range plan.Tasks {
		tasks = append(tasks, tui.Task{
			ID:       task.ID,
			ParentID: task.ParentID,
			Label:    task.Label,
		})
	}

	err := runInitTaskList(
		"Initializing worktrees",
		tasks,
		updates,
		tui.WithInput(os.Stdin),
		tui.WithOutput(os.Stderr),
		tui.WithThemeProfile(themeProfile),
	)

	<-done
	<-bridgeDone

	if err != nil {
		return err
	}
	return runErr
}

func mapInitTaskState(state app.InitTaskState) tui.TaskState {
	switch state {
	case app.InitTaskStateRunning:
		return tui.TaskStateRunning
	case app.InitTaskStateDone:
		return tui.TaskStateDone
	case app.InitTaskStateFailed:
		return tui.TaskStateFailed
	default:
		return tui.TaskStatePending
	}
}

type initEventReporter struct {
	events chan<- app.InitProgressEvent
}

func (r *initEventReporter) ReportInitProgress(event app.InitProgressEvent) {
	r.events <- event
}

type initPlainTextReporter struct {
	output io.Writer
	plan   app.InitProgressPlan
}

func (r *initPlainTextReporter) ReportInitProgress(event app.InitProgressEvent) {
	switch event.Type {
	case app.InitProgressEventPlan:
		r.plan = event.Plan
	case app.InitProgressEventTask:
		if err := r.writeTaskUpdate(event); err != nil {
			return
		}
	}
}

func (r *initPlainTextReporter) writeTaskUpdate(event app.InitProgressEvent) error {
	if r.output == nil {
		return nil
	}

	label, depth := r.lookupTask(event.TaskID)
	if label == "" {
		label = event.TaskID
	}

	symbol := "[ ]"
	switch event.State {
	case app.InitTaskStatePending:
		symbol = "[ ]"
	case app.InitTaskStateRunning:
		symbol = "[>]"
	case app.InitTaskStateDone:
		symbol = "[✓]"
	case app.InitTaskStateFailed:
		symbol = "[✗]"
	}

	line := strings.Repeat("  ", depth) + symbol + " " + label
	if detail := strings.TrimSpace(event.Detail); detail != "" {
		line += " (" + detail + ")"
	}
	_, err := fmt.Fprintln(r.output, line)
	return err
}

func (r *initPlainTextReporter) lookupTask(taskID string) (label string, depth int) {
	parents := make(map[string]string, len(r.plan.Tasks))
	for _, task := range r.plan.Tasks {
		parents[task.ID] = task.ParentID
		if task.ID == taskID {
			label = task.Label
		}
	}

	current := parents[taskID]
	for current != "" {
		depth++
		parent, ok := parents[current]
		if !ok {
			break
		}
		current = parent
	}

	return label, depth
}
