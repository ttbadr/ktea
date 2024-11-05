package tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTabs(t *testing.T) {

	t.Run("Movements", func(t *testing.T) {

		t.Run("Select specific tab when available", func(t *testing.T) {
			// given
			tabs := New("tab1", "tab2")
			// when
			tabs.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'2'},
				Alt:   true,
				Paste: false,
			})
			// then
			assert.Equal(t, 1, tabs.activeTab)
		})

		t.Run("Ignore tab selection when tab not available", func(t *testing.T) {
			// given
			tabs := New("tab1", "tab2")
			// when
			tabs.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'8'},
				Alt:   true,
				Paste: false,
			})
			// then
			assert.Equal(t, 0, tabs.activeTab)
		})

		t.Run("Next tab when available", func(t *testing.T) {
			// given
			tabs := New("tab1", "tab2")
			// when
			tabs.Next()
			// then
			assert.Equal(t, 1, tabs.activeTab)
		})

		t.Run("Next tab when last", func(t *testing.T) {
			// given
			tabs := New("tab1", "tab2")
			tabs.Next()
			// when
			tabs.Next()
			// then
			assert.Equal(t, 1, tabs.activeTab)
		})

		t.Run("Previous tab when available", func(t *testing.T) {
			// given
			tabs := New("tab1", "tab2")
			tabs.Next()
			// when
			tabs.Prev()
			// then
			assert.Equal(t, 0, tabs.activeTab)
		})

		t.Run("Previous tab when first", func(t *testing.T) {
			// given
			tabs := New("tab1", "tab2")
			// when
			tabs.Prev()
			// then
			assert.Equal(t, 0, tabs.activeTab)
		})
	})

	//t.Run("Rendering", func(t *testing.T) {
	//	t.Run("Initial render", func(t *testing.T) {
	//		// given
	//		tabs := New("tab1", "tab2")
	//		// when
	//		actual := tabs.View(&context)
	//		// then
	//		assert.Equal(t, " tab1  tab2 ", actual)
	//	})
	//})
}
