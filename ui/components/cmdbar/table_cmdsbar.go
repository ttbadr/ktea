package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type TableCmdsBar[T any] struct {
	notifierWidget   CmdBar
	deleteWidget     *DeleteCmdBar[T]
	searchWidget     CmdBar
	sortByCmdBar     *SortByCmdBar
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
	// when the notifier is active and has priority (because of a loading spinner) it should handle all msgs
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

	if _, ok := m.active.(*SearchCmdBar); ok {
		m.searchPrevActive = true
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			return m.handleSlash(msg)
		case "f2":
			if selection != nil {
				return m.handleF2(selection, msg)
			}
			return nil, nil
		case "f3":
			if selection != nil && m.sortByCmdBar != nil {
				return m.handleF3(msg, pmsg, cmd)
			}
			return pmsg, cmd
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

func (m *TableCmdsBar[T]) handleSlash(msg tea.Msg) (tea.Msg, tea.Cmd) {
	active, pmsg, cmd := m.searchWidget.Update(msg)
	if active {
		m.active = m.searchWidget
		m.deleteWidget.active = false
		if m.sortByCmdBar != nil {
			m.sortByCmdBar.active = false
		}
	} else {
		m.active = nil
	}
	return pmsg, cmd
}

func (m *TableCmdsBar[T]) handleF3(msg tea.Msg, pmsg tea.Msg, cmd tea.Cmd) (tea.Msg, tea.Cmd) {
	active, pmsg, cmd := m.sortByCmdBar.Update(msg)
	if !active {
		m.active = nil
	} else {
		m.active = m.sortByCmdBar
		m.searchWidget.(*SearchCmdBar).state = hidden
		m.deleteWidget.active = false
	}
	return pmsg, cmd
}

func (m *TableCmdsBar[T]) handleF2(selection *T, msg tea.Msg) (tea.Msg, tea.Cmd) {
	active, pmsg, cmd := m.deleteWidget.Update(msg)
	if active {
		m.active = m.deleteWidget
		m.deleteWidget.Delete(*selection)
		m.searchWidget.(*SearchCmdBar).state = hidden
		if m.sortByCmdBar != nil {
			m.sortByCmdBar.active = false
		}
	} else {
		m.active = nil
	}
	return pmsg, cmd
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
	sortByCmdBar *SortByCmdBar,
) *TableCmdsBar[T] {
	return &TableCmdsBar[T]{
		notifierCmdBar,
		deleteCmdBar,
		searchCmdBar,
		sortByCmdBar,
		notifierCmdBar,
		false,
	}
}
