package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShellInitScriptSupportsKnownShells(t *testing.T) {
	fish, err := shellInitScript("fish")
	require.NoError(t, err)
	require.Contains(t, fish, "function forest")
	require.Contains(t, fish, "FOREST_SWITCH_OUT_FILE")

	zsh, err := shellInitScript("zsh")
	require.NoError(t, err)
	require.Contains(t, zsh, `forest() {`)
	require.Contains(t, zsh, "FOREST_SWITCH_OUT_FILE")

	bash, err := shellInitScript("bash")
	require.NoError(t, err)
	require.Contains(t, bash, `forest() {`)
	require.Contains(t, bash, "FOREST_SWITCH_OUT_FILE")
}

func TestShellInitScriptRejectsUnknownShell(t *testing.T) {
	_, err := shellInitScript("pwsh")
	require.EqualError(t, err, "unsupported shell: use fish, zsh, or bash")
}
