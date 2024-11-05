package cmdbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type Widget interface {
	View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string

	Update(msg tea.Msg) (bool, tea.Msg, tea.Cmd)

	Shortcuts() []statusbar.Shortcut
}

type CmdBar interface {
	View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string

	Update(msg tea.Msg) tea.Cmd

	Shortcuts() []statusbar.Shortcut
}
