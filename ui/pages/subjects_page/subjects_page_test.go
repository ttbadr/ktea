package subjects_page

import (
	"fmt"
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

	t.Run("Render listed subjects and number of versions", func(t *testing.T) {

		subjectsPage, _ := New(
			MockSubjectsLister{},
			MockSubjectsDeleter{},
		)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 10; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		for i := 0; i < 10; i++ {
			assert.Regexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
		}
	})

	t.Run("Remove delete subject from table", func(t *testing.T) {

		subjectsPage, _ := New(
			MockSubjectsLister{},
			MockSubjectsDeleter{},
		)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 10; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		subjectsPage.Update(sradmin.SubjectDeletedMsg{SubjectName: subjects[4].Name})

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		assert.NotRegexp(t, "subject4\\W+5", render)
	})

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
