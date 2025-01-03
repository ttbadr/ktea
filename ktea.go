package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/tab"
	"ktea/ui/pages/clusters_page"
	"ktea/ui/tabs"
	"ktea/ui/tabs/cgroups_tab"
	"ktea/ui/tabs/clusters_tab"
	"ktea/ui/tabs/con_err_tab"
	"ktea/ui/tabs/loading_tab"
	"ktea/ui/tabs/sr_tab"
	"ktea/ui/tabs/topics_tab"
	"os"
	"time"
)

type Model struct {
	tabs           tab.Model
	tabCtrl        tabs.TabController
	ktx            *kontext.ProgramKtx
	activeTab      int
	topicsTabCtrl  *topics_tab.Model
	cgroupsTabCtrl *cgroups_tab.Model
	ka             *kadmin.SaramaKafkaAdmin
	sra            *sradmin.SrAdmin
	renderer       *ui.Renderer
}

// RetryClusterConnectionMsg is an internal Msg
// to actually retry the cluster connection
type RetryClusterConnectionMsg struct {
	Cluster *config.Cluster
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		return config.LoadedMsg{config.New(config.NewDefaultConfigIO())}
	}, tea.WindowSize())
}

func (m *Model) View() string {
	m.ktx = kontext.WithNewAvailableDimensions(m.ktx)
	if m.renderer == nil {
		m.renderer = ui.NewRenderer(m.ktx)
	}

	var views []string
	logoView := m.renderer.Render("   ___        \n |/ |  _   _.\n |\\ | (/_ (_|  v0.1.0")
	views = append(views, logoView)

	tabsView := m.tabs.View(m.ktx, m.renderer)
	views = append(views, tabsView)

	if m.tabCtrl != nil {
		view := m.tabCtrl.View(m.ktx, m.renderer)
		views = append(views, view)
	}

	return ui.JoinVerticalSkipEmptyViews(views...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case config.ClusterRegisteredMsg:
		// if the active cluster has been updated it needs to be reloaded
		if msg.Cluster.Active {
			m.activateCluster(msg.Cluster)
			// keep clusters tab focussed after recreating tabs
			if msg.Cluster.HasSchemaRegistry() {
				m.tabs.GoToTab(tabs.ClustersTab)
			} else {
				m.tabs.GoToTab(tabs.ClustersTab)
			}

		}
	case con_err_tab.RetryClusterConnectionMsg:
		m.tabCtrl, cmd = loading_tab.New()
		return m, tea.Batch(cmd, func() tea.Msg {
			return RetryClusterConnectionMsg{msg.Cluster}
		})
	case RetryClusterConnectionMsg:
		return m.initTopicsTabOrError(msg.Cluster)
	case config.LoadedMsg:
		m.ktx.Config = msg.Config
		if m.ktx.Config.HasEnvs() {
			m.tabs.GoToTab(tabs.TopicsTab)
			return m.initTopicsTabOrError(msg.Config.ActiveCluster())
		} else {
			t, c := clusters_tab.New(m.ktx)
			m.tabCtrl = t
			m.tabs.GoToTab(tabs.ClustersTab)
			return m, c
		}
	case clusters_page.ClusterSwitchedMsg:
		m.activateCluster(msg.Cluster)
		// make sure we stay on the clusters tab because,
		// tabs were recreated due to cluster switch,
		// which might have introduced or removed the schema-registry tab
		if msg.Cluster.HasSchemaRegistry() {
			m.tabs.GoToTab(3)
		} else {
			m.tabs.GoToTab(2)
		}
	case tea.WindowSizeMsg:
		m.onWindowSizeUpdated(msg)
	}

	// if no clusters configured,
	// do not allow to move away from create cluster form
	if m.ktx.Config != nil && m.ktx.Config.HasEnvs() {
		m.tabs.Update(msg)
	}
	if m.tabs.ActiveTab() != m.activeTab {
		m.activeTab = m.tabs.ActiveTab()
		switch m.activeTab {
		case 0:
			m.topicsTabCtrl, cmd = topics_tab.New(m.ktx, m.ka)
			m.tabCtrl = m.topicsTabCtrl
			return m, cmd
		case 1:
			m.cgroupsTabCtrl, cmd = cgroups_tab.New(m.ka, m.ka)
			m.tabCtrl = m.cgroupsTabCtrl
			return m, cmd
		case 2:
			if m.ktx.Config.ActiveCluster().HasSchemaRegistry() {
				t, cmd := sr_tab.New(m.sra, m.sra, m.sra, m.ktx)
				m.tabCtrl = t
				return m, cmd
			}
			fallthrough
		case 3:
			t, cmd := clusters_tab.New(m.ktx)
			m.tabCtrl = t
			return m, cmd
		}
	}

	if m.tabCtrl == nil {
		m.tabCtrl, cmd = loading_tab.New()
		return m, cmd
	}
	return m, m.tabCtrl.Update(msg)
}

func (m *Model) recreateTabs(cluster *config.Cluster) {
	if cluster.HasSchemaRegistry() {
		m.tabs = tab.New("Topics", "Consumer Groups", "Schema Registry", "Clusters")
		tabs.ClustersTab = 3
	} else {
		m.tabs = tab.New("Topics", "Consumer Groups", "Clusters")
		tabs.ClustersTab = 2
	}
}

// activateCluster creates the kadmin.Model and kadmin.SrAdmin
// based on the given cluster
func (m *Model) activateCluster(cluster *config.Cluster) error {
	var saslConfig *kadmin.SASLConfig
	if cluster.SASLConfig != nil {
		saslConfig = &kadmin.SASLConfig{
			Username: cluster.SASLConfig.Username,
			Password: cluster.SASLConfig.Password,
			Protocol: kadmin.SSL,
		}
	}

	connDetails := kadmin.ConnectionDetails{
		BootstrapServers: cluster.BootstrapServers,
		SASLConfig:       saslConfig,
	}
	if ka, err := kadmin.New(connDetails); err != nil {
		return err
	} else {
		m.ka = ka
	}

	if cluster.HasSchemaRegistry() {
		m.sra = sradmin.NewSrAdmin(m.ktx)
	}

	m.recreateTabs(cluster)

	return nil
}

func (m *Model) onWindowSizeUpdated(msg tea.WindowSizeMsg) {
	m.ktx.WindowWidth = msg.Width
	m.ktx.WindowHeight = msg.Height
	m.ktx.AvailableHeight = msg.Height
}

func (m *Model) initTopicsTabOrError(cluster *config.Cluster) (tea.Model, tea.Cmd) {
	if err := m.activateCluster(cluster); err != nil {
		var cmd tea.Cmd
		m.tabCtrl, cmd = con_err_tab.New(err, cluster)
		return m, cmd
	} else {
		var cmd tea.Cmd
		m.topicsTabCtrl, cmd = topics_tab.New(m.ktx, m.ka)
		m.tabCtrl = m.topicsTabCtrl
		return m, cmd
	}
}

func NewModel() *Model {
	return &Model{
		ktx: kontext.New(),
	}
}

func main() {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	var fileErr error
	newConfigFile, fileErr := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if fileErr == nil {
		log.SetOutput(newConfigFile)
		log.SetTimeFormat(time.Kitchen)
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
		log.Debug("Logging to debug.log")
		log.Info("started")
		if _, err := p.Run(); err != nil {
			log.Fatal("Failed starting the TUI", err)
		}
	}
}
