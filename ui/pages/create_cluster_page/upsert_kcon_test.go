package create_cluster_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/tests"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"testing"
)

func TestUpsertKcModel(t *testing.T) {
	var ktx = kontext.ProgramKtx{
		WindowWidth:     100,
		WindowHeight:    100,
		AvailableHeight: 100,
		Config: &config.Config{
			Clusters: []config.Cluster{},
		},
	}

	t.Run("Immediately show form when no clusters registered", func(t *testing.T) {
		m := NewUpsertKcModel(
			ui.NavBackMock,
			&ktx,
			nil,
			[]config.KafkaConnectConfig{},
			kcadmin.NewMockConnChecker(),
			cmdbar.NewNotifierCmdBar("test"),
			mockKafkaConnectRegisterer,
		)

		render := m.View(&ktx, tests.TestRenderer)

		assert.Contains(t, render, "Kafka Connect Name")
		assert.Contains(t, render, "Kafka Connect URL")
		assert.Contains(t, render, "Kafka Connect Username")
		assert.Contains(t, render, "Kafka Connect Password")

		t.Run("Tests connection upon creation", func(t *testing.T) {
			tests.UpdateKeys(m, "dev sink cluster")
			cmd := m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())

			tests.UpdateKeys(m, "http://localhost:8083")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())

			tests.UpdateKeys(m, "jane")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())

			tests.UpdateKeys(m, "doe")
			cmd = m.Update(tests.Key(tea.KeyEnter))
			m.Update(cmd())

			msgs := tests.Submit(m)

			assert.Len(t, msgs, 1)
			assert.IsType(t, kcadmin.MockConnectionCheckedMsg{}, msgs[0])
			assert.Equal(t, &config.KafkaConnectConfig{
				Name:     "dev sink cluster",
				Url:      "http://localhost:8083",
				Username: "jane",
				Password: "doe",
			}, msgs[0].(kcadmin.MockConnectionCheckedMsg).Config)
		})

		t.Run("Register Kafka Connect Cluster upon successful connection", func(t *testing.T) {
			cmd := m.Update(kcadmin.ConnCheckSucceededMsg{})

			msgs := tests.ExecuteBatchCmd(cmd)

			assert.Len(t, msgs, 1)
			assert.IsType(t, mockKafkaConnectRegistered{}, msgs[0])
		})
	})

	t.Run("List kafka connect clusters when at least one is already registered", func(t *testing.T) {
		m := NewUpsertKcModel(ui.NavBackMock, &ktx, nil, []config.KafkaConnectConfig{
			{
				Name:     "s3-sink",
				Url:      "http://localhost:8083",
				Username: "john",
				Password: "doe",
			},
		}, kcadmin.NewMockConnChecker(), cmdbar.NewNotifierCmdBar("test"), mockKafkaConnectRegisterer)

		render := m.View(&ktx, tests.TestRenderer)

		assert.NotContains(t, render, "Kafka Connect URL")
		assert.NotContains(t, render, "Kafka Connect Username")
		assert.NotContains(t, render, "Kafka Connect Password")

		assert.Contains(t, render, "s3-sink")
	})
}
