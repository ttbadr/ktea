package kadmin

import tea "github.com/charmbracelet/bubbletea"

type ConfigUpdater interface {
	UpdateConfig(t TopicConfigToUpdate) tea.Msg
}

type TopicConfigUpdatedMsg struct{}

type TopicConfigToUpdate struct {
	Topic string
	Key   string
	Value string
}

type UpdateTopicConfigErrorMsg struct {
	Reason string
}

func (ka *SaramaKafkaAdmin) UpdateConfig(t TopicConfigToUpdate) tea.Msg {
	err := ka.admin.AlterConfig(
		TOPIC_RESOURCE_TYPE,
		t.Topic,
		map[string]*string{t.Key: &t.Value},
		false,
	)
	if err != nil {
		return KAdminErrorMsg{err}
	}
	return TopicConfigUpdatedMsg{}
}
