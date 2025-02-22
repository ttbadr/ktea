package consumption_page

import (
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"testing"
)

func TestConsumptionPage(t *testing.T) {
	t.Run("Display empty topic message and adjusted shortcuts", func(t *testing.T) {
		m, _ := New(nil, kadmin.ReadDetails{}, &kadmin.ListedTopic{})

		m.Update(kadmin.EmptyTopicMsg{})

		render := m.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Contains(t, render, "Empty topic")

		assert.Equal(t, []statusbar.Shortcut{{"Go Back", "esc"}}, m.Shortcuts())
	})
}
