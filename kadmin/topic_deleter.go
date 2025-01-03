package kadmin

import tea "github.com/charmbracelet/bubbletea"

type TopicDeleter interface {
	DeleteTopic(topic string) tea.Msg
}

type TopicDeletedMsg struct {
	TopicName string
}

func (ka *SaramaKafkaAdmin) DeleteTopic(topic string) tea.Msg {
	maybeIntroduceLatency()
	err := ka.admin.DeleteTopic(topic)
	if err != nil {
		return KAdminErrorMsg{err}
	}
	return TopicDeletedMsg{topic}
}
