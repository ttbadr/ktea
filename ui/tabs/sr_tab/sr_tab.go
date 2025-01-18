package sr_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"ktea/ui/pages/register_new_schema_page"
	"ktea/ui/pages/schema_details_page"
	"ktea/ui/pages/subjects_page"
)

type Model struct {
	active            nav.Page
	statusbar         *statusbar.Model
	ktx               *kontext.ProgramKtx
	schemaCreator     sradmin.SchemaCreator
	subjectLister     sradmin.SubjectLister
	subjectDeleter    sradmin.SubjectDeleter
	subjectsPage      *subjects_page.Model
	schemaDetailsPage *schema_details_page.Model
	schemaLister      sradmin.VersionLister
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	statusBarView := m.statusbar.View(ktx, renderer)
	return ui.JoinVerticalSkipEmptyViews(
		lipgloss.Top,
		statusBarView,
		m.active.View(ktx, renderer),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case sradmin.SubjectsListedMsg:
		return m.subjectsPage.Update(msg)
	case nav.LoadCreateSubjectPageMsg:
		createPage, cmd := register_new_schema_page.New(m.schemaCreator, m.ktx)
		cmds = append(cmds, cmd)
		m.active = createPage
	case nav.LoadSubjectsPageMsg:
		if m.subjectsPage == nil || msg.Refresh {
			var cmd tea.Cmd
			m.subjectsPage, cmd = subjects_page.New(m.subjectLister, m.subjectDeleter)
			cmds = append(cmds, cmd)
		}
		m.active = m.subjectsPage
	case nav.LoadSchemaDetailsPageMsg:
		var cmd tea.Cmd
		m.schemaDetailsPage, cmd = schema_details_page.New(m.schemaLister, msg.Subject)
		m.active = m.schemaDetailsPage
		cmds = append(cmds, cmd)
	}

	m.statusbar = statusbar.New(m.active)

	cmds = append(cmds, m.active.Update(msg))

	return tea.Batch(cmds...)
}

func New(
	subjectLister sradmin.SubjectLister,
	schemaLister sradmin.VersionLister,
	subjectCreator sradmin.SchemaCreator,
	subjectDeleter sradmin.SubjectDeleter,
	ktx *kontext.ProgramKtx,
) (*Model, tea.Cmd) {
	subjectsPage, cmd := subjects_page.New(subjectLister, subjectDeleter)
	model := Model{active: subjectsPage}
	model.subjectsPage = subjectsPage
	model.statusbar = statusbar.New(subjectsPage)
	model.ktx = ktx
	model.schemaCreator = subjectCreator
	model.subjectLister = subjectLister
	model.schemaLister = schemaLister
	model.subjectDeleter = subjectDeleter
	return &model, cmd
}
