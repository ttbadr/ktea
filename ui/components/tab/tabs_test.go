package tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/tests"
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

	t.Run("Rendering", func(t *testing.T) {
		t.Run("Multiple tabs", func(t *testing.T) {
			// given
			tabs := New("tab1", "tab2")
			// when
			actual := tabs.View(tests.TestKontext, tests.TestRenderer)
			// then
			assert.Equal(t, "╭───────────────╮╭───────────────╮\n│ tab1 (Meta-1) ││ tab2 (Meta-2) │\n┘               └┴───────────────┴──────────────────────────────────────────────────────────────────", actual)
		})

		t.Run("Single tab no shortcut", func(t *testing.T) {
			// given
			tabs := New("tab1")
			// when
			actual := tabs.View(tests.TestKontext, tests.TestRenderer)
			// then
			assert.NotContains(t, actual, "Meta")
		})

	})
}
