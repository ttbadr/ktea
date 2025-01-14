package cgroups_page

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"sort"
	"strconv"
)

type Model struct {
	lister kadmin.CGroupLister
	cmdBar *CmdBar
	table  table.Model
	groups []*kadmin.ConsumerGroup
	rows   []table.Row
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Search", "/"},
		{"View", "enter"},
		{"Delete", "C-d"},
		{"Refresh", "F5"},
	}
}

func (m *Model) Title() string {
	return "Consumer Groups"
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string

	cmdBarView := m.cmdBar.View(ktx, renderer)
	if cmdBarView != "" {
		views = append(views, cmdBarView)
	}

	m.table.SetHeight(ktx.AvailableHeight - lipgloss.Height(cmdBarView))
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{"Consumer Group", int(float64(ktx.WindowWidth-5) * 0.7)},
		{"Members", int(float64(ktx.WindowWidth-5) * 0.3)},
	})
	m.table.SetRows(m.rows)
	views = append(views, styles.Table.Focus.Render(m.table.View()))

	return lipgloss.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return ui.PublishMsg(nav.LoadCGroupTopicsPageMsg{GroupName: m.table.SelectedRow()[0]})
		case "f5":
			return m.lister.ListConsumerGroups
		}
	case kadmin.ConsumerGroupsListedMsg:
		m.groups = msg.ConsumerGroups
		for _, group := range m.groups {
			m.rows = append(m.rows, table.Row{
				group.Name,
				strconv.Itoa(len(group.Members)),
			})
		}
		sort.SliceStable(m.rows, func(i, j int) bool {
			return m.rows[i][0] < m.rows[j][0]
		})
	}

	var cmd tea.Cmd

	cmd = m.cmdBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.table, cmd = m.table.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func New(lister kadmin.CGroupLister) (*Model, tea.Cmd) {
	m := &Model{}
	m.lister = lister
	m.table = table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	m.cmdBar = NewCmdBar()
	return m, m.lister.ListConsumerGroups
}
