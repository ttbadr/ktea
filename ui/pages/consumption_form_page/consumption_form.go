package consumption_form_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/navigation"
)

type Model struct {
	form  *huh.Form
	topic kadmin.Topic
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return renderer.RenderWithStyle(m.form.View(), styles.Form)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return ui.PublishMsg(navigation.LoadTopicsPageMsg{})
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		return ui.PublishMsg(navigation.LoadConsumptionPageMsg{
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

func newForm() *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[kadmin.StartPoint]().
				Title("Start form").
				Options(
					huh.NewOption("Beginning", kadmin.Beginning),
					huh.NewOption("Most Recent", kadmin.MostRecent),
					huh.NewOption("Today", kadmin.Today)),
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
		form:  newForm(),
	}
}
