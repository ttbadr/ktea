package record_details_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/tests/keys"
	"ktea/ui"
	"testing"
)

func TestRecordDetailsPage(t *testing.T) {
	t.Run("c-h or arrows toggles focus between content and headers", func(t *testing.T) {
		m := New(&kadmin.ConsumerRecord{
			Key:       "",
			Value:     "",
			Partition: 0,
			Offset:    0,
			Headers: []kadmin.Header{
				{
					Key:   "h1",
					Value: "v2",
				},
			},
		}, &kadmin.Topic{
			Name:       "",
			Partitions: 0,
			Replicas:   0,
			Isr:        0,
		})
		// init ui
		m.View(ui.TestKontext, ui.TestRenderer)

		assert.Equal(t, payloadFocus, m.focus)

		m.Update(keys.Key(tea.KeyCtrlH))

		assert.Equal(t, headersFocus, m.focus)

		m.Update(keys.Key(tea.KeyCtrlH))

		assert.Equal(t, payloadFocus, m.focus)

		m.Update(keys.Key(tea.KeyRight))

		assert.Equal(t, headersFocus, m.focus)

		m.Update(keys.Key(tea.KeyLeft))

		assert.Equal(t, payloadFocus, m.focus)
	})

	t.Run("Display record without headers", func(t *testing.T) {
		m := New(&kadmin.ConsumerRecord{
			Key:       "",
			Value:     "",
			Partition: 0,
			Offset:    0,
			Headers:   nil,
		}, &kadmin.Topic{
			Name:       "",
			Partitions: 0,
			Replicas:   0,
			Isr:        0,
		})

		render := m.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Contains(t, render, "No headers present")
	})
}
