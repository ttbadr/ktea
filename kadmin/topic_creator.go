package kadmin

import (
	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
)

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

func (ka *SaramaKafkaAdmin) doCreateTopic(tcd TopicCreationDetails, created chan bool, errChan chan error) {
	err := ka.admin.CreateTopic(tcd.Name, &sarama.TopicDetail{
		NumPartitions:     int32(tcd.NumPartitions),
		ReplicationFactor: 1,
		ReplicaAssignment: nil,
		ConfigEntries:     nil,
	}, false)
	if err != nil {
		errChan <- err
	}
	created <- true
}
