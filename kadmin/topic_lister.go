package kadmin

import (
	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"sync"
)

const UnknownRecordCount = -1

type TopicLister interface {
	ListTopics() tea.Msg
}

type TopicListedMsg struct {
	Topics []ListedTopic
}

type TopicRecordCount struct {
	Topic        string
	RecordCount  int64
	CountedTopic chan TopicRecordCount
}

type TopicListingStartedMsg struct {
	Err              chan error
	Topics           chan []ListedTopic
	TopicRecordCount chan TopicRecordCount
}

type TopicListedErrorMsg struct {
	Err error
}

func (m *TopicListingStartedMsg) AwaitTopicListCompletion() tea.Msg {
	select {
	case topics := <-m.Topics:
		return TopicListedMsg{Topics: topics}
	case err := <-m.Err:
		return TopicListedErrorMsg{Err: err}
	}
}

type TopicRecordCountCalculatedMsg struct {
	TopicRecordCount
	TopicRecordCountChan chan TopicRecordCount
}

func (m *TopicListingStartedMsg) AwaitRecordCountCompletion() tea.Msg {
	select {
	case recordCount := <-m.TopicRecordCount:
		return TopicRecordCountCalculatedMsg{recordCount, m.TopicRecordCount}
	case err := <-m.Err:
		return TopicListedErrorMsg{Err: err}
	}
}

type AllTopicRecordCountCalculatedMsg struct {
}

func (m *TopicRecordCountCalculatedMsg) AwaitRecordCountCompletion() tea.Msg {
	select {
	case recordCount, ok := <-m.TopicRecordCountChan:
		// channel closed
		if !ok {
			return AllTopicRecordCountCalculatedMsg{}
		}
		return TopicRecordCountCalculatedMsg{recordCount, m.TopicRecordCountChan}
	}
}

type ListedTopic struct {
	Name           string
	PartitionCount int
	Replicas       int
	RecordCount    int64
}

func (t *ListedTopic) Partitions() []int {
	partToConsume := make([]int, t.PartitionCount)
	for i := range t.PartitionCount {
		partToConsume[i] = i
	}
	return partToConsume
}

func (ka *SaramaKafkaAdmin) ListTopics() tea.Msg {
	errChan := make(chan error)
	topicsChan := make(chan []ListedTopic)
	topicRecordCountChan := make(chan TopicRecordCount)

	go ka.doListTopics(errChan, topicsChan, topicRecordCountChan)

	return TopicListingStartedMsg{
		errChan,
		topicsChan,
		topicRecordCountChan,
	}
}

func (ka *SaramaKafkaAdmin) doListTopics(
	errChan chan error,
	topicsChan chan []ListedTopic,
	recordCountChan chan TopicRecordCount,
) {
	MaybeIntroduceLatency()
	listResult, err := ka.admin.ListTopics()
	if err != nil {
		errChan <- err
		return
	}

	var topics []ListedTopic
	for name, t := range listResult {
		topics = append(topics, ListedTopic{
			name,
			int(t.NumPartitions),
			int(t.ReplicationFactor),
			UnknownRecordCount,
		})
	}
	topicsChan <- topics
	close(topicsChan)

	var wg sync.WaitGroup

	for name, topic := range listResult {
		wg.Add(1)

		go func(name string, topic sarama.TopicDetail) {
			defer wg.Done()

			partitions := make([]int, topic.NumPartitions)
			for i := range topic.NumPartitions {
				partitions[i] = int(i)
			}

			offsets, err := ka.fetchOffsets(partitions, name)
			if err != nil {
				errChan <- err
				return
			}

			var recordCount int64
			for _, offset := range offsets {
				recordCount += offset.firstAvailable - offset.oldest
			}

			recordCountChan <- TopicRecordCount{name, recordCount, recordCountChan}
			log.Debug("done calculating record count", "topic", name)
		}(name, topic)
	}
	wg.Wait()

	close(recordCountChan)
	log.Debug("done fetching offsets")
}
