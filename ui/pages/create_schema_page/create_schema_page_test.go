package create_schema_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/sradmin"
	"ktea/tests"
	"ktea/tests/keys"
	"ktea/ui"
	"ktea/ui/pages/nav"
	"testing"
)

type MockSubjectCreator struct{}

func (m *MockSubjectCreator) CreateSchema(details sradmin.SubjectCreationDetails) tea.Msg {
	return details
}

func TestCreateSubjectsPage(t *testing.T) {
	t.Run("Create schema", func(t *testing.T) {
		subjectPage, _ := New(&MockSubjectCreator{}, ui.TestKontext)
		// initialize form
		subjectPage.View(ui.TestKontext, ui.TestRenderer)

		keys.UpdateKeys(subjectPage, "subject")
		cmd := subjectPage.Update(keys.Key(tea.KeyEnter))
		// next field
		subjectPage.Update(cmd())

		keys.UpdateKeys(subjectPage, "{\"type\":\"string\"}")
		msgs := keys.Submit(subjectPage)

		assert.Len(t, msgs, 1)
		assert.IsType(t, sradmin.SubjectCreationDetails{}, msgs[0])
		assert.Equal(t, sradmin.SubjectCreationDetails{
			Subject: "subject",
			Schema:  "{\"type\":\"string\"}",
		}, msgs[0])

		t.Run("Create another schema", func(t *testing.T) {
			subjectPage.Update(sradmin.SchemaCreatedMsg{})
			// re-initialize form
			subjectPage.View(ui.TestKontext, ui.TestRenderer)

			keys.UpdateKeys(subjectPage, "subject")
			cmd := subjectPage.Update(keys.Key(tea.KeyEnter))
			// next field
			subjectPage.Update(cmd())

			keys.UpdateKeys(subjectPage, "{\"type\":\"string\"}")
			msgs = keys.Submit(subjectPage)

			assert.Len(t, msgs, 1)
			assert.Contains(t, msgs, sradmin.SubjectCreationDetails{
				Subject: "subject",
				Schema:  "{\"type\":\"string\"}",
			})
		})
	})

	t.Run("Unable to go back when schema is being created", func(t *testing.T) {
		subjectPage, _ := New(&MockSubjectCreator{}, ui.TestKontext)
		// initialize form
		subjectPage.View(ui.TestKontext, ui.TestRenderer)

		keys.UpdateKeys(subjectPage, "subject")
		cmd := subjectPage.Update(keys.Key(tea.KeyEnter))
		// next field
		subjectPage.Update(cmd())

		keys.UpdateKeys(subjectPage, "{\"type\":\"string\"}")
		keys.Submit(subjectPage)

		cmds := subjectPage.Update(keys.Key(tea.KeyEsc))

		assert.Nil(t, cmds)
	})

	t.Run("Esc goes back not refreshing when no schemas were created", func(t *testing.T) {
		subjectPage, _ := New(&MockSubjectCreator{}, ui.TestKontext)
		// initialize form
		subjectPage.View(ui.TestKontext, ui.TestRenderer)

		cmds := subjectPage.Update(keys.Key(tea.KeyEsc))
		msgs := tests.ExecuteBatchCmd(cmds)

		assert.Contains(t, msgs, nav.LoadSubjectsPageMsg{Refresh: false})
	})

	t.Run("Esc goes back refreshing when a schema has been created", func(t *testing.T) {
		subjectPage, _ := New(&MockSubjectCreator{}, ui.TestKontext)
		// initialize form
		subjectPage.View(ui.TestKontext, ui.TestRenderer)

		subjectPage.Update(sradmin.SchemaCreatedMsg{})

		cmds := subjectPage.Update(keys.Key(tea.KeyEsc))
		msgs := tests.ExecuteBatchCmd(cmds)

		assert.Contains(t, msgs, nav.LoadSubjectsPageMsg{Refresh: true})
	})

	t.Run("Subject is mandatory", func(t *testing.T) {
		subjectPage, _ := New(&MockSubjectCreator{}, ui.TestKontext)
		// initialize form
		subjectPage.View(ui.TestKontext, ui.TestRenderer)

		subjectPage.Update(keys.Key(tea.KeyEnter))

		render := subjectPage.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, "subject cannot be empty")
	})

	t.Run("Schema is mandatory", func(t *testing.T) {
		subjectPage, _ := New(&MockSubjectCreator{}, ui.TestKontext)
		// initialize form
		subjectPage.View(ui.TestKontext, ui.TestRenderer)

		keys.UpdateKeys(subjectPage, "subject")
		cmd := subjectPage.Update(keys.Key(tea.KeyEnter))
		// next field
		subjectPage.Update(cmd())

		subjectPage.Update(keys.Key(tea.KeyEnter))

		render := subjectPage.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, "schema cannot be empty")
	})
}
