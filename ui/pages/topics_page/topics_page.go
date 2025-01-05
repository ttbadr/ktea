package topics_page

import (
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
}

type TopicListedMsg struct {
	Topics   []kadmin.Topic
	newTopic string
}

type TopicListedErrorMsg struct {
	Err error
}

func (t *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	cmdBarView := t.cmdBar.View(ktx, renderer)
	if cmdBarView != "" {
		views = append(views, cmdBarView)
	}

	t.table.SetHeight(ktx.AvailableHeight)
	t.table.SetWidth(ktx.WindowWidth - 2)
	t.table.SetColumns([]table.Column{
		{"Name", int(float64(ktx.WindowWidth-9) * 0.7)},
		{"Partitions", int(float64(ktx.WindowWidth-9) * 0.1)},
		{"Replicas", int(float64(ktx.WindowWidth-9) * 0.1)},
		{"In Sync Replicas", int(float64(ktx.WindowWidth-9) * 0.1)},
	})
	t.table.SetRows(t.rows)

	if t.moveCursor != nil {
		t.moveCursor()
	}

	if t.cmdBar.IsFocused() {
		render := renderer.Render(styles.Table.Blur.Render(t.table.View()))
		views = append(views, render)
	} else {
		render := renderer.Render(styles.Table.Focus.Render(t.table.View()))
		views = append(views, render)
	}

	return ui.JoinVerticalSkipEmptyViews(views...)
}
func (t *Model) Update(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 2)
	var newTopic string
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if t.topics == nil {
			return nil
		}
		switch msg.String() {
		case "ctrl+n":
			return ui.PublishMsg(nav.LoadCreateTopicPageMsg{})
		case "ctrl+u":
			return ui.PublishMsg(nav.LoadTopicConfigPageMsg{})
		case "ctrl+p":
			return ui.PublishMsg(nav.LoadPublishPageMsg{Topic: t.SelectedTopic()})
		case "enter":
			if t.cmdBar.IsNotFocused() {
				return ui.PublishMsg(nav.LoadConsumptionFormPageMsg{
					Topic: t.SelectedTopic(),
				})
			}
		}
		selectedTopic := t.SelectedTopicName()
		m, c := t.cmdBar.Update(msg, selectedTopic)
		if c != nil {
			cmds = append(cmds, c)
		}
		if m != nil {
			t.table, c = t.table.Update(m)
			if c != nil {
				cmds = append(cmds, c)
			}
		}
	case spinner.TickMsg:
		selectedTopic := t.SelectedTopicName()
		_, c := t.cmdBar.Update(msg, selectedTopic)
		if c != nil {
			cmds = append(cmds, c)
		}
	case kadmin.TopicListingStartedMsg:
		tickCmd := t.cmdBar.notifier.SpinWithLoadingMsg("Loading Topics")
		return tea.Batch(tickCmd, func() tea.Msg {
			select {
			case topics := <-msg.Topics:
				return TopicListedMsg{Topics: topics}
			case err := <-msg.Err:
				return TopicListedErrorMsg{Err: err}
			}
		})
	case TopicListedErrorMsg:
		t.cmdBar.notifier.ShowErrorMsg("Loading Topics Failed", msg.Err)
	case TopicListedMsg:
		t.cmdBar.notifier.Idle()
		t.topics = msg.Topics
		newTopic = msg.newTopic
	case kadmin.TopicDeletedMsg:
		t.cmdBar.notifier.Idle()
		t.topics = slices.DeleteFunc(
			t.topics,
			func(t kadmin.Topic) bool { return msg.TopicName == t.Name },
		)
		m, c := t.cmdBar.Update(msg, "")
		if c != nil {
			cmds = append(cmds, c)
		}
		if m != nil {
			t.table, c = t.table.Update(m)
			if c != nil {
				cmds = append(cmds, c)
			}
		}
	}

	if t.cmdBar.HasSearchedAtLeastOneChar() {
		t.table.GotoTop()
	}

	var rows []table.Row
	for _, topic := range t.topics {
		if t.cmdBar.GetSearchTerm() != "" {
			if strings.Contains(strings.ToLower(topic.Name), strings.ToLower(t.cmdBar.GetSearchTerm())) {
				rows = append(rows, table.Row{topic.Name, strconv.Itoa(topic.Partitions), strconv.Itoa(topic.Replicas), "N/A"})
			}
		} else {
			rows = append(rows, table.Row{topic.Name, strconv.Itoa(topic.Partitions), strconv.Itoa(topic.Replicas), "N/A"})
		}
	}
	t.rows = rows
	sort.SliceStable(t.rows, func(i, j int) bool {
		return t.rows[i][0] < t.rows[j][0]
	})

	if newTopic != "" {
		for i, r := range t.rows {
			if r[0] == newTopic {
				t.moveCursor = func() {
					t.table.GotoTop()
					t.table.MoveDown(i)
					t.moveCursor = nil
				}
				break
			}
		}
	}

	return tea.Batch(cmds...)
}

func (t *Model) Reset() {
	t.cmdBar.Reset()
	t.table.GotoTop()
}

func (t *Model) SelectedTopic() kadmin.Topic {
	selectedTopic := t.SelectedTopicName()
	for _, t := range t.topics {
		if t.Name == selectedTopic {
			return t
		}
	}
	panic("selected topic not found")
}

func (t *Model) SelectedTopicName() string {
	selectedRow := t.table.SelectedRow()
	var selectedTopic string
	if selectedRow != nil {
		selectedTopic = selectedRow[0]
	}
	return selectedTopic
}

func (t *Model) Title() string {
	if t.cmdBar.IsFocused() {
		return t.cmdBar.Title()
	} else {
		return "Topics"
	}
}

func (t *Model) Shortcuts() []statusbar.Shortcut {
	if t.cmdBar.IsFocused() {
		return t.cmdBar.Shortcuts()
	} else {
		return t.shortcuts
	}
}

func New(topicDeleter kadmin.TopicDeleter, lister kadmin.TopicLister) (*Model, tea.Cmd) {
	var t = Model{}
	t.shortcuts = []statusbar.Shortcut{
		{"Search", "/"},
		{"Consume", "enter"},
		{"Publish", "C-p"},
		{"Create", "C-n"},
		{"Delete", "C-d"},
		{"Configs", "C-u"},
		{"ACLs", "C-a"},
		{"Groups", "C-g"},
		{"Sort By Partitions", "F2"},
		{"Sort By Name", "F3"},
	}
	t.table = table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	t.table.SetColumns([]table.Column{
		{"Name", 1},
		{"Partitions", 1},
		{"Replicas", 1},
		{"In Sync Replicas", 1},
	})
	t.cmdBar = NewCmdBar(topicDeleter)
	return &t, lister.ListTopics
}
