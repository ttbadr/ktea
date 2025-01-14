package sr_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/create_schema_page"
	"ktea/ui/pages/nav"
	"ktea/ui/pages/subjects_page"
)

type Model struct {
	active       nav.Page
	statusbar    *statusbar.Model
	ktx          *kontext.ProgramKtx
	creator      sradmin.SubjectCreator
	lister       sradmin.SubjectLister
	deleter      sradmin.SubjectDeleter
	subjectsPage *subjects_page.Model
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	view := m.statusbar.View(ktx, renderer)
	return ui.JoinVerticalSkipEmptyViews(lipgloss.Top, view, m.active.View(ktx, renderer))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg.(type) {
	case sradmin.SubjectsListedMsg:
		return m.subjectsPage.Update(msg)
	case nav.LoadCreateSubjectPageMsg:
		createPage, cmd := create_schema_page.New(m.creator, m.ktx)
		cmds = append(cmds, cmd)
		m.active = createPage
	case nav.LoadSubjectsPageMsg:
		var cmd tea.Cmd
		m.subjectsPage, cmd = subjects_page.New(m.lister, m.deleter)
		m.active = m.subjectsPage
		cmds = append(cmds, cmd)
	}

	m.statusbar = statusbar.New(m.active)

	cmds = append(cmds, m.active.Update(msg))

	return tea.Batch(cmds...)
}

func New(lister sradmin.SubjectLister, creator sradmin.SubjectCreator, deleter sradmin.SubjectDeleter, ktx *kontext.ProgramKtx) (*Model, tea.Cmd) {
	subjectsPage, cmd := subjects_page.New(lister, deleter)
	model := Model{active: subjectsPage}
	model.subjectsPage = subjectsPage
	model.statusbar = statusbar.New(subjectsPage)
	model.ktx = ktx
	model.creator = creator
	model.lister = lister
	model.deleter = deleter
	return &model, cmd
}
