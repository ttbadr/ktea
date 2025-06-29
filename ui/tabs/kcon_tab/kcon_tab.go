package kcon_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/kcons_page"
)

type Model struct {
	kconsPage *kcons_page.Model
	statusbar *statusbar.Model
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if m.statusbar != nil {
		views = append(views, m.statusbar.View(ktx, renderer))
	}

	views = append(views, m.kconsPage.View(ktx, renderer))

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	// always recreate the statusbar in case the active page might have changed
	m.statusbar = statusbar.New(m.kconsPage)

	return m.kconsPage.Update(msg)
}

func New(
	lister kcadmin.ConnectorLister,
	deleter kcadmin.ConnectorDeleter,
) (*Model, tea.Cmd) {
	kcPage, cmd := kcons_page.New(lister, deleter)
	return &Model{
		kconsPage: kcPage,
	}, cmd
}
