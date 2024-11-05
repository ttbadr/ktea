package configs_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/tests/keys"
	"ktea/ui"
	"strings"
	"testing"
)

type MockKAdmin struct {
	UpdateConfigFunc      func(t kadmin.TopicConfigToUpdate) tea.Msg
	TopicConfigListerFunc func(topic string) tea.Msg
}

func (m *MockKAdmin) UpdateConfig(t kadmin.TopicConfigToUpdate) tea.Msg {
	if m.UpdateConfigFunc != nil {
		return m.UpdateConfigFunc(t)
	}
	return nil
}

func (m *MockKAdmin) ListConfigs(topic string) tea.Msg {
	if m.TopicConfigListerFunc != nil {
		return m.TopicConfigListerFunc(topic)
	}
	return nil
}

func TestConfigsPage(t *testing.T) {
	t.Run("On Config Listing Started show loading message", func(t *testing.T) {
		// given
		section, _ := New(&MockKAdmin{}, &MockKAdmin{}, "topic1")

		// when
		section.Update(kadmin.TopicConfigListingStartedMsg{})

		// then
		render := ansi.Strip(section.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, ui.TestRenderer))
		assert.Contains(t, render, "Loading topic1 Topic Configs")
	})

	t.Run("Keep success message after update when refresh is triggered", func(t *testing.T) {
		// given
		section, _ := New(&MockKAdmin{}, &MockKAdmin{}, "topic1")

		// when
		section.Update(kadmin.TopicConfigUpdatedMsg{})
		section.Update(kadmin.TopicConfigListingStartedMsg{})
		section.Update(kadmin.TopicConfigsListedMsg{
			Configs: map[string]string{"k": "v"},
		})

		// then
		render := ansi.Strip(section.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, ui.TestRenderer))
		assert.Contains(t, render, "Update succeeded")
	})

	t.Run("Show error msg upon update failure", func(t *testing.T) {
		// given
		section, _ := New(&MockKAdmin{}, &MockKAdmin{}, "topic1")

		// when
		section.Update(kadmin.UpdateTopicConfigErrorMsg{
			Reason: "value -1 for configuration cleanup.policy: String must be one of: compact, delete",
		})

		// then
		render := ansi.Strip(section.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, ui.TestRenderer))
		assert.Contains(t, render, "value -1 for configuration cleanup.policy: String must be one of: compact, delete")
	})

	t.Run("When loading ignore going back (esc)", func(t *testing.T) {
		// given
		section, _ := New(&MockKAdmin{}, &MockKAdmin{}, "topic1")
		section.Update(kadmin.TopicConfigListingStartedMsg{})

		// when
		msg := section.Update(keys.Key(tea.KeyEsc))

		// then
		assert.Nil(t, msg)
		render := ansi.Strip(section.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, ui.TestRenderer))
		assert.Contains(t, render, "Loading topic1 Topic Configs")
	})

	// TODO wait until update is async
	//t.Run("When updating ignore going back (esc)", func(t *testing.T) {
	//	// given
	//	section, _ := New(&MockKAdmin{}, &MockKAdmin{}, "topic1")
	//	section.Update(kadmin.TopicConfigListingStartedMsg{})
	//
	//	// when
	//	msg := section.Update(keys.Key(tea.KeyEsc))
	//
	//	// then
	//	assert.Nil(t, msg)
	//	render := ansi.Strip(section.View(&pctx.ProgramContext{
	//		WindowWidth:  100,
	//		WindowHeight: 100,
	//	}))
	//	assert.Contains(t, render, "Loading topic1 Topic Configs")
	//})
}

func TestConfigsPage_Table(t *testing.T) {
	t.Run("Order properties by name desc", func(t *testing.T) {
		section, _ := New(&MockKAdmin{}, &MockKAdmin{}, "topic")

		section.Update(kadmin.TopicConfigsListedMsg{
			Configs: map[string]string{
				"delete.retention.ms": "86400000",
				"cleanup.policy":      "delete",
				"max.message":         "1048588",
				"segment.index":       "10485760",
			},
		})

		render := section.View(&kontext.ProgramKtx{
			WindowHeight:    50,
			WindowWidth:     100,
			AvailableHeight: 100,
		}, ui.TestRenderer)

		cleanupIdx := strings.Index(render, "cleanup.policy")
		deleteRetIdx := strings.Index(render, "delete.retention.ms")
		maxMsgIdx := strings.Index(render, "max.message")
		sgmtIdx := strings.Index(render, " segment.index")
		assert.Less(t, cleanupIdx, deleteRetIdx)
		assert.Less(t, deleteRetIdx, maxMsgIdx)
		assert.Less(t, maxMsgIdx, sgmtIdx)
	})
}

func TestConfigsPage_Searching(t *testing.T) {

	ktx := &kontext.ProgramKtx{
		WindowHeight:    19,
		WindowWidth:     100,
		AvailableHeight: 100,
	}

	t.Run("/ triggers search", func(t *testing.T) {
		section := newSection()
		render := section.View(ktx, ui.TestRenderer)
		assert.NotContains(t, render, "┃ > Search for Config\n")

		section.Update(keys.Key('/'))
		render = section.View(ktx, ui.TestRenderer)

		assert.Contains(t, render, "┃ > Search for Config")
	})

	t.Run("esc cancels search", func(t *testing.T) {
		section := newSection()
		render := section.View(ktx, ui.TestRenderer)
		assert.NotContains(t, render, "┃ > Search for Config")

		section.Update(keys.Key('/'))
		render = section.View(ktx, ui.TestRenderer)
		assert.Contains(t, render, "┃ > Search for Config")

		section.Update(keys.Key(tea.KeyEsc))
		render = section.View(ktx, ui.TestRenderer)
		assert.NotContains(t, render, "┃ > Search for Config\n")
	})

	t.Run("esc resets form", func(t *testing.T) {
		section := newSection()

		section.Update(keys.Key('/'))
		render := section.View(ktx, ui.TestRenderer)
		section.Update(keys.Key('a'))
		render = section.View(ktx, ui.TestRenderer)
		assert.Contains(t, render, "┃ > a ")

		section.Update(keys.Key(tea.KeyEsc))
		render = section.View(ktx, ui.TestRenderer)
		assert.NotContains(t, render, "┃ > Search for Config\n")

		section.Update(keys.Key('/'))
		render = section.View(ktx, ui.TestRenderer)
		assert.Contains(t, render, "┃ > Search for Config")

	})

	t.Run("enter empty form cancels search", func(t *testing.T) {
		section := newSection()

		section.Update(keys.Key('/'))

		section.Update(keys.Key(tea.KeyEnter))

		assert.False(t, section.cmdBar.IsFocused())
	})
}

func TestConfigsPage_Editing(t *testing.T) {
	t.Run("e prompt edit input field", func(t *testing.T) {
		section := newSection()
		render := section.View(&kontext.ProgramKtx{
			WindowHeight: 19,
			WindowWidth:  100,
		}, ui.TestRenderer)

		section.Update(keys.Key('e'))

		render = section.View(&kontext.ProgramKtx{
			WindowHeight: 19,
			WindowWidth:  100,
		}, ui.TestRenderer)
		assert.Contains(t, render, "┃ > delete")
	})

	t.Run("enter updates config", func(t *testing.T) {
		section := newSection()
		render := section.View(&kontext.ProgramKtx{
			WindowHeight: 19,
			WindowWidth:  100,
		}, ui.TestRenderer)

		section.Update(keys.Key('e'))
		section.Update(keys.Key(tea.KeyCtrlU))
		section.Update(keys.Key('f'))
		section.Update(keys.Key(tea.KeyEnter))

		render = section.View(&kontext.ProgramKtx{
			WindowHeight: 19,
			WindowWidth:  100,
		}, ui.TestRenderer)
		assert.NotContains(t, render, "┃ > delete \n")
		assert.Contains(t, render, "Updating Topic Config")
	})
}

func newSection() *Model {
	section, _ := New(&MockKAdmin{}, &MockKAdmin{}, "topic")

	section.Update(kadmin.TopicConfigsListedMsg{
		Configs: map[string]string{
			"delete.retention.ms": "86400000",
			"cleanup.policy":      "delete",
			"max.message":         "1048588",
			"segment.index":       "10485760",
		},
	})
	return section
}
