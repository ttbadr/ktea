package main

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kadmin"
	"ktea/tests/keys"
	"testing"
)

type MockKadmin struct {
}

func (m MockKadmin) CreateTopic(tcd kadmin.TopicCreationDetails) tea.Msg {
	return nil
}

func (m MockKadmin) DeleteTopic(topic string) tea.Msg {
	return nil
}

func (m MockKadmin) ListTopics() tea.Msg {
	return nil
}

func (m MockKadmin) PublishRecord(p *kadmin.ProducerRecord) kadmin.PublicationStartedMsg {
	return kadmin.PublicationStartedMsg{}
}

func (m MockKadmin) ReadRecords(ctx context.Context, rd kadmin.ReadDetails) kadmin.ReadingStartedMsg {
	return kadmin.ReadingStartedMsg{}
}

func (m MockKadmin) ListOffsets(group string) tea.Msg {
	return nil
}

func (m MockKadmin) ListConsumerGroups() tea.Msg {
	return nil
}

func (m MockKadmin) UpdateConfig(t kadmin.TopicConfigToUpdate) tea.Msg {
	return nil
}

func (m MockKadmin) ListConfigs(topic string) tea.Msg {
	return nil
}

func newMockKadmin() kadmin.Instantiator {
	return func(cd kadmin.ConnectionDetails) (kadmin.Kadmin, error) {
		return &MockKadmin{}, nil
	}
}

func TestKtea(t *testing.T) {
	t.Run("No clusters configured", func(t *testing.T) {
		t.Run("Shows create cluster page", func(t *testing.T) {
			model := NewModel(newMockKadmin(), config.NewInMemoryConfigIO(&config.Config{}))
			model.Update(config.LoadedMsg{
				Config: &config.Config{},
			})
			view := model.View()

			assert.Contains(t, view, "┃ Name")
		})
	})

	t.Run("Tabs", func(t *testing.T) {
		t.Run("Cycle through tabs", func(t *testing.T) {
			io := config.NewInMemoryConfigIO(
				&config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            "#808080",
							Active:           true,
							BootstrapServers: []string{":19092"},
							SchemaRegistry: &config.SchemaRegistryConfig{
								Url:      "",
								Username: "",
								Password: "",
							},
							SASLConfig: nil,
						},

						{
							Name:             "tst",
							Color:            "#808080",
							Active:           true,
							BootstrapServers: []string{":19092"},
							SchemaRegistry:   nil,
							SASLConfig:       nil,
						},
					},
				},
			)
			model := NewModel(newMockKadmin(), io)
			model.Update(config.LoadedMsg{Config: config.New(io)})
			//model.Update(config.LoadedMsg{
			//	Config: &config.Config{
			//	},
			//})

			model.Update(tea.WindowSizeMsg{
				Width:  100,
				Height: 100,
			})

			model.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'2'},
				Alt:   true,
				Paste: false,
			})

			view := model.View()

			var expectedLayout = `
╭────────────────╮╭─────────────────────────╮╭─────────────────────────╮╭──────────────────╮        
│ Topics (Alt-1) ││ Consumer Groups (Alt-2) ││ Schema Registry (Alt-3) ││ Clusters (Alt-4) │        
┴────────────────┴┘                         └┴─────────────────────────┴┴──────────────────┴────────
`
			assert.Contains(t, view, expectedLayout)

			model.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'3'},
				Alt:   true,
				Paste: false,
			})

			view = model.View()

			expectedLayout = `
╭────────────────╮╭─────────────────────────╮╭─────────────────────────╮╭──────────────────╮        
│ Topics (Alt-1) ││ Consumer Groups (Alt-2) ││ Schema Registry (Alt-3) ││ Clusters (Alt-4) │        
┴────────────────┴┴─────────────────────────┴┘                         └┴──────────────────┴────────
`
			assert.Contains(t, view, expectedLayout)

			model.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'4'},
				Alt:   true,
				Paste: false,
			})

			assert.Contains(t, view, expectedLayout)

			model.View()
			_, cmd := model.Update(keys.Key(tea.KeyDown))
			_, cmd = model.Update(keys.Key(tea.KeyEnter))
			model.Update(cmd())

			view = model.View()

			expectedLayout = `
╭────────────────╮╭─────────────────────────╮╭──────────────────╮                                   
│ Topics (Alt-1) ││ Consumer Groups (Alt-2) ││ Clusters (Alt-3) │                                   
┴────────────────┴┴─────────────────────────┴┘                  └───────────────────────────────────
`
			assert.Regexp(t, "X\\W+tst", view)
			assert.NotRegexp(t, "X\\W+prd", view)
		})
	})
}
