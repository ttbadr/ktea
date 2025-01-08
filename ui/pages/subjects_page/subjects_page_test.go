package subjects_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	sradmin "ktea/sradmin"
	"ktea/ui"
	"testing"
)

type MockSubjectsLister struct {
	subjectListingStartedMsg sradmin.SubjectListingStartedMsg
}

func (m MockSubjectsLister) ListSubjects() tea.Msg {
	return nil
}

type MockSubjectsDeleter struct {
}

func (m MockSubjectsDeleter) DeleteSubject(subject string, version int) tea.Msg {
	return nil
}

func TestSubjectsPage(t *testing.T) {
	t.Run("When listing started show spinning indicator", func(t *testing.T) {

		subjectsPage, _ := New(
			MockSubjectsLister{},
			MockSubjectsDeleter{},
		)
		subjectsPage.Update(sradmin.SubjectListingStartedMsg{})

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, " ⏳ Loading subjects")
	})

	t.Run("When deletion started show spinning indicator", func(t *testing.T) {

		subjectsPage, _ := New(
			MockSubjectsLister{},
			MockSubjectsDeleter{},
		)
		subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, " ⏳ Deleting Subject")
	})
}
