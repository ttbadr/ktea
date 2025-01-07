package publish_page

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"strconv"
)

type state int

const (
	none       = 0
	publishing = 1
)

type Model struct {
	state      state
	form       *huh.Form
	publisher  kadmin.Publisher
	topic      *kadmin.Topic
	notifier   *notifier.Model
	formValues *FormValues
}

type LoadPageMsg struct {
	Topic kadmin.Topic
}

type FormValues struct {
	Key       string
	Partition string
	Payload   string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	notifierView := m.notifier.View(ktx, renderer)
	if m.form == nil {
		m.form = m.newForm(ktx)
	}
	return ui.JoinVerticalSkipEmptyViews(
		notifierView,
		renderer.RenderWithStyle(m.form.View(), styles.Form),
	)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Reset Form", "C-r"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Topics / " + m.topic.Name + " / Publish"
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case spinner.TickMsg, notifier.HideNotificationMsg:
		return m.notifier.Update(msg)
	case kadmin.PublicationStartedMsg:
		return tea.Batch(
			m.notifier.SpinWithLoadingMsg("Publishing record"),
			func() tea.Msg { return msg.AwaitCompletion() },
		)
	case kadmin.PublicationFailed:
		m.state = none
		m.form.Init()
		return m.notifier.ShowErrorMsg("Publication failed!", fmt.Errorf("TODO"))
	case kadmin.PublicationSucceeded:
		m.state = none
		m.formValues.Key = ""
		m.formValues.Partition = ""
		m.formValues.Payload = ""
		m.form = nil
		return tea.Batch(
			m.notifier.ShowSuccessMsg("Record published!"),
			func() tea.Msg {
				return notifier.HideNotificationMsg{}
			})
	case tea.KeyMsg:
		m.notifier.Idle()
		switch msg.Type {
		case tea.KeyEsc:
			return ui.PublishMsg(nav.LoadTopicsPageMsg{})
		}
	}
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		if m.form != nil && m.form.State == huh.StateCompleted {
			m.state = publishing
			m.form.State = huh.StateNormal
			return tea.Batch(
				m.notifier.SpinWithRocketMsg("Publishing record"),
				func() tea.Msg {
					var part *int
					if m.formValues.Partition != "" {
						if p, err := strconv.Atoi(m.formValues.Partition); err == nil {
							part = &p
						}
					}
					return m.publisher.PublishRecord(&kadmin.ProducerRecord{
						Key:       m.formValues.Key,
						Value:     m.formValues.Payload,
						Topic:     m.topic.Name,
						Partition: part,
					})
				})
		}
		return cmd
	}
	return nil
}

func (m *Model) newForm(ktx *kontext.ProgramKtx) *huh.Form {
	content := huh.NewText().
		ShowLineNumbers(true).
		Value(&m.formValues.Payload).
		Title("Payload").
		WithHeight(ktx.AvailableHeight - 10)
	key := huh.NewInput().
		Title("Key").
		Value(&m.formValues.Key)
	partition := huh.NewInput().
		Value(&m.formValues.Partition).
		Title("Partition").
		Validate(func(str string) error {
			if str == "" {
				return nil
			}
			if n, e := strconv.Atoi(str); e != nil {
				return errors.New(fmt.Sprintf("'%s' is not a valid numeric partition value", str))
			} else if n < 0 {
				return errors.New("value must be at least zero")
			} else if n > m.topic.Partitions-1 {
				return errors.New(fmt.Sprintf("partition index %s is invalid, valid range is 0-%d", str, m.topic.Partitions-1))
			}
			return nil
		})
	form := huh.NewForm(
		huh.NewGroup(
			key,
			partition,
			content,
		),
	)
	form.QuitAfterSubmit = false
	form.Init()
	return form
}

func New(p kadmin.Publisher, topic *kadmin.Topic) *Model {
	return &Model{topic: topic, publisher: p, notifier: notifier.New(), formValues: &FormValues{}}
}
