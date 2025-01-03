package kadmin

import tea "github.com/charmbracelet/bubbletea"

type TopicCreator interface {
	CreateTopic(tcd TopicCreationDetails) tea.Msg
}

type TopicCreationDetails struct {
	Name          string
	NumPartitions int
	Properties    map[string]string
}

type TopicCreatedMsg struct {
}

type TopicCreationStartedMsg struct {
	Created chan bool
	Err     chan error
}

func (ka *SaramaKafkaAdmin) CreateTopic(tcd TopicCreationDetails) tea.Msg {
	created := make(chan bool)
	err := make(chan error)

	go ka.doCreateTopic(tcd, created, err)

	return TopicCreationStartedMsg{
		Created: created,
		Err:     err,
	}

}
