package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type state int

const (
	hidden    state = 0
	searching state = 1
	searched  state = 2
)

type SearchCmdBarModel struct {
	searchInput *huh.Input
	state       state
	placeholder string
}

func (s *SearchCmdBarModel) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{
			Name:       "Confirm",
			Keybinding: "enter",
		},
		{
			Name:       "Cancel",
			Keybinding: "esc",
		},
		{
			Name:       "Toggle",
			Keybinding: "/",
		},
	}
}

func (s *SearchCmdBarModel) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return renderer.Render(styles.CmdBar.Render(s.searchInput.View()))
}

func (s *SearchCmdBarModel) Update(msg tea.Msg) (bool, tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			s.searchInput.Focus()
			if s.state == searching {
				s.state = hidden
			} else {
				s.state = searching
			}
			return s.state == searching || s.state == searched, nil, nil
		case "enter":
			var pmsg tea.Msg
			if s.state == searched {
				pmsg = msg
			}
			s.searchInput.Blur()
			if s.GetSearchTerm() == "" {
				s.state = hidden
			} else {
				s.state = searched
			}
			return s.state == searching || s.state == searched, pmsg, nil
		case "esc":
			if s.state == searching {
				s.searchInput.Blur()
				s.state = hidden
				s.searchInput = newSearchInput(s.placeholder)
			} else if s.state == searched {

			} else {
				s.state = hidden
			}
			return s.state == searching || s.state == searched, nil, nil
		default:
			input, _ := s.searchInput.Update(msg)
			if i, ok := input.(*huh.Input); ok {
				s.searchInput = i
			}
		}
	}
	return s.state == searching || s.state == searched, msg, nil
}

func (s *SearchCmdBarModel) GetSearchTerm() string {
	return s.searchInput.GetValue().(string)
}

func (s *SearchCmdBarModel) IsSearching() bool {
	return s.state == searching
}

func newSearchInput(placeholder string) *huh.Input {
	searchInput := huh.NewInput().
		Placeholder(placeholder)
	searchInput.Init()
	return searchInput
}

func NewSearchCmdBar(placeholder string) Widget {
	return &SearchCmdBarModel{
		state:       hidden,
		searchInput: newSearchInput(placeholder),
		placeholder: placeholder,
	}
}
