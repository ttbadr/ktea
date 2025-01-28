package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type TableCmdsBar[T any] struct {
	deleteWidget     *DeleteCmdBar[T]
	searchWidget     CmdBar
	notifierWidget   CmdBar
	active           CmdBar
	searchPrevActive bool
}

type NotifierConfigurerFunc func(notifier *NotifierCmdBar)

func (m *TableCmdsBar[T]) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.active != nil {
		return m.active.View(ktx, renderer)
	}
	return ""
}

func (m *TableCmdsBar[T]) Update(msg tea.Msg, selection *T) (tea.Msg, tea.Cmd) {
	// when the notifier is active and has priority (because of a loading spinner)
	if m.active == m.notifierWidget {
		if m.notifierWidget.(*NotifierCmdBar).Notifier.HasPriority() {
			active, pmsg, cmd := m.active.Update(msg)
			if !active {
				m.active = nil
			}
			return pmsg, cmd
		}
	}

	// notifier was not actively spinning
	// if it is able to handle the msg it will return nil and the processing can stop
	active, pmsg, cmd := m.notifierWidget.Update(msg)
	if active && pmsg == nil {
		m.active = m.notifierWidget
		return msg, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			m.active = m.searchWidget
			active, pmsg, cmd := m.active.Update(msg)
			if !active {
				m.active = nil
			}
			return pmsg, cmd
		case "f2":
			if selection != nil {
				if _, ok := m.active.(*SearchCmdBar); ok {
					m.searchPrevActive = true
				}
				m.deleteWidget.Delete(*selection)
				_, pmsg, cmd := m.deleteWidget.Update(msg)
				m.active = m.deleteWidget
				return pmsg, cmd
			}
			return nil, nil
		}
	}

	if m.active != nil {
		active, pmsg, cmd := m.active.Update(msg)
		if !active {
			if m.searchPrevActive {
				m.searchPrevActive = false
				m.active = m.searchWidget
			} else {
				m.active = nil
			}
		}
		return pmsg, cmd
	}

	return msg, nil
}

func (m *TableCmdsBar[T]) HasSearchedAtLeastOneChar() bool {
	return m.searchWidget.(*SearchCmdBar).IsSearching() && len(m.GetSearchTerm()) > 0
}

func (m *TableCmdsBar[T]) IsFocussed() bool {
	return m.active != nil && m.active.IsFocussed()
}

func (m *TableCmdsBar[T]) GetSearchTerm() string {
	if searchBar, ok := m.searchWidget.(*SearchCmdBar); ok {
		return searchBar.GetSearchTerm()
	}
	return ""
}

func (m *TableCmdsBar[T]) Shortcuts() []statusbar.Shortcut {
	if m.active == nil {
		return nil
	}
	return m.active.Shortcuts()
}

func NewTableCmdsBar[T any](
	deleteCmdBar *DeleteCmdBar[T],
	searchCmdBar *SearchCmdBar,
	notifierCmdBar *NotifierCmdBar,
) *TableCmdsBar[T] {
	return &TableCmdsBar[T]{
		deleteCmdBar,
		searchCmdBar,
		notifierCmdBar,
		nil,
		false,
	}
}
