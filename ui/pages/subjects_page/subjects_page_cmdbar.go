package subjects_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
)

type SubjectsCmdBar struct {
	deleteWidget     *cmdbar.DeleteCmdBarModel[sradmin.Subject]
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

func (s *SubjectsCmdBar) Update(msg tea.Msg, selectedSubject sradmin.Subject) (tea.Msg, tea.Cmd) {
	// when the notifier is active and has priority (because of a loading spinner)
	if s.active == s.notifierWidget {
		if s.notifierWidget.(*cmdbar.NotifierCmdBar).Notifier.HasPriority() {
			active, pmsg, cmd := s.active.Update(msg)
			if !active {
				s.active = nil
			}
			return pmsg, cmd
		}
	}

	switch msg := msg.(type) {
	case sradmin.SubjectListingStartedMsg, sradmin.SubjectDeletionStartedMsg:
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
		case "f2":
			if _, ok := s.active.(*cmdbar.SearchCmdBarModel); ok {
				s.searchPrevActive = true
			}
			s.deleteWidget.Delete(selectedSubject)
			_, pmsg, cmd := s.deleteWidget.Update(msg)
			s.active = s.deleteWidget
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

func NewCmdBar(deleter sradmin.SubjectDeleter) *SubjectsCmdBar {
	deleteMsgFunc := func(subject sradmin.Subject) string {
		message := subject.Name + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7571F9")).
			Bold(true).
			Render(" will be deleted permanently")
		return message
	}

	deleteFunc := func(subject sradmin.Subject) tea.Cmd {
		return func() tea.Msg {
			return deleter.DeleteSubject(subject.Name)
		}
	}

	subjectListingStartedNotifier := func(msg sradmin.SubjectListingStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Loading subjects")
		return true, cmd
	}
	subjectsListedNotifier := func(msg sradmin.SubjectsListedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	// TODO maybe we can move this into the notifier
	hideNotificationNotifier := func(msg notifier.HideNotificationMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	subjectDeletionStartedNotifier := func(msg sradmin.SubjectDeletionStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Deleting Subject " + msg.Subject)
		return true, cmd
	}
	subjectListingErrorMsg := func(msg sradmin.SubjectListingErrorMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Error listing subjects", msg.Err)
		return true, nil
	}
	subjectDeletedNotifier := func(msg sradmin.SubjectDeletedMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowSuccessMsg("Subject deleted")
		return true, m.AutoHideCmd()
	}
	subjectDeletionErrorNotifier := func(msg sradmin.SubjectDeletionErrorMsg, m *notifier.Model) (bool, tea.Cmd) {
		m.ShowErrorMsg("Failed to delete subject", msg.Err)
		return true, m.AutoHideCmd()
	}
	notifierCmdBar := cmdbar.NewNotifierCmdBar()
	cmdbar.WithMapping(notifierCmdBar, subjectListingStartedNotifier)
	cmdbar.WithMapping(notifierCmdBar, subjectsListedNotifier)
	cmdbar.WithMapping(notifierCmdBar, hideNotificationNotifier)
	cmdbar.WithMapping(notifierCmdBar, subjectDeletionStartedNotifier)
	cmdbar.WithMapping(notifierCmdBar, subjectListingErrorMsg)
	cmdbar.WithMapping(notifierCmdBar, subjectDeletedNotifier)
	cmdbar.WithMapping(notifierCmdBar, subjectDeletionErrorNotifier)

	return &SubjectsCmdBar{
		cmdbar.NewDeleteCmdBar(deleteMsgFunc, deleteFunc),
		cmdbar.NewSearchCmdBar("Search subject by name"),
		notifierCmdBar,
		nil,
		false,
	}
}
