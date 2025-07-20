package ui

import tea "github.com/charmbracelet/bubbletea"

// NavBack logically navigates back
type NavBack func() tea.Cmd

type NavBackMockCalledMsg struct{}

// NavBackMock mocks NavBack by returning a tea.Cmd that returns a NavBackMockCalledMsg
func NavBackMock() tea.Cmd {
	return func() tea.Msg {
		return NavBackMockCalledMsg{}
	}
}
