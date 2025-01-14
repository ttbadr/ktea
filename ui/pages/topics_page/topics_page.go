package topics_page

import (
	"context"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type state int

type Model struct {
	topics     []kadmin.Topic
	table      table.Model
	shortcuts  []statusbar.Shortcut
	cmdBar     *CmdBarModel
	rows       []table.Row
	moveCursor func()
	lister     kadmin.TopicLister
	Ctx        context.Context
	ctx        context.Context
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	cmdBarView := m.cmdBar.View(ktx, renderer)
	if cmdBarView != "" {
		views = append(views, cmdBarView)
	}

	m.table.SetHeight(ktx.AvailableHeight)
	m.table.SetWidth(ktx.WindowWidth - 2)
	m.table.SetColumns([]table.Column{
		{"Name", int(float64(ktx.WindowWidth-9) * 0.7)},
		{"Partitions", int(float64(ktx.WindowWidth-9) * 0.1)},
		{"Replicas", int(float64(ktx.WindowWidth-9) * 0.1)},
		{"In Sync Replicas", int(float64(ktx.WindowWidth-9) * 0.1)},
	})
	m.table.SetRows(m.rows)

	if m.moveCursor != nil {
		m.moveCursor()
	}

	if m.cmdBar.IsFocused() {
		render := renderer.Render(styles.Table.Blur.Render(m.table.View()))
		views = append(views, render)
	} else {
		render := renderer.Render(styles.Table.Focus.Render(m.table.View()))
		views = append(views, render)
	}

	return ui.JoinVerticalSkipEmptyViews(views...)
}
func (m *Model) Update(msg tea.Msg) tea.Cmd {
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
			return m.lister.ListTopics
		case "enter":
			if m.cmdBar.IsNotFocused() {
				return ui.PublishMsg(nav.LoadConsumptionFormPageMsg{
					Topic: m.SelectedTopic(),
				})
			}
		}
		selectedTopic := m.SelectedTopicName()
		pmsg, c := m.cmdBar.Update(msg, selectedTopic)
		if c != nil {
			cmds = append(cmds, c)
		}
		if pmsg != nil {
			m.table, c = m.table.Update(pmsg)
			if c != nil {
				cmds = append(cmds, c)
			}
		}
	case spinner.TickMsg:
		selectedTopic := m.SelectedTopicName()
		_, c := m.cmdBar.Update(msg, selectedTopic)
		if c != nil {
			cmds = append(cmds, c)
		}
	case kadmin.TopicListingStartedMsg:
		tickCmd := m.cmdBar.notifier.SpinWithLoadingMsg("Loading Topics")
		return tea.Batch(
			tickCmd,
			msg.AwaitCompletion,
		)
	case kadmin.TopicListedErrorMsg:
		m.cmdBar.notifier.ShowErrorMsg("Loading Topics Failed", msg.Err)
	case kadmin.TopicListedMsg:
		m.cmdBar.notifier.Idle()
		m.topics = msg.Topics
	case kadmin.TopicDeletedMsg:
		m.cmdBar.notifier.Idle()
		m.topics = slices.DeleteFunc(
			m.topics,
			func(t kadmin.Topic) bool { return msg.TopicName == t.Name },
		)
		pmsg, c := m.cmdBar.Update(msg, "")
		if c != nil {
			cmds = append(cmds, c)
		}
		if pmsg != nil {
			m.table, c = m.table.Update(pmsg)
			if c != nil {
				cmds = append(cmds, c)
			}
		}
	}

	if m.cmdBar.HasSearchedAtLeastOneChar() {
		m.table.GotoTop()
	}

	var rows []table.Row
	for _, topic := range m.topics {
		if m.cmdBar.GetSearchTerm() != "" {
			if strings.Contains(strings.ToLower(topic.Name), strings.ToLower(m.cmdBar.GetSearchTerm())) {
				rows = append(rows, table.Row{topic.Name, strconv.Itoa(topic.Partitions), strconv.Itoa(topic.Replicas), "N/A"})
			}
		} else {
			rows = append(rows, table.Row{topic.Name, strconv.Itoa(topic.Partitions), strconv.Itoa(topic.Replicas), "N/A"})
		}
	}
	m.rows = rows
	sort.SliceStable(m.rows, func(i, j int) bool {
		return m.rows[i][0] < m.rows[j][0]
	})

	return tea.Batch(cmds...)
}

func (m *Model) Reset() {
	m.cmdBar.Reset()
	m.table.GotoTop()
}

func (m *Model) SelectedTopic() *kadmin.Topic {
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
	if m.cmdBar.IsFocused() {
		return m.cmdBar.Title()
	} else {
		return "Topics"
	}
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	if m.cmdBar.IsFocused() {
		return m.cmdBar.Shortcuts()
	} else {
		return m.shortcuts
	}
}

func New(topicDeleter kadmin.TopicDeleter, lister kadmin.TopicLister) (*Model, tea.Cmd) {
	var m = Model{}
	m.shortcuts = []statusbar.Shortcut{
		{"Search", "/"},
		{"Consume", "enter"},
		{"Publish", "C-p"},
		{"Create", "C-n"},
		{"Delete", "F2"},
		{"Configs", "C-o"},
		{"Refresh", "F5"},
	}
	m.table = table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	m.table.SetColumns([]table.Column{
		{"Name", 1},
		{"Partitions", 1},
		{"Replicas", 1},
		{"In Sync Replicas", 1},
	})
	m.cmdBar = NewCmdBar(topicDeleter)
	m.lister = lister
	return &m, lister.ListTopics
}
