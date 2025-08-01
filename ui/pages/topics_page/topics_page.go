package topics_page

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
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

const name = "topics-page"

type state int

const (
	stateRefreshing state = iota
	stateLoading
	stateLoaded
)

type Model struct {
	topics        []kadmin.ListedTopic
	table         table.Model
	shortcuts     []statusbar.Shortcut
	tcb           *cmdbar.TableCmdsBar[string]
	rows          []table.Row
	lister        kadmin.TopicLister
	ctx           context.Context
	tableFocussed bool
	state         state
	sortByCmdBar  *cmdbar.SortByCmdBar
	goToTop       bool
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	cmdBarView := m.tcb.View(ktx, renderer)
	views = append(views, cmdBarView)

	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{m.sortByCmdBar.PrefixSortIcon("Name"), int(float64(ktx.WindowWidth-7) * 0.6)},
		{m.sortByCmdBar.PrefixSortIcon("Partitions"), int(float64(ktx.WindowWidth-7) * 0.3)},
		{m.sortByCmdBar.PrefixSortIcon("Replicas"), int(float64(ktx.WindowWidth-7) * 0.1)},
	})
	m.table.SetRows(m.rows)
	m.table.SetHeight(ktx.AvailableHeight - 2)

	if m.table.SelectedRow() == nil && len(m.table.Rows()) > 0 {
		m.goToTop = true
	}

	if m.goToTop {
		m.table.GotoTop()
		m.goToTop = false
	}

	styledTable := renderer.RenderWithStyle(m.table.View(), styles.Table.Blur)

	embeddedText := map[styles.BorderPosition]styles.EmbeddedTextFunc{
		styles.TopMiddleBorder:    styles.EmbeddedBorderText("Total Topics", fmt.Sprintf("%d/%d", len(m.rows), len(m.topics))),
		styles.BottomMiddleBorder: styles.EmbeddedBorderText("Total Topics", fmt.Sprintf("%d/%d", len(m.rows), len(m.topics))),
	}

	tableView := styles.Borderize(styledTable, m.tableFocussed, embeddedText)

	return ui.JoinVertical(lipgloss.Top, cmdBarView, tableView)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {

	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	cmds := make([]tea.Cmd, 2)

	var cmd tea.Cmd
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+n":
			return ui.PublishMsg(nav.LoadCreateTopicPageMsg{})
		case "ctrl+o":
			if m.SelectedTopic() == nil {
				return nil
			}
			return ui.PublishMsg(nav.LoadTopicConfigPageMsg{})
		case "ctrl+p":
			if m.SelectedTopic() == nil {
				return nil
			}
			return ui.PublishMsg(nav.LoadPublishPageMsg{Topic: m.SelectedTopic()})
		case "f5":
			m.topics = nil
			m.state = stateRefreshing
			return m.lister.ListTopics
		case "L":
			if m.SelectedTopic() == nil {
				return nil
			}
			return ui.PublishMsg(nav.LoadLiveConsumePageMsg{Topic: m.SelectedTopic()})
		case "enter":
			// only accept enter when the table is focussed
			if !m.tcb.IsFocussed() {
				if m.SelectedTopic() != nil {
					return ui.PublishMsg(nav.LoadConsumptionFormPageMsg{
						Topic: m.SelectedTopic(),
					})
				}
			}
		}
	case spinner.TickMsg:
		selectedTopic := m.SelectedTopicName()
		_, c := m.tcb.Update(msg, &selectedTopic)
		if c != nil {
			cmds = append(cmds, c)
		}
	case kadmin.TopicDeletionStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.TopicListingStartedMsg:
		cmds = append(cmds, msg.AwaitTopicListCompletion)
	case kadmin.TopicsListedMsg:
		m.tcb.ResetSearch()
		m.topics = msg.Topics
		m.goToTop = true
		m.state = stateLoaded
	case kadmin.TopicDeletedMsg:
		m.topics = slices.DeleteFunc(
			m.topics,
			func(t kadmin.ListedTopic) bool { return msg.TopicName == t.Name },
		)
	}

	name := m.SelectedTopicName()
	msg, cmd = m.tcb.Update(msg, &name)
	m.tableFocussed = !m.tcb.IsFocussed()
	cmds = append(cmds, cmd)

	m.rows = m.createRows()

	// make sure table navigation is off when the cmdbar is focussed
	if !m.tcb.IsFocussed() {
		t, cmd := m.table.Update(msg)
		m.table = t
		cmds = append(cmds, cmd)
	}

	if m.tcb.HasSearchedAtLeastOneChar() {
		m.goToTop = true
	}

	return tea.Batch(cmds...)
}

func (m *Model) createRows() []table.Row {
	var rows []table.Row
	for _, topic := range m.topics {
		if m.tcb.GetSearchTerm() != "" {
			if strings.Contains(strings.ToLower(topic.Name), strings.ToLower(m.tcb.GetSearchTerm())) {
				rows = append(
					rows,
					table.Row{
						topic.Name,
						strconv.Itoa(topic.PartitionCount),
						strconv.Itoa(topic.Replicas),
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
				},
			)
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		switch m.sortByCmdBar.SortedBy().Label {
		case "Name":
			if m.sortByCmdBar.SortedBy().Direction == cmdbar.Asc {
				return rows[i][0] < rows[j][0]
			}
			return rows[i][0] > rows[j][0]
		case "Partitions":
			partitionI, _ := strconv.Atoi(rows[i][1])
			partitionJ, _ := strconv.Atoi(rows[j][1])
			if m.sortByCmdBar.SortedBy().Direction == cmdbar.Asc {
				return partitionI < partitionJ
			}
			return partitionI > partitionJ
		case "Replicas":
			replicasI, _ := strconv.Atoi(rows[i][2])
			replicasJ, _ := strconv.Atoi(rows[j][2])
			if m.sortByCmdBar.SortedBy().Direction == cmdbar.Asc {
				return replicasI < replicasJ
			}
			return replicasI > replicasJ
		case "~ Record Count":
			countI, _ := strconv.Atoi(strings.ReplaceAll(rows[i][3], ",", ""))
			countJ, _ := strconv.Atoi(strings.ReplaceAll(rows[j][3], ",", ""))
			if m.sortByCmdBar.SortedBy().Direction == cmdbar.Asc {
				return countI < countJ
			}
			return countI > countJ
		default:
			panic(fmt.Sprintf("unexpected sort label: %s", m.sortByCmdBar.SortedBy().Label))
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
	return nil
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
	if m.tcb.IsFocussed() {
		shortCuts := m.tcb.Shortcuts()
		if shortCuts != nil {
			return shortCuts
		}
	}
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
		{"Live Consume", "S-l"},
		{"Search", "/"},
		{"Produce", "C-p"},
		{"Create", "C-n"},
		{"Configs", "C-o"},
		{"Delete", "F2"},
		{"Sort", "F3"},
		{"Refresh", "F5"},
	}

	m.table = ktable.NewDefaultTable()

	deleteMsgFunc := func(topic string) string {
		message := topic + lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.ColorIndigo)).
			Bold(true).
			Render(" will be deleted permanently")
		return message
	}

	deleteFunc := func(topic string) tea.Cmd {
		return func() tea.Msg {
			return topicDeleter.DeleteTopic(topic)
		}
	}

	notifierCmdBar := cmdbar.NewNotifierCmdBar(name)

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
			msg ui.RegainedFocusMsg,
			model *notifier.Model,
		) (bool, tea.Cmd) {
			if m.state == stateRefreshing || m.state == stateLoading {
				log.Debug("skldfjkslfjsdlf//////////", m.state)
				cmd := model.SpinWithLoadingMsg("Loading Topics")
				return true, cmd
			}
			return false, nil
		},
	)

	cmdbar.WithMsgHandler(
		notifierCmdBar,
		func(
			msg kadmin.TopicsListedMsg,
			m *notifier.Model,
		) (bool, tea.Cmd) {
			m.Idle()
			return false, nil
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
			return true, m.AutoHideCmd(name)
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
			return true, m.AutoHideCmd(name)
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
		},
	)
	m.sortByCmdBar = sortByCmdBar
	bar := cmdbar.NewSearchCmdBar("Search topics by name")
	m.tcb = cmdbar.NewTableCmdsBar[string](
		cmdbar.NewDeleteCmdBar(deleteMsgFunc, deleteFunc, nil),
		bar,
		notifierCmdBar,
		sortByCmdBar,
	)
	m.lister = lister
	m.state = stateLoading
	var cmds []tea.Cmd
	cmds = append(cmds, m.lister.ListTopics)
	return &m, tea.Batch(cmds...)
}
