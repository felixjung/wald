package app

import (
	"github.com/felixjung/wald/internal/config"
	"github.com/felixjung/wald/internal/hooks"
)

func hookVars(project config.Project, worktree, worktreePath, targetPath string) hooks.Vars {
	return hooks.Vars{
		"project":         project.Name,
		"worktree":        worktree,
		"repo":            project.Repo,
		"default_branch":  project.DefaultBranch,
		"project_workdir": project.Workdir,
		"worktree_path":   worktreePath,
		"target_path":     targetPath,
	}
}
