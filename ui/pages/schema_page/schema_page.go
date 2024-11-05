package schema_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type Model struct {
}

func (m Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return ""
}

func (m Model) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (m Model) Shortcuts() []statusbar.Shortcut {
	return nil
}

func (m Model) Title() string {
	return "Subjects"
}

func New(subject string) (*Model, tea.Cmd) {
	return &Model{}, nil
}
