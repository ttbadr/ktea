package schema_details_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
)

type CmdBar struct {
	notifierWidget cmdbar.CmdBar
	active         cmdbar.CmdBar
}

func (c *CmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if c.active != nil {
		return c.active.View(ktx, renderer)
	}
	return ""
}

func (c *CmdBar) Update(msg tea.Msg) (tea.Msg, tea.Cmd) {
	if c.active != nil {
		active, pmsg, cmd := c.active.Update(msg)
		if !active {
			c.active = nil
		}
		return pmsg, cmd
	}
	return msg, nil
}

func NewCmdBar() *CmdBar {
	schemaListingStartedNotifier := func(msg sradmin.SchemaListingStarted, m *notifier.Model) (bool, tea.Cmd) {
		cmd := m.SpinWithLoadingMsg("Loading schema")
		return true, cmd
	}
	schemaListedNotifier := func(msg sradmin.SchemasListed, m *notifier.Model) (bool, tea.Cmd) {
		m.Idle()
		return false, nil
	}
	notifierCmdBar := cmdbar.NewNotifierCmdBar("schema-details-cmd-bar")
	cmdbar.WithMsgHandler(notifierCmdBar, schemaListingStartedNotifier)
	cmdbar.WithMsgHandler(notifierCmdBar, schemaListedNotifier)
	return &CmdBar{notifierWidget: notifierCmdBar, active: notifierCmdBar}
}
