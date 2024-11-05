package topics_page

// list topics command bar

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
)

const (
	HIDDEN              state = 0
	SEARCHING           state = 1
	SEARCHED            state = 2
	DELETE_CONFIRMATION state = 3
	SPINNING            state = 4
)

type CmdBarModel struct {
	state         state
	searchInput   *huh.Input
	deleteConfirm *huh.Confirm
	notifier      *notifier.Model
	topicDeleter  kadmin.TopicDeleter
}

func (m *CmdBarModel) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string

	if m.state == SEARCHING || m.state == SEARCHED {
		views = append(views, renderer.RenderWithStyle(m.searchInput.View(), styles.CmdBar))
	} else if m.state == DELETE_CONFIRMATION {
		views = append(views, renderer.RenderWithStyle(m.deleteConfirm.View(), styles.CmdBar))
	}

	views = append(views, renderer.Render(m.notifier.View(ktx, renderer)))
	return ui.JoinVerticalSkipEmptyViews(views...)
}

func (m *CmdBarModel) Shortcuts() []statusbar.Shortcut {
	if m.state == DELETE_CONFIRMATION {
		return []statusbar.Shortcut{
			{"Confirm", "enter"},
			{"Select Cancel", "c"},
			{"Select Delete", "d"},
			{"Quit", "esc"},
		}
	} else if m.state == SEARCHING {
		return []statusbar.Shortcut{
			{"Confirm", "enter"},
			{"Cancel", "esc"},
		}
	}
	return nil
}

func (m *CmdBarModel) Title() string {
	if m.state == DELETE_CONFIRMATION {
		return "Delete Topic"
	} else if m.state == SEARCHING {
		return "Topics / Search"
	}
	return ""
}

// Update returns the tea.Msg if it is not being handled or nil if it is
func (m *CmdBarModel) Update(msg tea.Msg, selectedTopic string) (tea.Msg, tea.Cmd) {
	var propagationMsg = msg
	var cmd tea.Cmd = nil
	switch msg := msg.(type) {
	case spinner.TickMsg:
		cmd = m.notifier.Update(msg)
	case tea.KeyMsg:
		if msg.String() == "/" {
			propagationMsg = m.handleSlashKeyMsg(msg)
		} else if msg.String() == "enter" {
			propagationMsg, cmd = m.handleEnterKeyMsg(msg, cmd, selectedTopic)
		} else if msg.String() == "ctrl+d" {
			propagationMsg = m.handleCtrlDKeyMsg(selectedTopic)
		} else if msg.String() == "esc" {
			m.handleEscKeyMsg()
		} else if m.state == SEARCHING {
			propagationMsg = nil
			input, _ := m.searchInput.Update(msg)
			if i, ok := input.(*huh.Input); ok {
				m.searchInput = i
			}
		} else if m.state == DELETE_CONFIRMATION {
			propagationMsg = nil
			confirm, _ := m.deleteConfirm.Update(msg)
			if c, ok := confirm.(*huh.Confirm); ok {
				m.deleteConfirm = c
			}
		}
	case kadmin.TopicDeletedMsg:
		m.state = HIDDEN
	}
	return propagationMsg, cmd
}

func (m *CmdBarModel) handleEscKeyMsg() {
	if m.state == DELETE_CONFIRMATION {
		m.deleteConfirm = newDeleteConfirm()
		m.state = HIDDEN
	} else if m.state == SEARCHING {
		m.searchInput = newSearchInput()
		m.state = HIDDEN
	} else if m.state == SEARCHED {

	} else {
		m.state = HIDDEN
	}
}

func (m *CmdBarModel) handleCtrlDKeyMsg(topic string) tea.Msg {
	var propagationMsg tea.Msg = nil
	m.state = DELETE_CONFIRMATION
	// TODO: clean-up style
	m.deleteConfirm.Title(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render("🗑️  "+topic) + lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7571F9")).
		Bold(true).
		Render(" will be delete permanently")).
		Focus()
	return propagationMsg
}

func (m *CmdBarModel) handleEnterKeyMsg(msg tea.Msg, cmd tea.Cmd, selectedTopic string) (tea.Msg, tea.Cmd) {
	var propagationMsg tea.Msg = nil
	if m.state == SEARCHING {
		propagationMsg = nil
		if m.GetSearchTerm() == "" {
			m.state = HIDDEN
		} else {
			m.state = SEARCHED
		}
		m.searchInput.Blur()
	} else if m.state == DELETE_CONFIRMATION {
		propagationMsg = nil
		confirm, _ := m.deleteConfirm.Update(msg)
		if c, ok := confirm.(*huh.Confirm); ok {
			m.deleteConfirm = c
		}
		m.state = HIDDEN
		if m.deleteConfirm.GetValue().(bool) {
			m.state = SPINNING
			cmd = tea.Batch(func() tea.Msg { return m.topicDeleter.DeleteTopic(selectedTopic) }, m.notifier.SpinWithRocketMsg("Deleting Topic"))
		}
	}
	return propagationMsg, cmd
}

func (m *CmdBarModel) handleSlashKeyMsg(msg tea.KeyMsg) tea.Msg {
	if m.state == SEARCHING {
		m.searchInput.Update(msg)
	} else {
		m.state = SEARCHING
		m.searchInput.Focus()
	}
	return nil
}

func (m *CmdBarModel) HasSearchedAtLeastOneChar() bool {
	return m.state == SEARCHING && len(m.searchInput.GetValue().(string)) > 0
}

func (m *CmdBarModel) GetSearchTerm() string {
	return m.searchInput.GetValue().(string)
}

func (m *CmdBarModel) IsFocused() bool {
	return m.state == SEARCHING || m.state == DELETE_CONFIRMATION
}

func (m *CmdBarModel) IsNotFocused() bool {
	return m.state != SEARCHING && m.state != DELETE_CONFIRMATION
}

func (m *CmdBarModel) Reset() {
	m.state = HIDDEN
}

func newDeleteConfirm() *huh.Confirm {
	return huh.NewConfirm().
		Inline(true).
		Affirmative("Delete!").
		Negative("Cancel.").
		WithKeyMap(&huh.KeyMap{
			Confirm: huh.ConfirmKeyMap{
				Submit: key.NewBinding(key.WithKeys("enter")),
				Toggle: key.NewBinding(key.WithKeys("h", "l", "right", "left")),
				Accept: key.NewBinding(key.WithKeys("d")),
				Reject: key.NewBinding(key.WithKeys("c")),
			},
		}).(*huh.Confirm)
}

func newSearchInput() *huh.Input {
	searchInput := huh.NewInput().
		Key("searchTerm").
		Placeholder("Search for Topic")
	searchInput.Init()
	return searchInput
}

func NewCmdBar(topicDeleter kadmin.TopicDeleter) *CmdBarModel {
	return &CmdBarModel{
		state:         HIDDEN,
		searchInput:   newSearchInput(),
		deleteConfirm: newDeleteConfirm(),
		notifier:      notifier.New(),
		topicDeleter:  topicDeleter,
	}
}
