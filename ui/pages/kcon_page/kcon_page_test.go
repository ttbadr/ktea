package kcon_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kcadmin"
	"ktea/tests"
	"ktea/ui"
	"strings"
	"testing"
)

type MockKcAdmin struct{}

type listActiveConnectorsCalledMsg struct{}

func (m *MockKcAdmin) DeleteConnector(string) tea.Msg {
	return nil
}

func (m *MockKcAdmin) ListActiveConnectors() tea.Msg {
	return listActiveConnectorsCalledMsg{}
}

func TestKconsPage(t *testing.T) {

	t.Run("esc goes back to Kafka Connect Clusters page", func(t *testing.T) {
		page, _ := New(ui.NavBackMock, &MockKcAdmin{}, &MockKcAdmin{}, "connector-name")

		cmd := page.Update(tests.Key(tea.KeyEsc))

		assert.IsType(t, ui.NavBackMockCalledMsg{}, cmd())
	})

	t.Run("F5 refreshes topic list", func(t *testing.T) {
		page, _ := New(ui.NavBackMock, &MockKcAdmin{}, &MockKcAdmin{}, "connector-name")

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		cmd := page.Update(tests.Key(tea.KeyF5))

		assert.Contains(t, tests.ExecuteBatchCmd(cmd), listActiveConnectorsCalledMsg{})

		t.Run("F5 blocks until connectors are loaded", func(t *testing.T) {
			cmd := page.Update(tests.Key(tea.KeyF5))

			assert.Nil(t, cmd)
		})

	})

	t.Run("Default sort by Name Asc", func(t *testing.T) {
		page, _ := New(ui.NavBackMock, &MockKcAdmin{}, &MockKcAdmin{}, "connector-name")

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		render := page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▲ Name")

		t1Idx := strings.Index(render, "a-my-connector")
		t2Idx := strings.Index(render, "b-my-connector")
		t3Idx := strings.Index(render, "z-my-connector")

		assert.Less(t, t1Idx, t3Idx)
		assert.Less(t, t1Idx, t3Idx)
		assert.Less(t, t2Idx, t3Idx)
	})

	t.Run("Toggle sort by Name", func(t *testing.T) {
		page, _ := New(ui.NavBackMock, &MockKcAdmin{}, &MockKcAdmin{}, "connector-name")

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))
		page.Update(tests.Key(tea.KeyEnter))
		render := page.View(tests.NewKontext(), tests.TestRenderer)

		render = page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▼ Name")

		t1Idx := strings.Index(render, "c-my-connector")
		t2Idx := strings.Index(render, "b-my-connector")
		t3Idx := strings.Index(render, "a-my-connector")

		assert.Less(t, t1Idx, t2Idx)
		assert.Less(t, t1Idx, t3Idx)
		assert.Less(t, t2Idx, t3Idx)
	})

	t.Run("Esc hides toggled sort by", func(t *testing.T) {
		page, _ := New(ui.NavBackMock, &MockKcAdmin{}, &MockKcAdmin{}, "connector-name")

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))

		render := page.View(tests.NewKontext(), tests.TestRenderer)
		assert.Contains(t, render, "Name ▲")

		cmd := page.Update(tests.Key(tea.KeyEsc))

		assert.Nil(t, cmd)
		render = page.View(tests.NewKontext(), tests.TestRenderer)
		assert.NotContains(t, render, "Name ▲")
	})

	t.Run("Search for connector", func(t *testing.T) {
		page, _ := New(ui.NavBackMock, &MockKcAdmin{}, &MockKcAdmin{}, "connector-name")

		_ = page.Update(kcadmin.ConnectorsListedMsg{
			Connectors: kcadmin.Connectors{
				"z-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "z-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"b-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "b-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
				"a-my-connector": {
					Status: kcadmin.ConnectorStatus{
						Name: "a-my-connector",
						Connector: kcadmin.ConnectorState{
							State:    "RUNNING",
							WorkerID: "10.0.0.6:8083",
						},
						Tasks: nil,
						Type:  "sink",
					},
				},
			},
		})

		page.Update(tests.Key('/'))
		page.Update(tests.Key('a'))

		render := page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "1/3")
		assert.Greater(t, strings.Index(render, "a-my-connector"), -1)
		assert.Equal(t, strings.Index(render, "b-my-connector"), -1)
		assert.Equal(t, strings.Index(render, "z-my-connector"), -1)
	})
}
