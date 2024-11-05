package clusters_page

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"sort"
	"strings"
)

type Model struct {
	table  table.Model
	rows   []table.Row
	ktx    *kontext.ProgramKtx
	cmdBar *CmdBar
}

func (m *Model) Title() string {
	return "Clusters"
}

type ClusterSwitchedMsg struct {
	Cluster *config.Cluster
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	builder := strings.Builder{}
	builder.WriteString(m.cmdBar.View(ktx, renderer))
	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{"Active", int(float64(ktx.WindowWidth-5) * 0.05)},
		{"Name", int(float64(ktx.WindowWidth-5) * 0.95)},
	})
	m.table.SetRows(m.rows)
	builder.WriteString(styles.Table.Focus.Render(m.table.View()))
	return builder.String()
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Switch Cluster", "enter"},
		{"Edit", "C-e"},
		{"Delete", "C-d"},
		{"Create", "C-n"},
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !m.cmdBar.IsFocused() {
				cmds = append(cmds, func() tea.Msg {
					activeCluster := m.ktx.Config.SwitchCluster(*m.SelectedCluster())
					// immediately recreate the rows updating the active cluster
					m.rows = m.createRows()
					return ClusterSwitchedMsg{activeCluster}
				})
			}
		}
		selectedCluster := m.SelectedCluster()
		msgToPropagate, cmd := m.cmdBar.Update(msg, *selectedCluster)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if msgToPropagate == nil && cmd == nil {
			return nil
		}
	}
	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)
	cmds = append(cmds, tableCmd)
	m.rows = m.createRows()
	return tea.Batch(cmds...)
}

func (m *Model) SelectedCluster() *string {
	row := m.table.SelectedRow()
	if row == nil {
		return nil
	}
	return &row[1]
}

func (m *Model) createRows() []table.Row {
	var rows []table.Row
	for _, c := range m.ktx.Config.Clusters {
		var activeCell string
		if c.Active {
			activeCell = "X"
		} else {
			activeCell = ""
		}
		rows = append(rows, table.Row{activeCell, c.Name})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][1] < rows[j][1]
	})
	return rows
}

func New(ktx *kontext.ProgramKtx) (pages.Page, tea.Cmd) {
	model := Model{}
	model.ktx = ktx
	model.table = table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	model.rows = model.createRows()
	model.cmdBar = NewCmdBar(ktx)
	return &model, nil
}
