package kcadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
)

type MockConnectionCheckedMsg struct {
	Config *config.KafkaConnectConfig
}

func CheckConn(c *config.KafkaConnectConfig) tea.Msg {
	return MockConnectionCheckedMsg{c}
}

func NewMockConnChecker() ConnChecker {
	return CheckConn
}
