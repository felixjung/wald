package app

import "path/filepath"

func workdirPath(worktreePath, workdir string) string {
	if workdir == "" {
		return worktreePath
	}
	return filepath.Join(worktreePath, filepath.FromSlash(workdir))
}
