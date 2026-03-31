package app

import "github.com/felixjung/wald/internal/config"

// InitProgressEventType identifies the kind of init progress event.
type InitProgressEventType string

const (
	// InitProgressEventPlan announces the full task list before work begins.
	InitProgressEventPlan InitProgressEventType = "plan"
	// InitProgressEventTask updates the state of a single task.
	InitProgressEventTask InitProgressEventType = "task"
)

// InitTaskState describes a task's current status.
type InitTaskState string

const (
	InitTaskStatePending InitTaskState = "pending"
	InitTaskStateRunning InitTaskState = "running"
	InitTaskStateDone    InitTaskState = "done"
	InitTaskStateFailed  InitTaskState = "failed"
)

const (
	initTaskPrepareFolders     = "prepare-folders"
	initTaskCheckGitHubCLI     = "check-github-cli"
	initTaskGitHubAvailability = "check-github-cli:availability"
	initTaskGitHubAuth         = "check-github-cli:authentication"
	initTaskCloneDefaultBranch = "clone-default-branches"
)

// InitTask describes one row in the init progress task list.
type InitTask struct {
	ID       string
	ParentID string
	Label    string
}

// InitProgressPlan contains the full task list for init.
type InitProgressPlan struct {
	Tasks []InitTask
}

// InitProgressEvent communicates init progress updates to a reporter.
type InitProgressEvent struct {
	Type   InitProgressEventType
	Plan   InitProgressPlan
	TaskID string
	State  InitTaskState
	Detail string
}

// InitProgressReporter receives structured init progress updates.
type InitProgressReporter interface {
	ReportInitProgress(event InitProgressEvent)
}

// NopInitProgressReporter ignores all init progress events.
type NopInitProgressReporter struct{}

// ReportInitProgress implements InitProgressReporter.
func (NopInitProgressReporter) ReportInitProgress(InitProgressEvent) {}

func initTaskIDForProject(parentID string, project config.Project) string {
	return parentID + ":" + project.Name
}

func initProgressPlan(projects []config.Project) InitProgressPlan {
	tasks := make([]InitTask, 0, 6+len(projects)*2)
	tasks = append(tasks, InitTask{
		ID:    initTaskPrepareFolders,
		Label: "Prepare project folders",
	})
	for _, project := range projects {
		tasks = append(tasks, InitTask{
			ID:       initTaskIDForProject(initTaskPrepareFolders, project),
			ParentID: initTaskPrepareFolders,
			Label:    project.Name,
		})
	}
	tasks = append(tasks,
		InitTask{
			ID:    initTaskCheckGitHubCLI,
			Label: "Check GitHub CLI",
		},
		InitTask{
			ID:       initTaskGitHubAvailability,
			ParentID: initTaskCheckGitHubCLI,
			Label:    "availability",
		},
		InitTask{
			ID:       initTaskGitHubAuth,
			ParentID: initTaskCheckGitHubCLI,
			Label:    "authentication",
		},
		InitTask{
			ID:    initTaskCloneDefaultBranch,
			Label: "Clone default branches",
		},
	)
	for _, project := range projects {
		tasks = append(tasks, InitTask{
			ID:       initTaskIDForProject(initTaskCloneDefaultBranch, project),
			ParentID: initTaskCloneDefaultBranch,
			Label:    project.Name,
		})
	}
	return InitProgressPlan{Tasks: tasks}
}

func reportInitProgress(reporter InitProgressReporter, event InitProgressEvent) {
	if reporter == nil {
		return
	}
	reporter.ReportInitProgress(event)
}
