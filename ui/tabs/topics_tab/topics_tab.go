package topics_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/configs_page"
	"ktea/ui/pages/consumption_form_page"
	"ktea/ui/pages/consumption_page"
	"ktea/ui/pages/create_topic_page"
	"ktea/ui/pages/error_page"
	"ktea/ui/pages/nav"
	"ktea/ui/pages/publish_page"
	"ktea/ui/pages/topics_page"
)

type Model struct {
	active     nav.Page
	topicsPage *topics_page.Model
	statusbar  *statusbar.Model
	ka         *kadmin.SaramaKafkaAdmin
	ktx        *kontext.ProgramKtx
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if m.statusbar != nil {
		views = append(views, m.statusbar.View(ktx, renderer))
	}

	views = append(views, m.active.View(ktx, renderer))

	return ui.JoinVerticalSkipEmptyViews(views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case nav.LoadTopicsPageMsg:
		var cmd tea.Cmd
		m.active, cmd = topics_page.New(m.ka, m.ka)
		cmds = append(cmds, cmd)

	case nav.LoadConsumptionFormPageMsg:
		m.active = consumption_form_page.New(msg.Topic)

	case nav.LoadTopicConfigPageMsg:
		page, cmd := configs_page.New(m.ka, m.ka, m.topicsPage.SelectedTopicName())
		cmds = append(cmds, cmd)
		m.active = page

	case nav.LoadCreateTopicPageMsg:
		m.active = create_topic_page.New(m.ka)

	case nav.LoadPublishPageMsg:
		m.active = publish_page.New(m.ka, msg.Topic)

	case nav.LoadConsumptionPageMsg:
		var cmd tea.Cmd
		m.active, cmd = consumption_page.New(m.ka, msg.ReadDetails)
		cmds = append(cmds, cmd)

	}

	if cmd := m.active.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// always recreate the statusbar in case the active page might have changed
	m.statusbar = statusbar.New(m.active)

	return tea.Batch(cmds...)
}

func New(ktx *kontext.ProgramKtx, ka *kadmin.SaramaKafkaAdmin) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	listTopicView, cmd := topics_page.New(ka, ka)

	model := &Model{}
	model.ka = ka
	model.ktx = ktx
	model.active = listTopicView
	model.topicsPage = listTopicView
	model.statusbar = statusbar.New(model.active)

	return model, cmd
}

func NewInError(ktx *kontext.ProgramKtx) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	model := &Model{}
	model.ktx = ktx
	model.active = error_page.New(nil)

	return model, cmd
}
