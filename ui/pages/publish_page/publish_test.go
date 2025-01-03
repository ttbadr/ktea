package publish_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/tests/keys"
	"ktea/ui"
	"ktea/ui/pages/navigation"
	"testing"
)

type MockPublisher struct {
	PublishRecordFunc func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg
}

func (m *MockPublisher) PublishRecord(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
	if m.PublishRecordFunc != nil {
		return m.PublishRecordFunc(p)
	}
	return kadmin.PublicationStartedMsg{}
}

func TestPublish(t *testing.T) {
	t.Run("esc goes back to topic list page", func(t *testing.T) {
		m := New(&MockPublisher{}, kadmin.Topic{
			Name:       "topic1",
			Partitions: 1,
			Replicas:   1,
			Isr:        1,
		})

		cmd := m.Update(keys.Key(tea.KeyEsc))

		assert.IsType(t, navigation.LoadTopicsPageMsg{}, cmd())
	})

	t.Run("publish plain text", func(t *testing.T) {
		var producerRecord *kadmin.ProducerRecord
		m := New(&MockPublisher{
			PublishRecordFunc: func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
				producerRecord = p
				return kadmin.PublicationStartedMsg{}
			},
		}, kadmin.Topic{
			Name:       "topic1",
			Partitions: 10,
			Replicas:   1,
			Isr:        1,
		})

		m.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, ui.TestRenderer)
		// Key
		keys.UpdateKeys(m, "key")
		cmd := m.Update(keys.Key(tea.KeyEnter))
		m.Update(cmd())
		// Partition
		keys.UpdateKeys(m, "2")
		cmd = m.Update(keys.Key(tea.KeyEnter))
		m.Update(cmd())
		// payload
		keys.UpdateKeys(m, "payload")
		cmd = m.Update(keys.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		cmd = m.Update(cmd())
		// execute cmd
		executeBatchCmd(cmd)

		assert.Equal(t, "key", producerRecord.Key)
		assert.Equal(t, "topic1", producerRecord.Topic)
		assert.Equal(t, 2, *producerRecord.Partition)
		assert.Equal(t, "payload", producerRecord.Value)
	})

	t.Run("publish without partition info", func(t *testing.T) {
		var producerRecord *kadmin.ProducerRecord
		m := New(&MockPublisher{
			PublishRecordFunc: func(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
				producerRecord = p
				return kadmin.PublicationStartedMsg{}
			},
		}, kadmin.Topic{
			Name:       "topic1",
			Partitions: 10,
			Replicas:   1,
			Isr:        1,
		})

		m.View(&kontext.ProgramKtx{
			WindowWidth:  100,
			WindowHeight: 100,
		}, ui.TestRenderer)
		// Key
		keys.UpdateKeys(m, "key")
		cmd := m.Update(keys.Key(tea.KeyEnter))
		m.Update(cmd())
		// Partition
		cmd = m.Update(keys.Key(tea.KeyEnter))
		m.Update(cmd())
		// payload
		keys.UpdateKeys(m, "payload")
		cmd = m.Update(keys.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		cmd = m.Update(cmd())
		// execute cmd
		executeBatchCmd(cmd)

		assert.Equal(t, "key", producerRecord.Key)
		assert.Equal(t, "topic1", producerRecord.Topic)
		assert.Nil(t, producerRecord.Partition)
		assert.Equal(t, "payload", producerRecord.Value)
	})

	t.Run("Validate", func(t *testing.T) {

		t.Run("When partition is not a number", func(t *testing.T) {
			m := New(&MockPublisher{}, kadmin.Topic{
				Name:       "topic1",
				Partitions: 1,
				Replicas:   1,
				Isr:        1,
			})

			m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, ui.TestRenderer)
			// Key
			keys.UpdateKeys(m, "key")
			cmd := m.Update(keys.Key(tea.KeyEnter))
			m.Update(cmd())
			// Partition
			keys.UpdateKeys(m, "a1")
			m.Update(keys.Key(tea.KeyEnter))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, ui.TestRenderer)
			assert.Contains(t, render, "'a1' is not a valid numeric partition value")
		})

		t.Run("When partition is negative", func(t *testing.T) {
			m := New(&MockPublisher{}, kadmin.Topic{
				Name:       "topic1",
				Partitions: 1,
				Replicas:   1,
				Isr:        1,
			})

			m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, ui.TestRenderer)
			// Key
			keys.UpdateKeys(m, "key")
			cmd := m.Update(keys.Key(tea.KeyEnter))
			m.Update(cmd())
			// Partition
			keys.UpdateKeys(m, "-1")
			m.Update(keys.Key(tea.KeyEnter))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, ui.TestRenderer)
			assert.Contains(t, render, "value must be greater than zero")
		})

		t.Run("When partition exceeds number of partitions", func(t *testing.T) {
			m := New(&MockPublisher{}, kadmin.Topic{
				Name:       "topic1",
				Partitions: 5,
				Replicas:   1,
				Isr:        1,
			})

			m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, ui.TestRenderer)
			// Key
			keys.UpdateKeys(m, "key")
			cmd := m.Update(keys.Key(tea.KeyEnter))
			m.Update(cmd())
			// Partition
			keys.UpdateKeys(m, "10")
			m.Update(keys.Key(tea.KeyEnter))

			render := m.View(&kontext.ProgramKtx{
				WindowWidth:  100,
				WindowHeight: 100,
			}, ui.TestRenderer)
			assert.Contains(t, render, "partition index 10 is invalid, valid range is 0-4")
		})
	})
}

func executeBatchCmd(cmd tea.Cmd) []tea.Msg {
	var msgs []tea.Msg
	if cmd == nil {
		return msgs
	}

	msg := cmd()
	if msg == nil {
		return msgs
	}

	// If the message is a batch, process its commands
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, subCmd := range batch {
			if subCmd != nil {
				msgs = append(msgs, executeBatchCmd(subCmd)...)
			}
		}
		return msgs
	}

	// Otherwise, it's a normal message
	msgs = append(msgs, msg)
	return msgs
}
