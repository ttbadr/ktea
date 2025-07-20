package create_cluster_page

import (
	tea "github.com/charmbracelet/bubbletea"
)

type mockKafkaConnectRegistered struct {
}

func mockKafkaConnectRegisterer() tea.Msg {
	return mockKafkaConnectRegistered{}
}
