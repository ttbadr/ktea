package sradmin

import tea "github.com/charmbracelet/bubbletea"

type SubjectLister interface {
	ListSubjects() tea.Msg
}

type SubjectCreationDetails struct {
	Subject string
	Schema  string
}

type SchemaCreator interface {
	CreateSchema(details SubjectCreationDetails) tea.Msg
}

type SubjectDeleter interface {
	DeleteSubject(subject string) tea.Msg
}

type VersionLister interface {
	ListVersions(subject string, versions []int) tea.Msg
}

type GlobalCompatibilityLister interface {
	ListGlobalCompatibility() tea.Msg
}
