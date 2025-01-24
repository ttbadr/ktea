package kadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type TopicLister interface {
	ListTopics() tea.Msg
}

type TopicListedMsg struct {
	Topics []Topic
}

type TopicListingStartedMsg struct {
	Err    chan error
	Topics chan []Topic
}

type TopicListedErrorMsg struct {
	Err error
}

func (m *TopicListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case topics := <-m.Topics:
		return TopicListedMsg{Topics: topics}
	case err := <-m.Err:
		return TopicListedErrorMsg{Err: err}
	}
}

type Topic struct {
	Name       string
	Partitions int
	Replicas   int
	Isr        int
}

func (ka *SaramaKafkaAdmin) ListTopics() tea.Msg {
	errChan := make(chan error)
	topicsChan := make(chan []Topic)

	go ka.doListTopics(errChan, topicsChan)

	return TopicListingStartedMsg{
		errChan,
		topicsChan,
	}
}

func (ka *SaramaKafkaAdmin) doListTopics(errChan chan error, topicsChan chan []Topic) {
	maybeIntroduceLatency()
	listResult, err := ka.admin.ListTopics()
	if err != nil {
		log.Errorf("Error %v while listing topics.", err)
		errChan <- err
	}
	partByTopic := make(map[string]Topic)
	for name, topic := range listResult {
		partByTopic[name] = Topic{
			Name:       name,
			Partitions: int(topic.NumPartitions),
			Replicas:   int(topic.ReplicationFactor),
			Isr:        0,
		}
	}
	var topics []Topic
	for _, t := range partByTopic {
		topics = append(topics, Topic{t.Name, t.Partitions, t.Replicas, t.Isr})
	}
	topicsChan <- topics
}
