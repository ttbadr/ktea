package create_subjects_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin/sr"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"time"
)

type state int

const (
	entering state = 0
	creating state = 1
)

type values struct {
	subject string
	schema  string
}

type Model struct {
	values
	form        *huh.Form
	creator     sr.SubjectCreator
	cmdBar      cmdbar.Widget
	state       state
	ktx         *kontext.ProgramKtx
	schemaInput *huh.Text
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	cmdbarView := m.cmdBar.View(ktx, renderer)

	if m.form == nil {
		m.form = newForm(m)
	}

	return ui.JoinVerticalSkipEmptyViews(
		cmdbarView,
		renderer.Render(lipgloss.NewStyle().PaddingTop(1).Render(m.form.View())),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	if m.form != nil {
		form, cmd := m.form.Update(msg)
		cmds = append(cmds, cmd)

		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}

		if m.form.State == huh.StateCompleted && m.state == entering {
			m.state = creating
			return func() tea.Msg {
				return m.creator.CreateSchema(sr.SubjectCreationDetails{
					Subject: m.subject,
					Schema:  m.schema,
				})
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		{
			switch msg.String() {
			case "esc":
				cmds = append(cmds, ui.PublishMsg(pages.LoadSubjectsPageMsg{}))
			}
		}
	case sr.SchemaCreatedMsg:
		m.form = nil
	case sr.SchemaCreationStartedMsg:
		cmds = append(cmds, msg.AwaitCompletion)
	}

	_, _, cmd := m.cmdBar.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Next Field", "tab"},
		{"Prev. Field", "s-tab"},
		{"Reset Form", "C-r"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Create Subject"
}

func newForm(model *Model) *huh.Form {
	model.subject = ""
	model.schema = ""
	schemaInput := huh.NewText().
		Value(&model.values.schema).
		Title("Schema").
		WithHeight(model.ktx.AvailableHeight - 5).(*huh.Text)
	model.schemaInput = schemaInput
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Value(&model.values.subject).
			Title("Subject"),
		schemaInput,
	))
	form.Init()
	form.QuitAfterSubmit = false
	return form
}

func New(creator sr.SubjectCreator, ktx *kontext.ProgramKtx) (*Model, tea.Cmd) {
	model := &Model{}
	model.ktx = ktx
	model.creator = creator
	model.state = entering
	notifierCmdBar := cmdbar.NewNotifierCmdBar()
	subjectListingStartedNotifier := func(msg sr.SchemaCreationStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Creating Schema")
		return true, cmd
	}
	hideNotificationMsgNotifier := func(msg cmdbar.HideNotificationMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return true, nil
	}
	f := func(msg sr.SchemaCreatedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Schema created")
		return true, func() tea.Msg {
			time.Sleep(2 * time.Second)
			return cmdbar.HideNotificationMsg{}
		}
	}
	cmdbar.WithMapping(notifierCmdBar, subjectListingStartedNotifier)
	cmdbar.WithMapping(notifierCmdBar, f)
	cmdbar.WithMapping(notifierCmdBar, hideNotificationMsgNotifier)
	model.cmdBar = notifierCmdBar
	return model, nil
}
