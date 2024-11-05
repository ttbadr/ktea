package main

import (
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"testing"
)

func TestKtea(t *testing.T) {
	t.Run("No clusters configured", func(t *testing.T) {
		t.Run("Shows create cluster page", func(t *testing.T) {
			model := NewModel()
			model.Update(config.LoadedMsg{
				Config: &config.Config{},
			})
			view := model.View()

			assert.Contains(t, view, "┃ Name")
		})
	})
}
