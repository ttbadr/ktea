package cgroups_topics_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/notifier"
)

type CmdBar struct {
	notifier *notifier.Model
}

func (c *CmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return c.notifier.View(ktx, renderer)
}

func (c *CmdBar) Update(msg tea.Msg) tea.Cmd {
	return c.notifier.Update(msg)
}

func NewCmdBar() *CmdBar {
	return &CmdBar{
		notifier: notifier.New(),
	}
}
