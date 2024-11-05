package kadmin

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

const (
	PLAIN_TEXT SASLProtocol = 0
	SSL        SASLProtocol = 1
)

const (
	TOPIC_RESOURCE_TYPE = 2
	DEFAULT_TIMEOUT     = 10 * time.Second
)

type SASLProtocol int

type SASLConfig struct {
	Username string
	Password string
	Protocol SASLProtocol
}

type TopicCreator interface {
	CreateTopic(tcd TopicCreationDetails) tea.Msg
}

type TopicLister interface {
	ListTopics() tea.Msg
}

type TopicConfigLister interface {
	ListConfigs(topic string) tea.Msg
}

type TopicDeleter interface {
	DeleteTopic(topic string) tea.Msg
}

type RecordReader interface {
	ReadRecords(ctx context.Context, rd ReadDetails) ReadingStartedMsg
}

type ConfigUpdater interface {
	UpdateConfig(t TopicConfigToUpdate) tea.Msg
}

type ConsumerGroupLister interface {
	ListConsumerGroups() tea.Msg
}

type ConsumerGroupOffsetLister interface {
	ListOffsets(group string) tea.Msg
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

type Publisher interface {
	PublishRecord(p *ProducerRecord) PublicationStartedMsg
}

type ProducerRecord struct {
	Key       string
	Value     string
	Topic     string
	Partition *int
}

type PublicationStartedMsg struct {
	Err       chan error
	Published chan bool
}

type ReadDetails struct {
	Topic      Topic
	Partitions []int
}

type ReadingStartedMsg struct {
	ConsumerRecord chan ConsumerRecord
	Err            chan error
}

type ConsumerRecord struct {
	Key       string
	Value     string
	Partition int64
	Offset    int64
	Headers   map[string]string
}

type TopicConfigListingStartedMsg struct {
	Err     chan error
	Configs chan map[string]string
}

type ConsumerGroup struct {
	Name    string
	Members []GroupMember
}

type GroupMember struct {
	MemberId   string
	ClientId   string
	ClientHost string
}

type ConsumerGroupsListedMsg struct {
	ConsumerGroups []*ConsumerGroup
}

type ConsumerGroupListingErrorMsg struct {
	Err error
}

type OffsetListingErrorMsg struct {
	Err error
}

type OffsetListingStartedMsg struct {
	Err     chan error
	Offsets chan []TopicPartitionOffset
}

func (msg *OffsetListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case offsets := <-msg.Offsets:
		return OffsetListedMsg{offsets}
	case err := <-msg.Err:
		return OffsetListingErrorMsg{err}
	}
}

type OffsetListedMsg struct {
	Offsets []TopicPartitionOffset
}

type ConsumerGroupListingStartedMsg struct {
	Err            chan error
	ConsumerGroups chan []*ConsumerGroup
}

func (msg *ConsumerGroupListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case groups := <-msg.ConsumerGroups:
		return ConsumerGroupsListedMsg{groups}
	case err := <-msg.Err:
		return ConsumerGroupListingErrorMsg{err}
	}
}

type TopicDeletedMsg struct {
	TopicName string
}

func (msg *TopicCreationStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-msg.Created:
		return TopicCreatedMsg{}
	}
}

type TopicListingStartedMsg struct {
	Err    chan error
	Topics chan []Topic
}

type TopicConfigsListedMsg struct {
	Configs map[string]string
}

func (m *TopicConfigListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case e := <-m.Err:
		return TopicConfigListingErrorMsg{e}
	case c := <-m.Configs:
		return TopicConfigsListedMsg{c}
	}
}

type TopicConfigListingErrorMsg struct {
	Err error
}

type TopicCreatedMsg struct {
}

type KAdminErrorMsg struct {
	Error error
}

type TopicCreationDetails struct {
	Name          string
	NumPartitions int
	Properties    map[string]string
}

type TopicConfig struct {
	Key   string
	Value string
}

type TopicCreationStartedMsg struct {
	Created chan bool
	Err     chan error
}

type Topic struct {
	Name       string
	Partitions int
	Replicas   int
	Isr        int
}

type ConnectionDetails struct {
	BootstrapServers []string
	SASLConfig       *SASLConfig
}
