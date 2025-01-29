package kadmin

import tea "github.com/charmbracelet/bubbletea"

type CGroupDeleter interface {
	DeleteCGroup(name string) tea.Msg
}

type CGroupDeletionStartedMsg struct {
	Deleted chan bool
	Err     chan error
}

func (c *CGroupDeletionStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.Deleted:
		return CGroupDeletedMsg{}
	case err := <-c.Err:
		return CGroupDeletionErrMsg{Err: err}
	}
}

type CGroupDeletionErrMsg struct {
	Err error
}

type CGroupDeletedMsg struct {
}

func (ka *SaramaKafkaAdmin) DeleteCGroup(name string) tea.Msg {
	errChan := make(chan error)
	deletedChan := make(chan bool)

	go ka.doDeleteCGroup(name, deletedChan, errChan)

	return CGroupDeletionStartedMsg{
		Deleted: deletedChan,
		Err:     errChan,
	}
}

func (ka *SaramaKafkaAdmin) doDeleteCGroup(
	name string,
	deletedChan chan bool,
	errChan chan error,
) {
	err := ka.admin.DeleteConsumerGroup(name)
	if err != nil {
		errChan <- err
		return
	}
	deletedChan <- true
}
