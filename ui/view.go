package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
)

type View interface {
	View(ktx *kontext.ProgramKtx, renderer *Renderer) string

	Update(msg tea.Msg) tea.Cmd
}
