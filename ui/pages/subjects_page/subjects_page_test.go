package subjects_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/sradmin"
	"ktea/tests"
	"ktea/tests/keys"
	"ktea/ui"
	"math/rand"
	"strings"
	"testing"
	"time"
)

type MockSubjectsLister struct {
	subjectListingStartedMsg sradmin.SubjectListingStartedMsg
}

func (m *MockSubjectsLister) ListSubjects() tea.Msg {
	return nil
}

type MockSubjectsDeleter struct {
	deletionResultMsg tea.Msg
}

type DeletedSubjectMsg struct {
	Subject string
	Version int
}

func (m *MockSubjectsDeleter) DeleteSubject(subject string) tea.Msg {
	return m.deletionResultMsg
}

func TestSubjectsPage(t *testing.T) {

	t.Run("Render listed subjects and number of versions", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockSubjectsDeleter{},
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
			&MockSubjectsLister{},
			&MockSubjectsDeleter{},
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

	t.Run("When deletion spinner active do not allow other cmdbars to activate", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockSubjectsDeleter{},
		)
		subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

		subjectsPage.Update(keys.Key('/'))

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, " ⏳ Deleting Subject")
	})

	t.Run("Order subjects by name desc", func(t *testing.T) {
		deleter := MockSubjectsDeleter{}
		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&deleter,
		)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 100; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		shuffle(subjects)

		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		render := subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)

		subject1Idx := strings.Index(render, "subject1")
		subject2Idx := strings.Index(render, "subject2")
		subject50Idx := strings.Index(render, "subject50")
		subject88Idx := strings.Index(render, "subject88")
		assert.Less(t, subject1Idx, subject2Idx, "subject2 came before subject1")
		assert.Less(t, subject2Idx, subject50Idx, "subject50 came before subject2")
		assert.Less(t, subject50Idx, subject88Idx, "subject88 came before subject50")
	})

	t.Run("Delete subject", func(t *testing.T) {
		deleter := MockSubjectsDeleter{}
		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&deleter,
		)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 100; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		// render so the table's first row is selected
		render := subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)
		assert.NotRegexp(t, "┃ 🗑️  subject1 will be deleted permanently\\W+Delete!\\W+Cancel.", render)

		t.Run("F2 triggers subject delete", func(t *testing.T) {
			subjectsPage.Update(keys.Key(tea.KeyDown))
			subjectsPage.Update(keys.Key(tea.KeyF2))

			render = subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)

			assert.Regexp(t, "┃ 🗑️  subject1 will be deleted permanently\\W+Delete!\\W+Cancel.", render)
		})

		t.Run("Delete after searching from selective list", func(t *testing.T) {
			subjectsPage.Update(keys.Key('/'))
			keys.UpdateKeys(subjectsPage, "1")
			subjectsPage.Update(keys.Key(tea.KeyEnter))
			subjectsPage.Update(keys.Key(tea.KeyDown))
			subjectsPage.Update(keys.Key(tea.KeyDown))
			subjectsPage.Update(keys.Key(tea.KeyF2))

			render = subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)

			assert.Regexp(t, "┃ 🗑️  subject11 will be deleted permanently\\W+Delete!\\W+Cancel.", render)

			// reset search
			subjectsPage.Update(keys.Key('/'))
			subjectsPage.Update(keys.Key(tea.KeyEsc))
		})

		t.Run("Enter effectively deletes the subject", func(t *testing.T) {
			deleter.deletionResultMsg = DeletedSubjectMsg{"subject1", 1}

			subjectsPage.Update(keys.Key(tea.KeyF2))
			subjectsPage.Update(keys.Key('d'))
			cmds := subjectsPage.Update(keys.Key(tea.KeyEnter))
			msgs := tests.ExecuteBatchCmd(cmds)

			assert.Contains(t, msgs, DeletedSubjectMsg{"subject1", 1})
		})

		t.Run("Display error when deletion fails", func(t *testing.T) {
			deleter.deletionResultMsg = sradmin.SubjectDeletionStartedMsg{}

			subjectsPage.Update(keys.Key(tea.KeyF2))
			subjectsPage.Update(keys.Key('d'))
			cmds := subjectsPage.Update(keys.Key(tea.KeyEnter))

			for _, msg := range tests.ExecuteBatchCmd(cmds) {
				subjectsPage.Update(msg)
			}
			subjectsPage.Update(sradmin.SubjectDeletionErrorMsg{
				Err: fmt.Errorf("unable to delete subject"),
			})

			render = subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)

			assert.Regexp(t, "unable to delete subject", render)

			t.Run("When deletion failure msg visible do allow other cmdbars to activate", func(t *testing.T) {
				subjectsPage.Update(keys.Key('/'))

				render = subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)

				assert.NotContains(t, render, "Failed to delete subject: unable to delete subject")
				assert.Contains(t, render, "> Search subject by name")
			})
		})

		t.Run("When deletion started show spinning indicator", func(t *testing.T) {

			subjectsPage, _ := New(
				&MockSubjectsLister{},
				&MockSubjectsDeleter{},
			)
			subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

			render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

			assert.Contains(t, render, " ⏳ Deleting Subject")
		})
	})

	t.Run("When listing started show spinning indicator", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockSubjectsDeleter{},
		)
		subjectsPage.Update(sradmin.SubjectListingStartedMsg{})

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, " ⏳ Loading subjects")
	})
}

func shuffle[T any](slice []T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(slice); n > 0; n-- {
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
	}
}
