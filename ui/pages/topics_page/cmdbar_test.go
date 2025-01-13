package topics_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/ui"
	"testing"
)

type MockTopicDeleter struct {
}

func (m *MockTopicDeleter) DeleteTopic(_ string) tea.Msg {
	return nil
}

func TestNewListTopicsCommandBar_Search(t *testing.T) {

	t.Run("Initially hide command bar", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Equal(t, "", render)
		assert.False(t, bar.IsFocused())
	})

	t.Run("Show search field upon pressing slash", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'/'},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, "┃ > Search for Topic")
		assert.True(t, bar.IsFocused())
	})

	t.Run("Allow searching for /", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'/'},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'/'},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Equal(t, "       \n┃ > a/ ", render)
		assert.Equal(t, "a/", bar.GetSearchTerm())
		assert.True(t, bar.IsFocused())
	})

	t.Run("Quit searching", func(t *testing.T) {

		t.Run("When pressing esc", func(t *testing.T) {
			bar := NewCmdBar(&MockTopicDeleter{})

			bar.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'/'},
				Alt:   false,
				Paste: false,
			}, "topic.a")
			bar.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'a'},
				Alt:   false,
				Paste: false,
			}, "topic.a")

			bar.Update(tea.KeyMsg{
				Type:  tea.KeyEsc,
				Runes: []rune{},
				Alt:   false,
				Paste: false,
			}, "topic.a")
			bar.Update(tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'a'},
				Alt:   false,
				Paste: false,
			}, "topic.a")

			render := bar.View(ui.TestKontext, ui.TestRenderer)

			assert.Equal(t, "", render)
			assert.Equal(t, "", bar.GetSearchTerm())
			assert.False(t, bar.IsFocused())

			t.Run("Reset input field", func(t *testing.T) {
				bar.Update(tea.KeyMsg{
					Type:  tea.KeyRunes,
					Runes: []rune{'/'},
					Alt:   false,
					Paste: false,
				}, "topic.a")

				render := bar.View(ui.TestKontext, ui.TestRenderer)

				assert.Contains(t, render, "┃ > Search for Topic")
				assert.Equal(t, "", bar.GetSearchTerm())
				assert.True(t, bar.IsFocused())
			})
		})
	})

	t.Run("Allow input when searching", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'/'},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, "┃ > a ")
		assert.Equal(t, "a", bar.GetSearchTerm())
		assert.True(t, bar.IsFocused())
	})

	t.Run("Blur search input upon enter", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'/'},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyEnter,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, "  > a ")
		assert.Equal(t, "a", bar.GetSearchTerm())
		assert.False(t, bar.IsFocused())
	})

	t.Run("Ignore esc after confirming (enter) search", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'/'},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyEnter,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type: tea.KeyEsc,
		}, "topic.a")

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, "  > a ")
		assert.Equal(t, "a", bar.GetSearchTerm())
		assert.False(t, bar.IsFocused())
	})

	t.Run("Blur and hide search input upon enter when no input", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'/'},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyEnter,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Equal(t, "", render)
		assert.Equal(t, "", bar.GetSearchTerm())
		assert.False(t, bar.IsFocused())
	})
}

func TestNewListTopicsCommandBar_Delete(t *testing.T) {
	t.Run("c-d raises delete confirmation", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyCtrlD,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, "┃ 🗑️  topic.a will be deleted permanently          Delete!     Cancel.         ")
		assert.True(t, bar.IsFocused())
	})

	t.Run("esc cancels raised delete confirmation", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyCtrlD,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyEsc,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		render := bar.View(ui.TestKontext, ui.TestRenderer)

		assert.Equal(t, "", render)
		assert.False(t, bar.IsFocused())
	})

	t.Run("esc resets raised delete confirmation", func(t *testing.T) {
		bar := NewCmdBar(&MockTopicDeleter{})

		bar.Update(tea.KeyMsg{
			Type:  tea.KeyCtrlD,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyRight,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		assert.True(t, bar.deleteConfirm.GetValue().(bool))
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyEsc,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")
		bar.Update(tea.KeyMsg{
			Type:  tea.KeyCtrlD,
			Runes: []rune{},
			Alt:   false,
			Paste: false,
		}, "topic.a")

		assert.False(t, bar.deleteConfirm.GetValue().(bool))
	})
}
