package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kadmin"
	"testing"
)

func TestKtea(t *testing.T) {
	t.Run("No clusters configured", func(t *testing.T) {
		t.Run("Shows create cluster page", func(t *testing.T) {
			model := NewModel(kadmin.NewMockKadminInstantiator(), config.NewInMemoryConfigIO(&config.Config{}))
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
			model := NewModel(kadmin.NewMockKadminInstantiator(), io)
			model.Update(config.LoadedMsg{Config: config.New(io)})

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
╭─────────────────╮╭──────────────────────────╮╭──────────────────────────╮╭───────────────────╮    
│ Topics (Meta-1) ││ Consumer Groups (Meta-2) ││ Schema Registry (Meta-3) ││ Clusters (Meta-4) │    
┴─────────────────┴┘                          └┴──────────────────────────┴┴───────────────────┴────
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
╭─────────────────╮╭──────────────────────────╮╭──────────────────────────╮╭───────────────────╮    
│ Topics (Meta-1) ││ Consumer Groups (Meta-2) ││ Schema Registry (Meta-3) ││ Clusters (Meta-4) │    
┴─────────────────┴┴──────────────────────────┴┘                          └┴───────────────────┴────
`
			assert.Contains(t, view, expectedLayout)

			model.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'4'},
				Alt:   true,
				Paste: false,
			})

			view = model.View()

			expectedLayout = `
╭─────────────────╮╭──────────────────────────╮╭──────────────────────────╮╭───────────────────╮    
│ Topics (Meta-1) ││ Consumer Groups (Meta-2) ││ Schema Registry (Meta-3) ││ Clusters (Meta-4) │    
┴─────────────────┴┴──────────────────────────┴┴──────────────────────────┴┘                   └────
`

			assert.Contains(t, view, expectedLayout)
		})
	})
}
