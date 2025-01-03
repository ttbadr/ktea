package clusters_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/clusters_page"
	"ktea/ui/pages/create_cluster_page"
	"ktea/ui/pages/navigation"
	"strings"
)

type state int

type Model struct {
	state      state
	active     navigation.Page
	createPage navigation.Page
	config     *config.Config
	statusbar  *statusbar.Model
	ktx        *kontext.ProgramKtx
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	builder := strings.Builder{}
	if m.config.HasEnvs() && m.statusbar != nil {
		builder.WriteString(m.statusbar.View(ktx, renderer))
	}
	builder.WriteString(m.active.View(ktx, renderer))
	return builder.String()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.active == nil {
		return nil
	}
	switch msg := msg.(type) {
	case config.ClusterRegisteredMsg:
		listPage, _ := clusters_page.New(m.ktx)
		m.active = listPage
		m.statusbar = statusbar.New(m.active)
		return func() tea.Msg { return config.ReLoadConfig() }
	case config.ClusterDeletedMsg:
		if m.config.HasEnvs() {
			var listPage, _ = clusters_page.New(m.ktx)
			m.active = create_cluster_page.NewForm(m.ktx.Config, m.ktx)
			m.statusbar = statusbar.New(m.active)
			m.active = listPage
			// Emit switch cluster msg to make sure if the active cluster was deleted,
			// a switch to the new active cluster is made.
			return func() tea.Msg {
				return clusters_page.ClusterSwitchedMsg{
					Cluster: m.ktx.Config.ActiveCluster(),
				}
			}
		} else {
			m.active = create_cluster_page.NewForm(m.ktx.Config, m.ktx)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.active, _ = clusters_page.New(m.ktx)
			return nil
		case "ctrl+n":
			m.active = create_cluster_page.NewForm(m.ktx.Config, m.ktx)
			return nil
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
			}
			if selectedCluster.SchemaRegistry != nil {
				formValues.SrEnabled = true
				formValues.SrUrl = selectedCluster.SchemaRegistry.Url
				formValues.SrUsername = selectedCluster.SchemaRegistry.Username
				formValues.SrPassword = selectedCluster.SchemaRegistry.Password
			}
			m.active = create_cluster_page.NewEditForm(
				m.ktx.Config,
				m.ktx,
				formValues,
			)
			return nil
		}
	}
	m.statusbar = statusbar.New(m.active)
	return m.active.Update(msg)
}

func New(ktx *kontext.ProgramKtx) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	m := Model{}
	m.ktx = ktx
	m.config = ktx.Config
	if m.config.HasEnvs() {
		var listPage, c = clusters_page.New(ktx)
		cmd = c
		m.active = listPage
		m.statusbar = statusbar.New(m.active)
	} else {
		m.active = create_cluster_page.NewForm(m.ktx.Config, m.ktx)
	}
	return &m, cmd
}
