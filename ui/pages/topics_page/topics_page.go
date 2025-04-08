package topics_page

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/dustin/go-humanize"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type Model struct {
	topics        []kadmin.ListedTopic
	table         table.Model
	shortcuts     []statusbar.Shortcut
	cmdBar        *cmdbar.TableCmdsBar[string]
	rows          []table.Row
	lister        kadmin.TopicLister
	ctx           context.Context
	tableFocussed bool
	sort          cmdbar.SortLabel
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	cmdBarView := m.cmdBar.View(ktx, renderer)
	views = append(views, cmdBarView)

	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetWidth(ktx.WindowWidth - 3)
	m.table.SetColumns([]table.Column{
		{m.columnTitle("Name"), int(float64(ktx.WindowWidth-10) * 0.7)},
		{m.columnTitle("Partitions"), int(float64(ktx.WindowWidth-10) * 0.1)},
		{m.columnTitle("Replicas"), int(float64(ktx.WindowWidth-10) * 0.1)},
		{m.columnTitle("~ Record Count"), int(float64(ktx.WindowWidth-10) * 0.1)},
	})
	m.table.SetRows(m.rows)

	var tableView string
	if m.tableFocussed {
		// Apply focus style
		styledTable := renderer.RenderWithStyle(m.table.View(), styles.Table.Focus)

		embeddedText := map[styles.BorderPosition]string{
			styles.TopMiddleBorder: lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorPink)).
				Bold(true).
				Render(fmt.Sprintf("Total Topics: %d", len(m.rows))),
		}
		tableView = styles.Borderize(styledTable, true, embeddedText)
	} else {
		// Regular border approach for unfocused state
		styledTable := renderer.RenderWithStyle(m.table.View(), styles.Table.Blur)

		embeddedText := map[styles.BorderPosition]string{
			styles.TopMiddleBorder: lipgloss.NewStyle().
				Foreground(lipgloss.Color(styles.ColorPink)).
				Bold(true).
				Render(fmt.Sprintf("Total Topics: %d", len(m.rows))),
		}
		tableView = styles.Borderize(styledTable, false, embeddedText)
	}

	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

func (m *Model) columnTitle(title string) string {
	if m.sort.Label == title {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorPink)).
			Bold(true).
			Render(m.sort.Direction.String()) + " " + title
	}
	return title
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	cmds := make([]tea.Cmd, 2)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.topics == nil {
			return nil
		}
		switch msg.String() {
		case "ctrl+n":
			return ui.PublishMsg(nav.LoadCreateTopicPageMsg{})
		case "ctrl+o":
			return ui.PublishMsg(nav.LoadTopicConfigPageMsg{})
		case "ctrl+p":
			return ui.PublishMsg(nav.LoadPublishPageMsg{Topic: m.SelectedTopic()})
		case "f5":
			m.topics = nil
			return m.lister.ListTopics
		case "ctrl+l":
			return ui.PublishMsg(nav.LoadLiveConsumePageMsg{Topic: m.SelectedTopic()})
		case "enter":
			// only accept enter when the table is focussed
			if !m.cmdBar.IsFocussed() {
				// TODO ignore enter when there are no topics loaded
				return ui.PublishMsg(nav.LoadConsumptionFormPageMsg{
					Topic: m.SelectedTopic(),
				})
			}
		}
	case spinner.TickMsg:
		selectedTopic := m.SelectedTopicName()
		_, c := m.cmdBar.Update(msg, &selectedTopic)
		if c != nil {
			cmds = append(cmds, c)
		}
	case kadmin.TopicDeletionStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.TopicListingStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.TopicListedMsg:
		log.Debug("Topics listed")
		m.topics = msg.Topics
	case kadmin.TopicDeletedMsg:
		m.topics = slices.DeleteFunc(
			m.topics,
			func(t kadmin.ListedTopic) bool { return msg.TopicName == t.Name },
		)
	}

	name := m.SelectedTopicName()
	msg, cmd := m.cmdBar.Update(msg, &name)
	m.tableFocussed = !m.cmdBar.IsFocussed()
	cmds = append(cmds, cmd)

	m.rows = m.createRows()

	// make sure table navigation is off when the cmdbar is focussed
	if !m.cmdBar.IsFocussed() {
		t, cmd := m.table.Update(msg)
		m.table = t
		cmds = append(cmds, cmd)
	}

	if m.cmdBar.HasSearchedAtLeastOneChar() {
		m.table.GotoTop()
	}
	return tea.Batch(cmds...)
}

func (m *Model) createRows() []table.Row {
	var rows []table.Row
	for _, topic := range m.topics {
		if m.cmdBar.GetSearchTerm() != "" {
			if strings.Contains(strings.ToLower(topic.Name), strings.ToLower(m.cmdBar.GetSearchTerm())) {
				rows = append(
					rows,
					table.Row{
						topic.Name,
						strconv.Itoa(topic.PartitionCount),
						strconv.Itoa(topic.Replicas),
						humanize.Comma(topic.RecordCount),
					},
				)
			}
		} else {
			rows = append(
				rows,
				table.Row{
					topic.Name,
					strconv.Itoa(topic.PartitionCount),
					strconv.Itoa(topic.Replicas),
					humanize.Comma(topic.RecordCount),
				},
			)
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		switch m.sort.Label {
		case "Name":
			if m.sort.Direction == cmdbar.Asc {
				return rows[i][0] < rows[j][0]
			}
			return rows[i][0] > rows[j][0]
		case "Partitions":
			partitionI, _ := strconv.Atoi(rows[i][1])
			partitionJ, _ := strconv.Atoi(rows[j][1])
			if m.sort.Direction == cmdbar.Asc {
				return partitionI < partitionJ
			}
			return partitionI > partitionJ
		case "Replicas":
			replicasI, _ := strconv.Atoi(rows[i][2])
			replicasJ, _ := strconv.Atoi(rows[j][2])
			if m.sort.Direction == cmdbar.Asc {
				return replicasI < replicasJ
			}
			return replicasI > replicasJ
		case "~ Record Count":
			countI, _ := strconv.Atoi(strings.ReplaceAll(rows[i][3], ",", ""))
			countJ, _ := strconv.Atoi(strings.ReplaceAll(rows[j][3], ",", ""))
			if m.sort.Direction == cmdbar.Asc {
				return countI < countJ
			}
			return countI > countJ
		default:
			panic(fmt.Sprintf("unexpected sort label: %s", m.sort.Label))
		}
	})
	return rows
}

func (m *Model) SelectedTopic() *kadmin.ListedTopic {
	selectedTopic := m.SelectedTopicName()
	for _, t := range m.topics {
		if t.Name == selectedTopic {
			return &t
		}
	}
	panic("selected topic not found")
}

func (m *Model) SelectedTopicName() string {
	selectedRow := m.table.SelectedRow()
	var selectedTopic string
	if selectedRow != nil {
		selectedTopic = selectedRow[0]
	}
	return selectedTopic
}

func (m *Model) Title() string {
	return "Topics"
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return m.shortcuts
}

func (m *Model) Refresh() tea.Cmd {
	m.topics = nil
	return m.lister.ListTopics
}

func New(topicDeleter kadmin.TopicDeleter, lister kadmin.TopicLister) (*Model, tea.Cmd) {
	var m = Model{}
	m.shortcuts = []statusbar.Shortcut{
		{"Consume", "enter"},
		{"Live Consume", "C-l"},
		{"Search", "/"},
		{"Publish", "C-p"},
		{"Create", "C-n"},
		{"Delete", "F2"},
		{"Configs", "C-o"},
		{"Delete", "F2"},
		{"Sort", "F3"},
		{"Refresh", "F5"},
	}

	m.table = ktable.NewDefaultTable()

	m.table.SetColumns([]table.Column{
		{"Name", 1},
		{"Partitions", 1},
		{"Replicas", 1},
		{"In Sync Replicas", 1},
	})

	deleteMsgFunc := func(topic string) string {
		message := topic + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7571F9")).
			Bold(true).
			Render(" will be deleted permanently")
		return message
	}

	deleteFunc := func(topic string) tea.Cmd {
		return func() tea.Msg {
			return topicDeleter.DeleteTopic(topic)
		}
	}

	notifierCmdBar := cmdbar.NewNotifierCmdBar("topics-page")

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.TopicListingStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Loading Topics")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.TopicListedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.Idle()
			return true, m.AutoHideCmd("topics-page")
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.TopicListedErrorMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.ShowErrorMsg("Error listing Topics", msg.Err)
			return true, nil
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.TopicDeletedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.ShowSuccessMsg("Topic Deleted")
			return true, m.AutoHideCmd("topics-page")
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.TopicDeletionStartedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			cmd := m.SpinWithLoadingMsg("Deleting Topic")
			return true, cmd
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.TopicDeletionErrorMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.ShowErrorMsg("Error Deleting Topic", msg.Err)
			return true, m.AutoHideCmd("topics-page")
		},
	)

	sortByCmdBar := cmdbar.NewSortByCmdBar(
		[]cmdbar.SortLabel{
			{
				Label:     "Name",
				Direction: cmdbar.Asc,
			},
			{
				Label:     "Partitions",
				Direction: cmdbar.Desc,
			},
			{
				Label:     "Replicas",
				Direction: cmdbar.Desc,
			},
			{
				Label:     "~ Record Count",
				Direction: cmdbar.Desc,
			},
		},
		cmdbar.WithSortSelectedCallback(func(label cmdbar.SortLabel) {
			m.sort = label
		}),
	)
	m.cmdBar = cmdbar.NewTableCmdsBar[string](
		cmdbar.NewDeleteCmdBar(deleteMsgFunc, deleteFunc, nil),
		cmdbar.NewSearchCmdBar("Search topics by name"),
		notifierCmdBar,
		sortByCmdBar,
	)
	m.sort = sortByCmdBar.SortedBy()
	m.lister = lister
	return &m, lister.ListTopics
}
