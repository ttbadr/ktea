package cmdbar

import (
	"github.com/stretchr/testify/assert"
	"ktea/tests/keys"
	"ktea/ui"
	"testing"
)

func TestSearchCmdBar(t *testing.T) {
	t.Run("Is hidden by default", func(t *testing.T) {
		model := NewSearchCmdBar("placeholder")

		render := model.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Empty(t, render)

		t.Run("/ activates search", func(t *testing.T) {
			model.Update(keys.Key('/'))

			render = model.View(ui.NewTestKontext(), ui.TestRenderer)

			assert.Contains(t, render, "placeholder")
		})
	})
}
