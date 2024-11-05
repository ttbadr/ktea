package topics_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/tests/keys"
	"testing"
)

type MockTopicLister struct {
}

func (m MockTopicLister) ListTopics() tea.Msg {
	return nil
}

func TestTopicsPage(t *testing.T) {
	t.Run("Ignore KeyMsg when topics aren't loaded yet", func(t *testing.T) {
		page, _ := New(&MockTopicDeleter{}, &MockTopicLister{})

		msg := page.Update(keys.Key(tea.KeyCtrlN))
		assert.Nil(t, msg)

		msg = page.Update(keys.Key(tea.KeyCtrlI))
		assert.Nil(t, msg)

		msg = page.Update(keys.Key(tea.KeyCtrlP))
		assert.Nil(t, msg)
	})
}
