package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
)

type MockSrAdmin struct {
	GetSchemaByIdFunc func(id int) tea.Msg
}

func (m *MockSrAdmin) ListGlobalCompatibility() tea.Msg {
	return nil
}

func (m *MockSrAdmin) GetSchemaById(id int) tea.Msg {
	if m.GetSchemaByIdFunc != nil {
		return m.GetSchemaByIdFunc(id)
	}
	return nil
}

func (m *MockSrAdmin) DeleteSubject(subject string) tea.Msg {
	return nil
}

func (m *MockSrAdmin) ListSubjects() tea.Msg {
	return nil
}

func (m *MockSrAdmin) CreateSchema(details SubjectCreationDetails) tea.Msg {
	return nil
}

func (m *MockSrAdmin) ListVersions(subject string, versions []int) tea.Msg {
	return nil
}

func NewMock() *MockSrAdmin {
	return &MockSrAdmin{}
}
