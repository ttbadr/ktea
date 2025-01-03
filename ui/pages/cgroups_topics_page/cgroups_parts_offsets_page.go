package cgroups_topics_page

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/navigation"
	"slices"
	"sort"
	"strconv"
)

const (
	topicFocus  tableFocus = 0
	offsetFocus tableFocus = 1
)

type tableFocus int

type Model struct {
	tableFocus        tableFocus
	topicsTable       table.Model
	offsetsTable      table.Model
	topicsRows        []table.Row
	offsetRows        []table.Row
	groupName         string
	topicByPartOffset map[string][]partOffset
	cmdBar            *CmdBar
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	m.topicsTable.SetHeight(ktx.AvailableHeight)
	i := int(float64(ktx.WindowWidth / 2))
	m.topicsTable.SetWidth(i - 2)
	m.topicsTable.SetColumns([]table.Column{
		{"Topic Name", int(float64(i - 4))},
	})
	m.topicsTable.SetRows(m.topicsRows)

	m.offsetsTable.SetHeight(ktx.AvailableHeight)
	m.offsetsTable.SetColumns([]table.Column{
		{"Partition", int(float64(i-6) * 0.5)},
		{"Offset", int(float64(i-5) * 0.5)},
	})
	m.offsetsTable.SetRows(m.offsetRows)

	topicTableRender := styles.Table.Blur
	offsetTableRender := styles.Table.Blur
	if m.tableFocus == topicFocus {
		topicTableRender = styles.Table.Focus
		offsetTableRender = styles.Table.Blur
	}

	return lg.JoinVertical(lg.Left,
		m.cmdBar.View(ktx, renderer),
		lg.JoinHorizontal(
			lg.Top,
			topicTableRender.Render(m.topicsTable.View()),
			offsetTableRender.Render(m.offsetsTable.View()),
		))
}

type partOffset struct {
	partition string
	offset    string
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return ui.PublishMsg(navigation.LoadCGroupsPageMsg{})
		}
	case kadmin.OffsetListingStartedMsg:
		cmds = append(
			cmds,
			msg.AwaitCompletion,
			m.cmdBar.notifier.SpinWithLoadingMsg("Loading Offsets"),
		)
	case kadmin.OffsetListedMsg:
		m.cmdBar.notifier.Idle()
		m.handleOffsetListed(msg)
	}

	var cmd tea.Cmd
	if m.tableFocus == topicFocus {
		m.topicsTable, cmd = m.topicsTable.Update(msg)
	} else {
		m.offsetsTable, cmd = m.offsetsTable.Update(msg)
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	cmd = m.cmdBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// recreate offset rows after topic table has been updated
	m.recreateOffsetRows()

	return tea.Batch(cmds...)
}

func (m *Model) recreateOffsetRows() {
	// if topics aren't listed yet
	if m.topicsRows == nil {
		return
	}
	selectedTopic := m.selectedRow()
	if selectedTopic == "" {
		selectedTopic = m.topicsRows[0][0]
	}
	m.offsetRows = []table.Row{}
	for _, partOffset := range m.topicByPartOffset[selectedTopic] {
		m.offsetRows = append(m.offsetRows, table.Row{
			partOffset.partition,
			partOffset.offset,
		})
	}
	sort.SliceStable(m.offsetRows, func(i, j int) bool {
		return m.offsetRows[i][0] < m.offsetRows[j][0]
	})
}

func (m *Model) handleOffsetListed(msg kadmin.OffsetListedMsg) {
	var topics []string
	m.topicByPartOffset = make(map[string][]partOffset)
	for _, offset := range msg.Offsets {
		if !slices.Contains(topics, offset.Topic) {
			topics = append(topics, offset.Topic)
		}
		partOffset := partOffset{
			partition: strconv.FormatInt(int64(offset.Partition), 10),
			offset:    strconv.FormatInt(offset.Offset, 10),
		}
		m.topicByPartOffset[offset.Topic] = append(m.topicByPartOffset[offset.Topic], partOffset)
	}
	m.topicsRows = []table.Row{}
	for _, topic := range topics {
		m.topicsRows = append(m.topicsRows, table.Row{topic})
	}
	sort.SliceStable(m.topicsRows, func(i, j int) bool {
		return m.topicsRows[i][0] < m.topicsRows[j][0]
	})
}

func (m *Model) selectedRow() string {
	row := m.topicsTable.SelectedRow()
	if row == nil {
		return ""
	}
	return row[0]
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{}
}

func (m *Model) Title() string {
	return "Consumer Groups / " + m.groupName
}

func New(lister kadmin.OffsetLister, group string) (*Model, tea.Cmd) {
	tt := table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	ot := table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	return &Model{
			cmdBar:       NewCmdBar(),
			tableFocus:   topicFocus,
			groupName:    group,
			topicsTable:  tt,
			offsetsTable: ot,
		}, func() tea.Msg {
			return lister.ListOffsets(group)
		}
}
