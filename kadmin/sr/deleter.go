package sr

import tea "github.com/charmbracelet/bubbletea"

type SubjectDeletedMsg struct{}

type SubjectDeletionErrorMsg struct {
	Err error
}

type SubjectDeletionStartedMsg struct {
	Deleted chan bool
	Err     chan error
}

func (msg *SubjectDeletionStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-msg.Deleted:
		return SubjectDeletedMsg{}
	case err := <-msg.Err:
		return SubjectDeletionErrorMsg{err}
	}
}

func (s *SrAdmin) DeleteSubject(subject string, version int) tea.Msg {
	deletedChan := make(chan bool)
	errChan := make(chan error)
	go func(deletedChan chan bool, errChan chan error) {
		err := s.client.DeleteSubjectByVersion(subject, version, true)
		if err != nil {
			errChan <- err
			return
		} else {
			deletedChan <- true
			return
		}
	}(deletedChan, errChan)
	return SubjectDeletionStartedMsg{
		deletedChan,
		errChan,
	}
}
