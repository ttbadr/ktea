package record_details_page

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	ktable "ktea/ui/components/table"
	"ktea/ui/pages/nav"
	"strconv"
	"strings"
	"time"
)

type focus bool

const (
	content focus = true
	headers focus = false
)

type Model struct {
	record        *kadmin.ConsumerRecord
	contentVp     *viewport.Model
	headerValueVp *viewport.Model
	topic         *kadmin.Topic
	headersTable  *table.Model
	rows          []table.Row
	focus         focus
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	key := m.record.Key
	if key == "" {
		key = "<null>"
	}

	payloadWidth := int(float64(ktx.WindowWidth) * 0.70)
	height := ktx.AvailableHeight - 2
	vp := viewport.New(payloadWidth, height)
	vp.SetContent(ui.PrettyPrintJson(m.record.Value))
	m.contentVp = &vp

	m.rows = []table.Row{}
	for _, header := range m.record.Headers {
		m.rows = append(m.rows, table.Row{header.Key})
	}

	metaInfoVp := fmt.Sprintf("key: %s\ntimestamp: %s", key, m.record.Timestamp.Format(time.UnixDate))
	sideBarWidth := ktx.WindowWidth - (payloadWidth + 7)
	if len(m.record.Headers) == 0 {

	} else {
		headerValueTableHeight := len(m.record.Headers) + 4

		headerValueVp := viewport.New(sideBarWidth, height-headerValueTableHeight-4)
		m.headerValueVp = &headerValueVp
		m.headersTable.SetColumns([]table.Column{
			{"Header Key", sideBarWidth},
		})
		m.headersTable.SetHeight(headerValueTableHeight)
		m.headersTable.SetRows(m.rows)
	}

	row := m.headersTable.SelectedRow()
	var headerValue string
	if row == nil {
		if len(m.record.Headers) > 0 {
			headerValue = m.record.Headers[0].Value
		}
	} else {
		headerValue = m.record.Headers[m.headersTable.Cursor()].Value
	}
	headerValueLine := strings.Builder{}
	for i := 0; i < sideBarWidth; i++ {
		headerValueLine.WriteString("─")
	}
	m.headerValueVp.SetContent("Header Value\n" + headerValueLine.String() + "\n" + headerValue)

	var contentStyle lipgloss.Style
	var headersTableStyle lipgloss.Style
	if m.focus == content {
		contentStyle = lipgloss.NewStyle().
			Inherit(styles.TextViewPort).
			BorderForeground(lipgloss.Color(styles.ColorFocusBorder))
		headersTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0).
			Margin(0).
			BorderForeground(lipgloss.Color(styles.ColorBlurBorder))
	} else {
		contentStyle = lipgloss.NewStyle().
			Inherit(styles.TextViewPort).
			BorderForeground(lipgloss.Color(styles.ColorBlurBorder))
		headersTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0).
			Margin(0).
			BorderForeground(lipgloss.Color(styles.ColorFocusBorder))
	}

	return lipgloss.NewStyle().
		Render(lipgloss.JoinHorizontal(
			lipgloss.Top,
			renderer.RenderWithStyle(m.contentVp.View(), contentStyle),
			ui.JoinVertical(
				lipgloss.Top,
				lipgloss.NewStyle().Padding(1).Render(metaInfoVp),
				headersTableStyle.Render(lipgloss.JoinVertical(lipgloss.Top, m.headersTable.View(), m.headerValueVp.View())),
			)))
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.contentVp == nil {
		return nil
	}

	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return ui.PublishMsg(nav.LoadCachedConsumptionPageMsg{})
		case "ctrl+h":
			m.focus = !m.focus
		case "c":
			if m.focus == content {
				err := clipboard.WriteAll(m.record.Value)
				if err != nil {
					return nil
				}
			} else {
			}
		default:
			if m.focus == content {
				vp, cmd := m.contentVp.Update(msg)
				cmds = append(cmds, cmd)
				m.contentVp = &vp
			} else {
				t, cmd := m.headersTable.Update(msg)
				cmds = append(cmds, cmd)
				m.headersTable = &t
			}
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	whatToCopy := "Header Value"
	if m.focus == content {
		whatToCopy = "Content"
	}
	return []statusbar.Shortcut{
		{"Toggle Headers/Content", "C-h"},
		{"Go Back", "esc"},
		{"Copy " + whatToCopy, "c"},
	}
}

func (m *Model) Title() string {
	return "Topics / " + m.topic.Name + " / Records / " + strconv.FormatInt(m.record.Offset, 10)
}

func New(record *kadmin.ConsumerRecord, topic *kadmin.Topic) *Model {
	headersTable := ktable.NewDefaultTable()
	return &Model{
		record:       record,
		topic:        topic,
		headersTable: &headersTable,
		focus:        content,
	}
}
