package cmdbar

import (
	"github.com/stretchr/testify/assert"
	"ktea/tests/keys"
	"ktea/ui"
	"testing"
)

func TestSearchCmdBar(t *testing.T) {
	t.Run("Is hidden by default", func(t *testing.T) {
		model := NewSearchCmdBar(">")

		render := model.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Empty(t, render)

		t.Run("/ activates search", func(t *testing.T) {
			model.Update(keys.Key('/'))

			render = model.View(ui.NewTestKontext(), ui.TestRenderer)

			assert.Contains(t, render, ">")
		})
	})

	t.Run("Resets search value upon toggle", func(t *testing.T) {
		model := NewSearchCmdBar(">")

		model.Update(keys.Key('/'))

		model.Update(keys.Key('s'))
		model.Update(keys.Key('e'))
		model.Update(keys.Key('a'))
		model.Update(keys.Key('r'))
		model.Update(keys.Key('c'))
		model.Update(keys.Key('h'))

		render := model.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Contains(t, render, "> search")

		model.Update(keys.Key('/'))

		render = model.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.NotContains(t, render, "> search")
	})
}
