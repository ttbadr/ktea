package consumption_page

import (
	"context"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"strconv"
)

type Model struct {
	readDetails       kadmin.ReadDetails
	consumedRecords   []kadmin.ConsumerRecord
	table             *table.Model
	receivingChan     chan kadmin.ConsumerRecord
	cancelConsumption context.CancelFunc
	errChan           chan error
	reader            kadmin.RecordReader
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{}
}

func (m *Model) Title() string {
	return ""
}

type ConsumerRecordReceived struct {
	Record kadmin.ConsumerRecord
}

func (m *Model) waitForActivity() tea.Cmd {
	return func() tea.Msg {
		select {
		case m := <-m.receivingChan:
			return ConsumerRecordReceived{m}
		case e := <-m.errChan:
			log.Info(e)
			return e
		}
	}
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if len(m.consumedRecords) > 0 {
		m.table.SetColumns([]table.Column{
			{Title: "Key", Width: int(float64(ktx.WindowWidth-5) * 0.5)},
			{Title: "Partition", Width: int(float64(ktx.WindowWidth-5) * 0.25)},
			{Title: "Offset", Width: int(float64(ktx.WindowWidth-5) * 0.25)},
		})
		var rows []table.Row
		for _, r := range m.consumedRecords {
			var key string
			if r.Key == "" {
				key = "null"
			} else {
				key = r.Key
			}
			rows = append(rows, table.Row{key, strconv.FormatInt(r.Partition, 10), strconv.FormatInt(r.Offset, 10)})
		}
		m.table.SetRows(rows)
	}
	return m.table.View()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.cancelConsumption()
			return ui.PublishMsg(nav.LoadTopicsPageMsg{})
		}
	case kadmin.ReadingStartedMsg:
		m.receivingChan = msg.ConsumerRecord
		m.errChan = msg.Err
		return m.waitForActivity()
	case ConsumerRecordReceived:
		m.consumedRecords = append(m.consumedRecords, msg.Record)
		return m.waitForActivity()
	}
	return nil
}

func New(reader kadmin.RecordReader, readDetails kadmin.ReadDetails) (nav.Page, tea.Cmd) {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t := table.New(
		table.WithStyles(s),
	)
	m := &Model{}
	m.reader = reader
	m.readDetails = readDetails
	m.table = &t
	ctx, cancelFunc := context.WithCancel(context.Background())
	m.cancelConsumption = cancelFunc
	cmd := func() tea.Msg { return m.reader.ReadRecords(ctx, kadmin.ReadDetails{Topic: m.readDetails.Topic}) }
	return m, cmd
}
