package publish_page

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"strconv"
)

type state int

const (
	none       = 0
	publishing = 1
)

type Model struct {
	state     state
	form      *huh.Form
	content   huh.Field
	title     *huh.Input
	partition *huh.Input
	publisher kadmin.Publisher
	topic     kadmin.Topic
	notifier  *notifier.Model
}

type LoadPageMsg struct {
	Topic kadmin.Topic
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	notifierView := m.notifier.View(ktx, renderer)
	if m.form == nil {
		m.form = m.newForm(ktx)
	}
	return lipgloss.JoinVertical(
		lipgloss.Top,
		notifierView,
		renderer.Render(m.form.View()),
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
	case spinner.TickMsg:
		return m.notifier.Update(msg)
	case kadmin.PublicationStartedMsg:
		return waitForPublicationToCompleteCmd(msg)
	case PublicationFailed:
		m.state = none
		m.form.Init()
		return m.notifier.ShowErrorMsg("Publication failed!", fmt.Errorf("TODO"))
	case PublicationSucceeded:
		m.state = none
		m.form = nil
		return m.notifier.ShowSuccessMsg("Record published!")
	case tea.KeyMsg:
		m.notifier.Idle()
		switch msg.Type {
		case tea.KeyEsc:
			return ui.PublishMsg(pages.LoadTopicsPageMsg{})
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
					if partStr := m.form.GetString("PARTITION"); partStr != "" {
						if p, err := strconv.Atoi(partStr); err == nil {
							part = &p
						}
					}
					return m.publisher.PublishRecord(&kadmin.ProducerRecord{
						Key:       m.form.GetString("KEY"),
						Value:     m.form.GetString("PAYLOAD"),
						Topic:     m.topic.Name,
						Partition: part,
					})
				})
		}
		return cmd
	}
	return nil
}

type PublicationFailed struct {
}

type PublicationSucceeded struct {
}

func waitForPublicationToCompleteCmd(msg kadmin.PublicationStartedMsg) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-msg.Err:
			return PublicationFailed{}
		case <-msg.Published:
			return PublicationSucceeded{}
		}
	}
}

func (m *Model) newForm(ktx *kontext.ProgramKtx) *huh.Form {
	m.content = huh.NewText().
		ShowLineNumbers(true).
		Key("PAYLOAD").
		Title("Payload").
		WithHeight(ktx.AvailableHeight - 9)
	m.title = huh.NewInput().
		Title("Key").
		Key("KEY")
	m.partition = huh.NewInput().
		Title("Partition").
		Key("PARTITION").
		Validate(func(str string) error {
			if str == "" {
				return nil
			}
			if n, e := strconv.Atoi(str); e != nil {
				return errors.New(fmt.Sprintf("'%s' is not a valid numeric partition value", str))
			} else if n <= 0 {
				return errors.New("value must be greater than zero")
			} else if n > m.topic.Partitions-1 {
				return errors.New(fmt.Sprintf("partition index %s is invalid, valid range is 0-%d", str, m.topic.Partitions-1))
			}
			return nil
		})
	form := huh.NewForm(
		huh.NewGroup(
			m.title,
			m.partition,
			m.content,
		),
	)
	form.QuitAfterSubmit = false
	form.Init()
	return form
}

func New(p kadmin.Publisher, topic kadmin.Topic) *Model {
	return &Model{topic: topic, publisher: p, notifier: notifier.New()}
}
