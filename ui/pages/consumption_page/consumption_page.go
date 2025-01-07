package consumption_page

import (
	"context"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"strconv"
)

type Model struct {
	table              *table.Model
	cmdBar             *ConsumptionCmdBar
	consumerRecordChan chan kadmin.ConsumerRecord
	cancelConsumption  context.CancelFunc
	errChan            chan error
	reader             kadmin.RecordReader
	rows               []table.Row
	readDetails        kadmin.ReadDetails
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	cmdBarView := m.cmdBar.View(ktx, renderer)

	m.table.SetColumns([]table.Column{
		{Title: "Key", Width: int(float64(ktx.WindowWidth-5) * 0.5)},
		{Title: "Partition", Width: int(float64(ktx.WindowWidth-5) * 0.25)},
		{Title: "Offset", Width: int(float64(ktx.WindowWidth-5) * 0.25)},
	})
	m.table.SetHeight(ktx.AvailableHeight - 2)
	m.table.SetRows(m.rows)
	tableRender := renderer.Render(styles.Table.Focus.Render(m.table.View()))
	return ui.JoinVerticalSkipEmptyViews(cmdBarView, tableRender)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{}
}

func (m *Model) Title() string {
	return "Topics / " + m.readDetails.Topic.Name + " / Records"
}

type ConsumerRecordReceived struct {
	Record kadmin.ConsumerRecord
}

type ConsumptionEndedMsg struct{}

func (m *Model) waitForActivity() tea.Cmd {
	return func() tea.Msg {
		for {
			select {
			case record, ok := <-m.consumerRecordChan:
				if !ok {
					return ConsumptionEndedMsg{}
				}
				return ConsumerRecordReceived{Record: record}
			case err := <-m.errChan:
				return err
			}
		}
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	cmd := m.cmdBar.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.cancelConsumption()
			return ui.PublishMsg(nav.LoadConsumptionFormPageMsg{ReadDetails: &m.readDetails})
		} else if msg.String() == "f2" {
			m.cancelConsumption()
		} else {
			t, cmd := m.table.Update(msg)
			m.table = &t
			cmds = append(cmds, cmd)
		}
	case kadmin.ReadingStartedMsg:
		m.consumerRecordChan = msg.ConsumerRecord
		m.errChan = msg.Err
		cmds = append(cmds, m.waitForActivity())
	case ConsumptionEndedMsg:
		return nil
	case ConsumerRecordReceived:
		var key string
		if msg.Record.Key == "" {
			key = "null"
		} else {
			key = msg.Record.Key
		}
		m.rows = append(
			[]table.Row{
				{
					key,
					strconv.FormatInt(msg.Record.Partition, 10),
					strconv.FormatInt(msg.Record.Offset, 10),
				},
			},
			m.rows...,
		)
		return m.waitForActivity()
	}
	return tea.Batch(cmds...)
}

func New(reader kadmin.RecordReader, readDetails kadmin.ReadDetails) (nav.Page, tea.Cmd) {
	m := &Model{}

	t := table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
	m.table = &t
	m.reader = reader
	m.cmdBar = NewConsumptionCmdbar()
	m.readDetails = readDetails

	ctx, cancelFunc := context.WithCancel(context.Background())
	m.cancelConsumption = cancelFunc

	return m, func() tea.Msg {
		return m.reader.ReadRecords(ctx, readDetails)
	}
}
