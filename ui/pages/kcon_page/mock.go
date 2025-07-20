package kcon_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
)

type LoadKConPageMockCalled struct {
	Config config.KafkaConnectConfig
}

func LoadKConPageMock(c config.KafkaConnectConfig) tea.Cmd {
	return func() tea.Msg {
		return LoadKConPageMockCalled{c}
	}
}
