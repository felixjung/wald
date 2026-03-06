package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigIfNeededSkipsForConfigInit(t *testing.T) {
	cfg, err := loadConfigIfNeeded([]string{"config", "init"})
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestLoadConfigIfNeededLoadsForConfigAdd(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("HOME", temp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(temp, "xdg"))

	_, err := loadConfigIfNeeded([]string{"config", "add"})
	require.ErrorContains(t, err, "config not found")
}

func TestIsConfigInitCommand(t *testing.T) {
	require.True(t, isConfigInitCommand([]string{"config", "init"}))
	require.True(t, isConfigInitCommand([]string{"--verbose", "config", "init"}))
	require.False(t, isConfigInitCommand([]string{"config", "add"}))
}
