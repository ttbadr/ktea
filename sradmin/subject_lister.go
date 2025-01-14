package sradmin

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"sync"
)

type SubjectsListedMsg struct {
	Subjects []Subject
}

type SubjectListingErrorMsg struct {
	Err error
}

type SubjectListingStartedMsg struct {
	subjects chan []Subject
	err      chan error
}

func (msg *SubjectListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case subjects := <-msg.subjects:
		return SubjectsListedMsg{subjects}
	case err := <-msg.err:
		log.Error("Failed to fetch subjects", "err", err)
		return SubjectListingErrorMsg{err}
	}
}

func (s *SrAdmin) ListSubjects() tea.Msg {
	subjectsChan := make(chan []Subject)
	errChan := make(chan error)

	go s.doListSubject(subjectsChan, errChan)

	return SubjectListingStartedMsg{subjectsChan, errChan}
}

type Subject struct {
	Name     string
	Versions []int
}

func (s *SrAdmin) doListSubject(subjectsChan chan []Subject, errChan chan error) {
	maybeIntroduceLatency()

	subjects, err := s.client.GetSubjects()
	if err != nil {
		errChan <- err
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([][]int, len(subjects))

	for i, subject := range subjects {

		wg.Add(1)

		go func(index int, subject string) {
			defer wg.Done()
			versions, err := s.client.GetSchemaVersions(subject)
			if err != nil {
				errChan <- fmt.Errorf("failed to get versions for subject %s: %w", subject, err)
				return
			}
			mu.Lock()
			results[i] = versions
			mu.Unlock()
		}(i, subject)

	}

	wg.Wait()

	var subjectPtrs []Subject
	for i, str := range subjects {
		ptr := str
		subjectPtrs = append(subjectPtrs, Subject{
			Name:     ptr,
			Versions: results[i],
		})
	}
	subjectsChan <- subjectPtrs
}
