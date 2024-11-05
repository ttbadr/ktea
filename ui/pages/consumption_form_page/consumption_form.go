package consumption_form_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
)

type Model struct {
	form  *huh.Form
	topic kadmin.Topic
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return m.form.View()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return ui.PublishMsg(pages.LoadTopicsPageMsg{})
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		return ui.PublishMsg(pages.LoadConsumptionPageMsg{
			Topic: m.topic,
		})
	}
	return cmd
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Next Field", "tab"},
		{"Prev. Field", "s-tab"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Consumption details"
}

func newCreateTopicForm() *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Start form").
				Options(
					huh.NewOption("Beginning", "beginning"),
					huh.NewOption("Most Recent", "most-recent"),
					huh.NewOption("Today", "today")),
			huh.NewSelect[string]().
				Title("Limit").
				Options(
					huh.NewOption("50", "50"),
					huh.NewOption("500", "500"),
					huh.NewOption("5000", "5000")),
		),
	)
	form.Init()
	return form
}

func New(topic kadmin.Topic) *Model {
	return &Model{
		topic: topic,
		form:  newCreateTopicForm(),
	}
}
