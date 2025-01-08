package subjects_page

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"strconv"
	"strings"
)

type Model struct {
	table         table.Model
	rows          []table.Row
	cmdBar        *SubjectsCmdBar
	subjects      []sradmin.Subject
	tableFocussed bool
	lister        sradmin.SubjectLister
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	cmdBarView := m.cmdBar.View(ktx, renderer)

	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{"Subject Name", int(float64(ktx.WindowWidth-5) * 0.9)},
		{"Version Count", int(float64(ktx.WindowWidth-5) * 0.1)},
	})
	m.table.SetRows(m.rows)

	var render string
	if m.tableFocussed {
		render = renderer.Render(styles.Table.Focus.Render(m.table.View()))
	} else {
		render = renderer.Render(styles.Table.Blur.Render(m.table.View()))
	}

	return ui.JoinVerticalSkipEmptyViews(cmdBarView, render)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	msg, cmd := m.cmdBar.Update(msg, m.SelectedSubject())
	m.tableFocussed = !m.cmdBar.IsFocussed()
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "f5":
			return m.lister.ListSubjects
		case "ctrl+n":
			return ui.PublishMsg(nav.LoadCreateSubjectPageMsg{})
		case "enter":
			return ui.PublishMsg(
				nav.LoadSchemaDetailsPageMsg{
					Subject: m.table.SelectedRow()[0],
				},
			)
		}
	case sradmin.SubjectListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.SubjectsListedMsg:
		m.subjects = msg.Subjects
	case sradmin.SubjectDeletionStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	}

	searchTerm := m.cmdBar.GetSearchTerm()
	m.rows = make([]table.Row, 0)
	for _, subject := range m.subjects {
		if searchTerm != "" {
			if strings.Contains(strings.ToUpper(subject.Name), strings.ToUpper(searchTerm)) {
				m.rows = append(m.rows, table.Row{subject.Name, strconv.Itoa(len(subject.Versions))})
			}
		} else {
			m.rows = append(m.rows, table.Row{subject.Name, strconv.Itoa(len(subject.Versions))})
		}
	}

	t, cmd := m.table.Update(msg)
	m.table = t
	cmds = append(cmds, cmd)

	if m.cmdBar.HasSearchedAtLeastOneChar() {
		m.table.GotoTop()
	}

	return tea.Batch(cmds...)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	shortcuts := m.cmdBar.Shortcuts()
	if shortcuts == nil {
		return []statusbar.Shortcut{
			{
				Name:       "search",
				Keybinding: "/",
			},
			{
				Name:       "register",
				Keybinding: "C-n",
			},
			{
				Name:       "delete",
				Keybinding: "C-d",
			},
			{
				Name:       "refresh",
				Keybinding: "F5",
			},
		}
	} else {
		return shortcuts
	}
}

func (m *Model) SelectedSubject() sradmin.Subject {
	selectedRow := m.table.SelectedRow()
	var selectedTopic sradmin.Subject
	if selectedRow != nil {
		return m.subjects[m.table.Cursor()]
	}
	return selectedTopic
}

func (m *Model) Title() string {
	return "Subjects"
}

func New(lister sradmin.SubjectLister, deleter sradmin.SubjectDeleter) (*Model, tea.Cmd) {
	return &Model{
		cmdBar:        NewCmdBar(deleter),
		table:         ktable.NewDefaultTable(),
		tableFocussed: true,
		lister:        lister,
	}, lister.ListSubjects
}
