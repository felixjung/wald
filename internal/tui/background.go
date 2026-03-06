package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const backgroundPollInterval = 2 * time.Second

type backgroundPollMsg struct{}

func requestBackgroundColorCmd() tea.Cmd {
	return tea.RequestBackgroundColor
}

func scheduleBackgroundPollCmd() tea.Cmd {
	return tea.Tick(backgroundPollInterval, func(time.Time) tea.Msg {
		return backgroundPollMsg{}
	})
}
