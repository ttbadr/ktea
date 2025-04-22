package clusters_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/clusters_page"
	"ktea/ui/pages/create_cluster_page"
	"ktea/ui/pages/nav"
)

type state int

type Model struct {
	state       state
	active      nav.Page
	createPage  nav.Page
	config      *config.Config
	statusbar   *statusbar.Model
	ktx         *kontext.ProgramKtx
	connChecker kadmin.ConnChecker
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if m.config.HasClusters() && m.statusbar != nil {
		views = append(views, m.statusbar.View(ktx, renderer))
	}

	views = append(views, m.active.View(ktx, renderer))

	return ui.JoinVertical(
		lipgloss.Top,
		views...,
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.active == nil {
		return nil
	}
	switch msg := msg.(type) {
	case config.ClusterRegisteredMsg:
		listPage, _ := clusters_page.New(m.ktx, m.connChecker)
		m.active = listPage
		m.statusbar = statusbar.New(m.active)
		return func() tea.Msg {
			return config.ReLoadConfig()
		}
	case config.ClusterDeletedMsg:
		if m.config.HasClusters() {
			cmd := m.active.Update(msg)
			return tea.Batch(cmd, func() tea.Msg {
				// Emit Msg to trigger table recreation without deleted cluster
				return clusters_page.ClusterSwitchedMsg{
					Cluster: m.ktx.Config.ActiveCluster(),
				}
			})
		} else {
			m.active = create_cluster_page.NewForm(m.connChecker, m.ktx.Config, m.ktx)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.active, _ = clusters_page.New(m.ktx, m.connChecker)
		case "ctrl+n":
			m.active = create_cluster_page.NewForm(m.connChecker, m.ktx.Config, m.ktx)
		case "ctrl+e":
			clusterName := m.active.(*clusters_page.Model).SelectedCluster()
			selectedCluster := m.ktx.Config.FindClusterByName(*clusterName)
			formValues := &create_cluster_page.FormValues{
				Name:  selectedCluster.Name,
				Color: selectedCluster.Color,
				Host:  selectedCluster.BootstrapServers[0],
			}
			if selectedCluster.SASLConfig != nil {
				formValues.SecurityProtocol = selectedCluster.SASLConfig.SecurityProtocol
				formValues.Username = selectedCluster.SASLConfig.Username
				formValues.Password = selectedCluster.SASLConfig.Password
				formValues.AuthMethod = config.SASLAuthMethod
				formValues.SSLEnabled = selectedCluster.SSLEnabled
			}
			if selectedCluster.SchemaRegistry != nil {
				formValues.SrEnabled = true
				formValues.SrUrl = selectedCluster.SchemaRegistry.Url
				formValues.SrUsername = selectedCluster.SchemaRegistry.Username
				formValues.SrPassword = selectedCluster.SchemaRegistry.Password
			}
			m.active = create_cluster_page.NewEditForm(
				m.connChecker,
				m.ktx.Config,
				m.ktx,
				formValues,
			)
		}
	}

	// always recreate the statusbar in case the active page might have changed
	m.statusbar = statusbar.New(m.active)

	return m.active.Update(msg)
}

func New(
	ktx *kontext.ProgramKtx,
	connChecker kadmin.ConnChecker,
) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	m := Model{}
	m.connChecker = connChecker
	m.ktx = ktx
	m.config = ktx.Config
	if m.config.HasClusters() {
		var listPage, c = clusters_page.New(ktx, m.connChecker)
		cmd = c
		m.active = listPage
		m.statusbar = statusbar.New(m.active)
	} else {
		m.active = create_cluster_page.NewForm(m.connChecker, m.ktx.Config, m.ktx)
	}
	return &m, cmd
}
