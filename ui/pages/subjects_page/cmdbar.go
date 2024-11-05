package subjects_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kadmin/sr"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"time"
)

type SubjectsCmdBar struct {
	deleteWidget     *cmdbar.DeleteCmdBarModel[sr.Subject]
	searchWidget     cmdbar.Widget
	notifierWidget   cmdbar.Widget
	active           cmdbar.Widget
	searchPrevActive bool
}

func (s *SubjectsCmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if s.active != nil {
		return s.active.View(ktx, renderer)
	}
	return ""
}

func (s *SubjectsCmdBar) Update(msg tea.Msg, selectedSubject sr.Subject) (tea.Msg, tea.Cmd) {
	// when notifier is active it is receiving priority to handle messages
	// until a message comes in that deactivates the notifier
	if s.active == s.notifierWidget {
		s.active = s.notifierWidget
		active, pmsg, cmd := s.active.Update(msg)
		if !active {
			s.active = nil
		}
		return pmsg, cmd
	}

	switch msg := msg.(type) {
	case sr.SubjectListingStartedMsg, sr.SubjectDeletionStartedMsg:
		s.active = s.notifierWidget
		_, pmsg, cmd := s.active.Update(msg)
		return pmsg, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			s.active = s.searchWidget
			active, pmsg, cmd := s.active.Update(msg)
			if !active {
				s.active = nil
			}
			return pmsg, cmd
		case "ctrl+d":
			if _, ok := s.active.(*cmdbar.SearchCmdBarModel); ok {
				s.searchPrevActive = true
			}
			s.deleteWidget.Delete(selectedSubject)
			s.active = s.deleteWidget
			_, pmsg, cmd := s.active.Update(msg)
			return pmsg, cmd
		}
	}

	if s.active != nil {
		active, pmsg, cmd := s.active.Update(msg)
		if !active {
			if s.searchPrevActive {
				s.searchPrevActive = false
				s.active = s.searchWidget
			} else {
				s.active = nil
			}
		}
		return pmsg, cmd
	}

	return msg, nil
}

func (s *SubjectsCmdBar) HasSearchedAtLeastOneChar() bool {
	return s.searchWidget.(*cmdbar.SearchCmdBarModel).IsSearching() && len(s.GetSearchTerm()) > 0
}

func (s *SubjectsCmdBar) IsFocussed() bool {
	return s.active != nil
}

func (s *SubjectsCmdBar) GetSearchTerm() string {
	if searchBar, ok := s.searchWidget.(*cmdbar.SearchCmdBarModel); ok {
		return searchBar.GetSearchTerm()
	}
	return ""
}

func (s *SubjectsCmdBar) Shortcuts() []statusbar.Shortcut {
	if s.active == nil {
		return nil
	}
	return s.active.Shortcuts()
}

func NewCmdBar(deleter sr.SubjectDeleter) *SubjectsCmdBar {
	deleteMsgFunc := func(subject sr.Subject) string {
		message := subject.Name + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7571F9")).
			Bold(true).
			Render(" will be delete permanently")
		return message
	}

	deleteFunc := func(subject sr.Subject) tea.Cmd {
		return func() tea.Msg {
			return deleter.DeleteSubject(subject.Name, 1)
		}
	}

	subjectListingStartedNotifier := func(msg sr.SubjectListingStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Loading subjects")
		return true, cmd
	}
	subjectsListedNotifier := func(msg sr.SubjectsListedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	hideNotificationNotifier := func(msg cmdbar.HideNotificationMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	subjectDeletionStartedNotifier := func(msg sr.SubjectDeletionStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Deleting Subject")
		return true, cmd
	}
	subjectListingErrorMsg := func(msg sr.SubjetListingErrorMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Error listing subjects", msg.Err)
		return true, nil
	}
	subjectDeletedNotifier := func(msg sr.SubjectDeletedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Subject deleted")
		return true, func() tea.Msg {
			time.Sleep(2 * time.Second)
			return cmdbar.HideNotificationMsg{}
		}
	}
	notifierCmdBar := cmdbar.NewNotifierCmdBar()
	cmdbar.WithMapping(notifierCmdBar, subjectListingStartedNotifier)
	cmdbar.WithMapping(notifierCmdBar, subjectsListedNotifier)
	cmdbar.WithMapping(notifierCmdBar, hideNotificationNotifier)
	cmdbar.WithMapping(notifierCmdBar, subjectDeletionStartedNotifier)
	cmdbar.WithMapping(notifierCmdBar, subjectListingErrorMsg)
	cmdbar.WithMapping(notifierCmdBar, subjectDeletedNotifier)

	return &SubjectsCmdBar{
		cmdbar.NewDeleteCmdBar(deleteMsgFunc, deleteFunc),
		cmdbar.NewSearchCmdBar("Search subject by name"),
		notifierCmdBar,
		nil,
		false,
	}
}
