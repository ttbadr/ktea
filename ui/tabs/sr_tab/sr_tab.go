package sr_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/create_subjects_page"
	"ktea/ui/pages/navigation"
	"ktea/ui/pages/subjects_page"
)

type Model struct {
	active    navigation.Page
	statusbar *statusbar.Model
	ktx       *kontext.ProgramKtx
	creator   sradmin.SubjectCreator
	lister    sradmin.SubjectLister
	deleter   sradmin.SubjectDeleter
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	view := m.statusbar.View(ktx, renderer)
	return ui.JoinVerticalSkipEmptyViews(view, m.active.View(ktx, renderer))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg.(type) {
	case navigation.LoadCreateSubjectPageMsg:
		createPage, cmd := create_subjects_page.New(m.creator, m.ktx)
		cmds = append(cmds, cmd)
		m.active = createPage
	case navigation.LoadSubjectsPageMsg:
		subjectsPage, cmd := subjects_page.New(m.lister, m.deleter)
		cmds = append(cmds, cmd)
		m.active = subjectsPage
	}

	m.statusbar = statusbar.New(m.active)

	cmds = append(cmds, m.active.Update(msg))

	return tea.Batch(cmds...)
}

func New(lister sradmin.SubjectLister, creator sradmin.SubjectCreator, deleter sradmin.SubjectDeleter, ktx *kontext.ProgramKtx) (*Model, tea.Cmd) {
	subjectsPage, cmd := subjects_page.New(lister, deleter)
	model := Model{active: subjectsPage}
	model.statusbar = statusbar.New(subjectsPage)
	model.ktx = ktx
	model.creator = creator
	model.lister = lister
	model.deleter = deleter
	return &model, cmd
}
