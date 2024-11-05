package clusters_page

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/config"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"strings"
)

type state int

const (
	HIDDEN              state = 0
	DELETE_CONFIRMATION state = 1
	SPINNING            state = 2
)

type CmdBar struct {
	state         state
	deleteConfirm *huh.Confirm
	ktx           *kontext.ProgramKtx
}

func (c *CmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	builder := strings.Builder{}
	if c.state == DELETE_CONFIRMATION {
		builder.WriteString(styles.CmdBar.Render(c.deleteConfirm.View()))
	}
	builder.WriteString("\n")
	return renderer.Render(builder.String())
}

// Update returns the tea.Msg if it is not being handled or nil if it is
func (c *CmdBar) Update(msg tea.Msg, selectedCluster string) (tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+d" {
			c.state = DELETE_CONFIRMATION
			// TODO: clean-up style
			c.deleteConfirm = newDeleteConfirm()
			c.deleteConfirm.Title(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Render("🗑️  "+selectedCluster) + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7571F9")).
				Bold(true).
				Render(" will be delete permanently")).
				Focus()
			return nil, nil
		} else if msg.String() == "enter" {
			confirm, _ := c.deleteConfirm.Update(msg)
			if f, ok := confirm.(*huh.Confirm); ok {
				c.deleteConfirm = f
			}
			c.state = HIDDEN
			if c.deleteConfirm.GetValue().(bool) {
				return nil, func() tea.Msg {
					c.ktx.Config.DeleteCluster(selectedCluster)
					return config.ClusterDeletedMsg{Name: selectedCluster}
				}
			}

		} else if c.state == DELETE_CONFIRMATION {
			confirm, _ := c.deleteConfirm.Update(msg)
			if f, ok := confirm.(*huh.Confirm); ok {
				c.deleteConfirm = f
			}
			return nil, nil
		}
	}
	return msg, nil
}

func (c *CmdBar) IsFocused() bool {
	return c.state == DELETE_CONFIRMATION
}

func newDeleteConfirm() *huh.Confirm {
	return huh.NewConfirm().
		Inline(true).
		Affirmative("Delete!").
		Negative("Cancel.").
		WithKeyMap(&huh.KeyMap{
			Confirm: huh.ConfirmKeyMap{
				Submit: key.NewBinding(key.WithKeys("enter")),
				Toggle: key.NewBinding(key.WithKeys("h", "l", "right", "left")),
				Accept: key.NewBinding(key.WithKeys("d")),
				Reject: key.NewBinding(key.WithKeys("c")),
			},
		}).(*huh.Confirm)
}

func NewCmdBar(ktx *kontext.ProgramKtx) *CmdBar {
	return &CmdBar{
		deleteConfirm: newDeleteConfirm(),
		ktx:           ktx,
	}
}
