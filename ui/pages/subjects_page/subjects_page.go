package subjects_page

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"sort"
	"strconv"
	"strings"
)

type state int

const (
	initialized     state = 0
	subjectsLoaded  state = 1
	loading         state = 2
	noSubjectsFound state = 3
	deleting        state = 4
)

type Model struct {
	table            table.Model
	rows             []table.Row
	cmdBar           *SubjectsCmdBar
	subjects         []sradmin.Subject
	renderedSubjects []sradmin.Subject
	tableFocussed    bool
	lister           sradmin.SubjectLister
	state            state
	// when last subject in table is deleted no subject is focussed anymore
	deletedLast bool
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.state == noSubjectsFound {
		return lipgloss.NewStyle().
			Width(ktx.WindowWidth - 2).
			Height(ktx.AvailableHeight - 2).
			AlignVertical(lipgloss.Center).
			AlignHorizontal(lipgloss.Center).
			BorderStyle(lipgloss.RoundedBorder()).
			Render("No Subjects Found")
	}

	cmdBarView := m.cmdBar.View(ktx, renderer)

	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{"Subject Name", int(float64(ktx.WindowWidth-5) * 0.9)},
		{"Version Count", int(float64(ktx.WindowWidth-5) * 0.1)},
	})
	m.table.SetRows(m.rows)

	if m.deletedLast && m.table.SelectedRow() == nil {
		m.table.GotoBottom()
		m.deletedLast = false
	}
	if m.table.SelectedRow() == nil {
		m.table.GotoTop()
	}

	var tableView string
	if m.tableFocussed {
		tableView = renderer.Render(styles.Table.Focus.Render(m.table.View()))
	} else {
		tableView = renderer.Render(styles.Table.Blur.Render(m.table.View()))
	}

	return ui.JoinVerticalSkipEmptyViews(lipgloss.Top, cmdBarView, tableView)
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
			m.state = loading
			m.subjects = nil
			return m.lister.ListSubjects
		case "ctrl+n":
			if m.state != loading && m.state != deleting {
				return ui.PublishMsg(nav.LoadCreateSubjectPageMsg{})
			}
		case "enter":
			// ignore enter when there are no schemas loaded
			if m.state == subjectsLoaded && len(m.subjects) > 0 {
				return ui.PublishMsg(
					nav.LoadSchemaDetailsPageMsg{
						Subject: *m.SelectedSubject(),
					},
				)
			}
		}
	case sradmin.SubjectListingStartedMsg:
		m.state = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.SubjectsListedMsg:
		if len(msg.Subjects) > 0 {
			m.state = subjectsLoaded
			m.subjects = msg.Subjects
		} else {
			m.state = noSubjectsFound
		}
	case sradmin.SubjectDeletionStartedMsg:
		m.state = deleting
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.SubjectDeletedMsg:
		// set state back to loaded after removing the deleted subject
		m.state = subjectsLoaded
		for i, subject := range m.subjects {
			if subject.Name == msg.SubjectName {
				if i == len(m.subjects)-1 {
					m.deletedLast = true
				}
				m.subjects = append(m.subjects[:i], m.subjects[i+1:]...)
			}
		}
		if len(m.subjects) == 0 {
			m.state = noSubjectsFound
		}
	}

	subjects := m.filterSubjectsBySearchTerm()
	subjects = m.sortSubjects(subjects)
	m.renderedSubjects = subjects
	m.rows = m.createRows(subjects)

	t, cmd := m.table.Update(msg)
	m.table = t
	cmds = append(cmds, cmd)

	if m.cmdBar.HasSearchedAtLeastOneChar() {
		m.table.GotoTop()
	}

	return tea.Batch(cmds...)
}

func (m *Model) sortSubjects(subjects []sradmin.Subject) []sradmin.Subject {
	sort.Slice(subjects, func(i int, y int) bool {
		return subjects[i].Name < subjects[y].Name
	})
	return subjects
}

func (m *Model) createRows(subjects []sradmin.Subject) []table.Row {
	var rows []table.Row
	for _, subject := range subjects {
		rows = append(rows, table.Row{
			subject.Name,
			strconv.Itoa(len(subject.Versions)),
		})
	}
	return rows
}

func (m *Model) filterSubjectsBySearchTerm() []sradmin.Subject {
	var subjects []sradmin.Subject
	searchTerm := m.cmdBar.GetSearchTerm()
	for _, subject := range m.subjects {
		if searchTerm != "" {
			if strings.Contains(strings.ToUpper(subject.Name), strings.ToUpper(searchTerm)) {
				subjects = append(subjects, subject)
			}
		} else {
			subjects = append(subjects, subject)
		}
	}
	return subjects
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	shortcuts := m.cmdBar.Shortcuts()
	if shortcuts == nil {
		return []statusbar.Shortcut{
			{
				Name:       "Search",
				Keybinding: "/",
			},
			{
				Name:       "Delete",
				Keybinding: "F2",
			},
			{
				Name:       "Register New Schema",
				Keybinding: "C-n",
			},
			//{
			//	Name:       "Evolve Selected Schema",
			//	Keybinding: "C-e",
			//},
			{
				Name:       "Refresh",
				Keybinding: "F5",
			},
		}
	} else {
		return shortcuts
	}
}

func (m *Model) SelectedSubject() *sradmin.Subject {
	if len(m.renderedSubjects) > 0 {
		selectedRow := m.table.SelectedRow()
		if selectedRow != nil {
			return &m.renderedSubjects[m.table.Cursor()]
		}
		return nil
	}
	return nil
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
		state:         initialized,
	}, lister.ListSubjects
}
