package common

import tea "github.com/charmbracelet/bubbletea"

type BackToRootMsg struct{}

func BackToRoot() tea.Cmd {
	return func() tea.Msg {
		return BackToRootMsg{}
	}
}
