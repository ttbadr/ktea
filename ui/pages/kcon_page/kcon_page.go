package kcon_page

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"ktea/config"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"reflect"
	"sort"
	"strings"
)

type Model struct {
	connectors    *kcadmin.Connectors
	table         *table.Model
	cmdBar        *cmdbar.TableCmdsBar[string]
	rows          []table.Row
	sort          cmdbar.SortLabel
	lister        kcadmin.ConnectorLister
	border        *border.Model
	sortByCmdBar  *cmdbar.SortByCmdBar
	navBack       ui.NavBack
	connectorName string
	state
}

type state int

type LoadPage func(c config.KafkaConnectConfig) tea.Cmd

const (
	loading state = iota
	loaded
	errored
)

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	cmdBarView := m.cmdBar.View(ktx, renderer)

	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{m.sortByCmdBar.PrefixSortIcon("Name"), int(float64(ktx.WindowWidth-5) * 0.7)},
		{m.sortByCmdBar.PrefixSortIcon("Status"), int(float64(ktx.WindowWidth-5) * 0.3)},
	})
	m.table.SetRows(m.rows)
	m.table.SetHeight(ktx.AvailableHeight - 2)

	tableView := m.border.View(m.table.View())
	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if !m.cmdBar.IsFocussed() {
				return m.navBack()
			}
		case tea.KeyF5:
			if m.state == loading {
				log.Debug("not refreshing connectors due to loading state")
				return nil
			}
			m.connectors = nil
			m.rows = m.createRows()
			m.state = loading
			log.Debug("refreshing connectors")
			return m.lister.ListActiveConnectors
		}
	case kcadmin.ConnectorListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case kcadmin.ConnectorsListedMsg:
		m.connectors = &msg.Connectors
		m.state = loaded
	case kcadmin.ConnectorDeletionStartedMsg:
		m.state = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case kcadmin.ConnectorDeletedMsg:
		m.state = loaded
		delete(*m.connectors, msg.Name)
	}

	var (
		selection = m.selectedConnector()
		cmd       tea.Cmd
	)
	_, cmd = m.cmdBar.Update(msg, &selection)
	m.border.Focused = !m.cmdBar.IsFocussed()
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// make sure table navigation is off when the cmdbar is focussed
	if !m.cmdBar.IsFocussed() {
		t, cmd := m.table.Update(msg)
		m.table = &t
		cmds = append(cmds, cmd)
	}

	m.rows = m.createRows()

	return tea.Batch(cmds...)
}

func (m *Model) selectedConnector() string {
	selectedRow := m.table.SelectedRow()
	var selectedConnector string
	if selectedRow != nil {
		selectedConnector = selectedRow[0]
	}
	return selectedConnector
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	if m.cmdBar.IsFocussed() {
		shortCuts := m.cmdBar.Shortcuts()
		if shortCuts != nil {
			return shortCuts
		}
	}

	return []statusbar.Shortcut{
		{"Search", "/"},
		{"Delete", "F2"},
		{"Sort", "F3"},
		{"Refresh", "F5"},
	}
}

func (m *Model) Title() string {
	return "Kafka Connect Clusters / " + m.connectorName
}

func (m *Model) createRows() []table.Row {
	var rows []table.Row
	if m.connectors != nil {
		for k, c := range *m.connectors {
			if m.cmdBar.GetSearchTerm() != "" {
				if strings.Contains(strings.ToLower(k), strings.ToLower(m.cmdBar.GetSearchTerm())) {
					rows = append(rows, table.Row{k, c.Status.Connector.State})
				}
			} else {
				rows = append(rows, table.Row{k, c.Status.Connector.State})
			}
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		switch m.sortByCmdBar.SortedBy().Label {
		case "Name":
			if m.sortByCmdBar.SortedBy().Direction == cmdbar.Asc {
				return rows[i][0] < rows[j][0]
			}
			return rows[i][0] > rows[j][0]
		default:
			panic(fmt.Sprintf("unexpected sort label: %s", m.sortByCmdBar.SortedBy().Label))
		}
	})

	return rows
}

func New(
	navBack ui.NavBack,
	lister kcadmin.ConnectorLister,
	deleter kcadmin.ConnectorDeleter,
	name string,
) (*Model, tea.Cmd) {
	m := Model{}
	m.connectorName = name
	m.navBack = navBack
	m.lister = lister
	m.state = loading

	m.border = border.New(
		border.WithTitleFunc(func() string {
			if m.connectors == nil {
				return styles.BorderTopTitle("Loading Connectors", "", false)
			}
			return styles.BorderTopTitle("Total Connectors", fmt.Sprintf("%d/%d", len(m.rows), len(*m.connectors)), true)
		}))
	m.border.Focused = false

	t := ktable.NewDefaultTable()
	m.table = &t

	sortByCmdBar := cmdbar.NewSortByCmdBar(
		[]cmdbar.SortLabel{
			{
				Label:     "Name",
				Direction: cmdbar.Asc,
			},
		},
		cmdbar.WithSortSelectedCallback(func(label cmdbar.SortLabel) {
			m.sort = label
		}),
	)
	m.sortByCmdBar = sortByCmdBar

	notifierCmdBar := cmdbar.NewNotifierCmdBar("kcons-page")
	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorListingStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Loading connectors")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorsListedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.Idle()
			return true, nil
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorListingErrMsg,
			model *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := model.ShowError(fmt.Errorf("unable to reach connect cluster: %v", msg.Err))
			m.state = errored
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorDeletionStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Deleting Connector")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorDeletedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.ShowSuccessMsg("Connector deleted")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kcadmin.ConnectorDeletionErrMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.ShowError(msg.Err)
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg ui.RegainedFocusMsg,
			model *notifier.Model,
		) (bool, tea.Cmd) {
			if model.State == notifier.Spinning {
				cmd := model.SpinWithLoadingMsg("Loading connectors")
				return true, cmd
			}
			if model.State == notifier.Err {
				return true, nil
			}
			return false, nil
		},
	)

	deleteMsgFunc := func(connector string) string {
		message := connector + lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorIndigo)).
			Bold(true).
			Render(" will be deleted permanently")
		return message
	}

	deleteFunc := func(name string) tea.Cmd {
		return func() tea.Msg {
			return deleter.DeleteConnector(name)
		}
	}

	m.cmdBar = cmdbar.NewTableCmdsBar(
		cmdbar.NewDeleteCmdBar[string](deleteMsgFunc, deleteFunc, nil),
		cmdbar.NewSearchCmdBar("Search connectors by name"),
		notifierCmdBar,
		sortByCmdBar)

	m.sort = sortByCmdBar.SortedBy()

	return &m, lister.ListActiveConnectors
}
