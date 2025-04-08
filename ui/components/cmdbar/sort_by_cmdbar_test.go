package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/tests/keys"
	"ktea/ui"
	"testing"
)

func newTestBar(callback SortSelectedCallback) *SortByCmdBar {
	return NewSortByCmdBar(
		[]SortLabel{
			{
				Label:     "Name",
				Direction: Desc,
			},
			{
				Label:     "Size",
				Direction: Desc,
			},
			{
				Label:     "Date",
				Direction: Desc,
			},
			{
				Label:     "Type",
				Direction: Desc,
			},
		},
		WithSortSelectedCallback(callback),
	)
}

func TestSortByCmdBar(t *testing.T) {

	t.Run("Renders all options", func(t *testing.T) {
		bar := newTestBar(func(label SortLabel) {})

		render := bar.View(ui.NewTestKontext(), ui.TestRenderer)

		assert.Contains(t, `
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃   Name ▼     Size ▼     Date ▼     Type ▼                                                        ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛`, render)
	})

	t.Run("Sorted by", func(t *testing.T) {
		bar := newTestBar(func(label SortLabel) {
		})

		bar.Update(keys.Key(tea.KeyRight))
		bar.Update(keys.Key(tea.KeyRight))
		bar.Update(keys.Key(tea.KeyLeft))
		bar.Update(keys.Key('h'))
		bar.Update(keys.Key('l'))

		assert.Equal(t, SortLabel{
			Label:     "Size",
			Direction: Desc,
		}, bar.SortedBy())
	})

	t.Run("Enter selects new sorting", func(t *testing.T) {
		var selectedLabel SortLabel
		bar := newTestBar(func(label SortLabel) {
			selectedLabel = label
		})
		bar.Update(keys.Key(tea.KeyRight))

		bar.Update(keys.Key(tea.KeyEnter))

		assert.Equal(t, SortLabel{
			Label:     "Size",
			Direction: Desc,
		}, selectedLabel)
	})

	t.Run("Enter selected toggles sorting directions", func(t *testing.T) {
		var selectedLabel SortLabel
		bar := newTestBar(func(label SortLabel) {
			selectedLabel = label
		})

		bar.Update(keys.Key(tea.KeyRight))

		bar.Update(keys.Key(tea.KeyEnter))
		bar.Update(keys.Key(tea.KeyEnter))

		assert.Equal(t, SortLabel{
			Label:     "Size",
			Direction: Asc,
		}, selectedLabel)
	})
}
