package topics_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/tests/keys"
	"testing"
)

type MockTopicLister struct {
}

type ListTopicsCalledMsg struct{}

func (m *MockTopicLister) ListTopics() tea.Msg {
	return ListTopicsCalledMsg{}
}

type MockTopicDeleter struct {
}

func (m *MockTopicDeleter) DeleteTopic(_ string) tea.Msg {
	return nil
}

func TestTopicsPage(t *testing.T) {
	t.Run("Ignore KeyMsg when topics aren't loaded yet", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		cmd := page.Update(keys.Key(tea.KeyCtrlN))
		assert.Nil(t, cmd)

		cmd = page.Update(keys.Key(tea.KeyCtrlI))
		assert.Nil(t, cmd)

		cmd = page.Update(keys.Key(tea.KeyCtrlP))
		assert.Nil(t, cmd)
	})

	t.Run("F5 refreshes topic list", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		_ = page.Update(kadmin.TopicListedMsg{
			Topics: []kadmin.ListedTopic{
				{
					Name:           "topic1",
					PartitionCount: 1,
					Replicas:       1,
				},
			},
		})

		cmd := page.Update(keys.Key(tea.KeyF5))

		assert.IsType(t, ListTopicsCalledMsg{}, cmd())
	})
}
