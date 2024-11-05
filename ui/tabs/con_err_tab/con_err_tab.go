package con_err_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/pages/error_page"
	"ktea/ui/tabs"
)

type Model struct {
	cluster *config.Cluster
	err     error
}

type RetryClusterConnectionMsg struct {
	Cluster *config.Cluster
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return error_page.New(m.err).View(ktx, renderer)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "f5" {
			return func() tea.Msg {
				return RetryClusterConnectionMsg{m.cluster}
			}
		}
	}
	return nil
}

func New(err error, cluster *config.Cluster) (tabs.TabController, tea.Cmd) {
	return &Model{err: err, cluster: cluster}, nil
}
