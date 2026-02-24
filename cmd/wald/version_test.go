package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderVersion(t *testing.T) {
	originalVersion := version
	originalCommit := commit
	originalDate := date
	t.Cleanup(func() {
		version = originalVersion
		commit = originalCommit
		date = originalDate
	})

	version = "v1.2.3"
	commit = "abc1234"
	date = "2026-02-12T22:00:00Z"

	require.Equal(t, "wald v1.2.3 (commit abc1234, built 2026-02-12T22:00:00Z)", renderVersion())
}
