package kadmin

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/sradmin"
)

type MockKadmin struct {
}

func (m MockKadmin) CreateTopic(tcd TopicCreationDetails) tea.Msg {
	return nil
}

func (m MockKadmin) DeleteTopic(topic string) tea.Msg {
	return nil
}

func (m MockKadmin) ListTopics() tea.Msg {
	return nil
}

func (m MockKadmin) PublishRecord(p *ProducerRecord) PublicationStartedMsg {
	return PublicationStartedMsg{}
}

func (m MockKadmin) ReadRecords(ctx context.Context, rd ReadDetails) ReadingStartedMsg {
	return ReadingStartedMsg{}
}

func (m MockKadmin) ListOffsets(group string) tea.Msg {
	return nil
}

func (m MockKadmin) ListConsumerGroups() tea.Msg {
	return nil
}

func (m MockKadmin) UpdateConfig(t TopicConfigToUpdate) tea.Msg {
	return nil
}

func (m MockKadmin) ListConfigs(topic string) tea.Msg {
	return nil
}

func (m MockKadmin) SetSra(sra sradmin.SrAdmin) {
}

func NewMockKadmin() Instantiator {
	return func(cd ConnectionDetails) (Kadmin, error) {
		return &MockKadmin{}, nil
	}
}
