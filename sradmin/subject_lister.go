package sradmin

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"slices"
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

// AwaitCompletion return
// a SubjectsListedMsg upon success
// or SubjectListingErrorMsg upon failure.
func (msg *SubjectListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case subjects := <-msg.subjects:
		return SubjectsListedMsg{subjects}
	case err := <-msg.err:
		log.Error("Failed to fetch subjects", "err", err)
		return SubjectListingErrorMsg{err}
	}
}

func (s *DefaultSrAdmin) ListSubjects() tea.Msg {
	subjectsChan := make(chan []Subject)
	errChan := make(chan error)

	go s.doListSubject(subjectsChan, errChan)

	return SubjectListingStartedMsg{subjectsChan, errChan}
}

type Subject struct {
	Name          string
	Versions      []int
	Compatibility string
}

func (s *Subject) LatestVersion() int {
	return slices.Max(s.Versions)
}

func (s *DefaultSrAdmin) doListSubject(subjectsChan chan []Subject, errChan chan error) {
	maybeIntroduceLatency()

	subjects, err := s.client.GetSubjects()
	if err != nil {
		errChan <- err
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	versionResults := make([][]int, len(subjects))
	compResults := make([]string, len(subjects))

	for i, subject := range subjects {

		wg.Add(2)

		go func(index int, subject string) {
			defer wg.Done()
			versions, err := s.client.GetSchemaVersions(subject)
			if err != nil {
				errChan <- fmt.Errorf("failed to get versions for subject %s: %w", subject, err)
				return
			}
			mu.Lock()
			versionResults[i] = versions
			mu.Unlock()
		}(i, subject)

		go func(index int, subject string) {
			defer wg.Done()
			comp, err := s.client.GetCompatibilityLevel(subject, true)
			if err != nil {
				errChan <- fmt.Errorf("failed to get compatibility for subject %s: %w", subject, err)
				return
			}
			mu.Lock()
			compResults[i] = comp.String()
			mu.Unlock()
		}(i, subject)

	}

	wg.Wait()

	var subjectPtrs []Subject
	for i, str := range subjects {
		ptr := str
		subjectPtrs = append(subjectPtrs, Subject{
			Name:          ptr,
			Versions:      versionResults[i],
			Compatibility: compResults[i],
		})
	}

	s.mu.Lock()
	s.subjects = subjectPtrs
	s.mu.Unlock()

	subjectsChan <- subjectPtrs
}
