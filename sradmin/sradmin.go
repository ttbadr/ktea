package sradmin

import tea "github.com/charmbracelet/bubbletea"

type SubjectLister interface {
	ListSubjects() tea.Msg
}

type SubjectCreationDetails struct {
	Subject string
	Schema  string
}

type SubjectCreator interface {
	CreateSchema(details SubjectCreationDetails) tea.Msg
}

type SubjectDeleter interface {
	DeleteSubject(subject string) tea.Msg
}
