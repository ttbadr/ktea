package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kadmin"
	"ktea/tests/keys"
	"testing"
)

func TestKtea(t *testing.T) {
	t.Run("No clusters configured", func(t *testing.T) {
		t.Run("Shows create cluster page", func(t *testing.T) {
			model := NewModel(kadmin.NewMockKadmin(), config.NewInMemoryConfigIO(&config.Config{}))
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
			model := NewModel(kadmin.NewMockKadmin(), io)
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
