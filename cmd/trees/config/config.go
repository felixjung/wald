package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	EnvXDGConfigHome = "XDG_CONFIG_HOME"

	XDGTreesFolderName     = "trees"
	XDGTreesConfigFileName = "config.yaml"

	HomeConfigFileName = ".trees.yaml"
)

func Paths(getenv func(string) string, homeDir string) (xdgPath, dotPath string) {
	xdg := getenv(EnvXDGConfigHome)
	if xdg == "" {
		xdg = filepath.Join(homeDir, ".config")
	}
	xdgPath = filepath.Join(xdg, XDGTreesFolderName, XDGTreesConfigFileName)
	dotPath = filepath.Join(homeDir, HomeConfigFileName)
	return xdgPath, dotPath
}

func ResolvePath(getenv func(string) string, homeDir string, stat func(string) (os.FileInfo, error)) (string, error) {
	xdgPath, dotPath := Paths(getenv, homeDir)
	if fileExists(stat, xdgPath) {
		return xdgPath, nil
	}
	if fileExists(stat, dotPath) {
		return dotPath, nil
	}
	return "", fmt.Errorf("config not found; create %s or %s", xdgPath, dotPath)
}

func fileExists(stat func(string) (os.FileInfo, error), path string) bool {
	_, err := stat(path)
	return err == nil
}
