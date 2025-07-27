package create_cluster_page

import (
	"errors"
	"fmt"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kcadmin"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/border"
	"ktea/ui/components/cmdbar"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"reflect"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type selectionState int

type formState int

type Option func(m *Model)

const (
	noneSelected      selectionState  = 0
	saslSelected      selectionState  = 1
	nothingSelected   selectionState  = 2
	sslEnabledState   selectionState  = 3
	sslDisabledState  selectionState  = 4
	none              formState       = 0
	loading           formState       = 1
	notifierCmdbarTag                 = "upsert-cluster-page"
	cTab              border.TabLabel = "f4"
	srTab             border.TabLabel = "f5"
	kcTab             border.TabLabel = "f6"
)

type Model struct {
	NavBack            ui.NavBack
	form               *huh.Form // the active form
	state              formState
	srForm             *huh.Form
	cForm              *huh.Form
	kForm              *huh.Form
	clusterValues      *clusterValues
	clusterToEdit      *config.Cluster
	notifierCmdBar     *cmdbar.NotifierCmdBar
	ktx                *kontext.ProgramKtx
	clusterRegisterer  config.ClusterRegisterer
	kConnChecker       kadmin.ConnChecker
	srConnChecker      sradmin.ConnChecker
	authSelectionState selectionState
	sslSelectionState  selectionState
	preEditName        *string
	shortcuts          []statusbar.Shortcut
	title              string
	border             *border.Model
	kcModel            *UpsertKcModel
}

type clusterValues struct {
	name                  string
	color                 string
	host                  string
	authMethod            config.AuthMethod
	securityProtocol      config.SecurityProtocol
	sslEnabled            bool
	tlsCertFile           string
	tlsKeyFile            string
	tlsCAFile             string
	tlsInsecureSkipVerify bool
	username              string
	password              string
	srUrl                 string
	srUsername            string
	srPassword            string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	if !ktx.Config.HasClusters() {
		builder := strings.Builder{}
		builder.WriteString("\n")
		builder.WriteString(lipgloss.NewStyle().PaddingLeft(1).Render("No clusters configured. Please create your first cluster!"))
		builder.WriteString("\n")
		views = append(views, renderer.Render(builder.String()))
	}

	notifierView := m.notifierCmdBar.View(ktx, renderer)

	deleteCmdbar := ""
	if m.kcModel.deleteCmdbar.IsFocussed() {
		deleteCmdbar = m.kcModel.deleteCmdbar.View(ktx, renderer)
	}

	var mainView string
	if m.border.ActiveTab() == kcTab {
		mainView = renderer.Render(lipgloss.
			NewStyle().
			Width(ktx.WindowWidth - 2).
			Render(m.kcModel.View(ktx, renderer)))
	} else {
		mainView = renderer.RenderWithStyle(m.form.View(), styles.Form)
	}

	mainView = m.border.View(lipgloss.NewStyle().
		PaddingBottom(ktx.AvailableHeight - 1).
		Render(mainView))

	views = append(views, deleteCmdbar, notifierView, mainView)

	return ui.JoinVertical(lipgloss.Top, views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	log.Debug("Received Update", "msg", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	activeTab := m.border.ActiveTab()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if activeTab == kcTab {
				return m.kcModel.Update(msg)
			}
			m.title = "Clusters"
			return m.NavBack()
		case "ctrl+r":
			m.clusterValues = &clusterValues{}
			if activeTab == cTab {
				m.cForm = m.createCForm()
				m.form = m.cForm
				m.authSelectionState = noneSelected
				m.sslSelectionState = noneSelected
			} else {
				m.srForm = m.createSrForm()
				m.form = m.srForm
			}
		case "f4":
			m.form = m.cForm
			m.border.GoTo("f4")
			return nil
		case "f5":
			m.border.GoTo("handling f5")
			if m.inEditingMode() {
				m.form = m.srForm
				m.form.State = huh.StateNormal
				m.border.GoTo("f5")
				log.Debug("go to f5")
				return nil
			} else {
				log.Debug("not in edit")
				return tea.Batch(
					m.notifierCmdBar.Notifier.ShowError(fmt.Errorf("create a cluster before adding a schema registry")),
					m.notifierCmdBar.Notifier.AutoHideCmd(notifierCmdbarTag),
				)
			}
		case "f6":
			if m.inEditingMode() {
				m.form.State = huh.StateNormal
				m.border.GoTo("f6")
				return nil
			} else {
				return tea.Batch(
					m.notifierCmdBar.Notifier.ShowError(fmt.Errorf("create a cluster before adding a Kafka Connect Cluster")),
					m.notifierCmdBar.Notifier.AutoHideCmd(notifierCmdbarTag),
				)
			}
		}
	case kadmin.ConnCheckStartedMsg:
		m.state = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case kadmin.ConnCheckSucceededMsg:
		m.state = none
		cmds = append(cmds, m.registerCluster)
	case sradmin.ConnCheckStartedMsg:
		m.state = loading
		cmds = append(cmds, msg.AwaitCompletion)
	case sradmin.ConnCheckSucceededMsg:
		m.state = none
		return m.registerCluster
	case config.ClusterRegisteredMsg:
		m.preEditName = &msg.Cluster.Name
		m.clusterToEdit = msg.Cluster
		m.state = none
		m.border.WithInActiveColor(styles.ColorGrey)
		if activeTab == cTab {
			m.cForm = m.createCForm()
			m.form = m.cForm
		} else if activeTab == srTab {
			m.srForm = m.createSrForm()
			m.form = m.srForm
		} else {
			m.kcModel.Update(msg)
		}
	}

	if activeTab == kcTab {
		cmd := m.kcModel.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	_, msg, cmd := m.notifierCmdBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if msg == nil {
		return tea.Batch(cmds...)
	}

	if activeTab == cTab || activeTab == srTab {
		form, cmd := m.form.Update(msg)
		cmds = append(cmds, cmd)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
	}

	if activeTab == cTab {
		// 检查SSL启用状态是否改变
		if !m.clusterValues.isSSLEnabled() && m.sslSelectionState == sslEnabledState {
			// SSL状态改变，重新创建表单以显示或隐藏TLS证书配置选项
			m.cForm = m.createCForm()
			m.form = m.cForm
			m.NextField(3) // 跳到认证方法选择
			m.sslSelectionState = sslDisabledState
		} else if m.clusterValues.isSSLEnabled() && m.sslSelectionState == sslDisabledState {
			// SSL状态改变，重新创建表单以显示或隐藏TLS证书配置选项
			m.cForm = m.createCForm()
			m.form = m.cForm
			m.NextField(3)
			m.sslSelectionState = sslEnabledState
		}

		if !m.clusterValues.HasSASLAuthMethodSelected() &&
			m.authSelectionState == saslSelected {
			// if SASL authentication mode was previously selected and switched back to none
			m.cForm = m.createCForm()
			m.form = m.cForm
			if m.sslSelectionState == sslEnabledState {
				m.NextField(8)
			} else {
				m.NextField(4)
			}
			m.authSelectionState = noneSelected
		} else if m.clusterValues.HasSASLAuthMethodSelected() &&
			(m.authSelectionState == nothingSelected || m.authSelectionState == noneSelected) {
			// SASL authentication mode selected and previously nothing or none auth mode was selected
			m.cForm = m.createCForm()
			m.form = m.cForm
			if m.sslSelectionState == sslEnabledState {
				m.NextField(8)
			} else {
				m.NextField(4)
			}
			m.authSelectionState = saslSelected
		}

		if m.form.State == huh.StateCompleted && m.state != loading {
			return m.processClusterSubmission()
		}
	}

	if activeTab == srTab {
		if m.form.State == huh.StateCompleted && m.state != loading {
			return m.processSrSubmission()
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) registerCluster() tea.Msg {
	details := m.getRegistrationDetails()
	return m.clusterRegisterer.RegisterCluster(details)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	if m.border.ActiveTab() == kcTab {
		return m.kcModel.Shortcuts()
	}
	return m.shortcuts
}

func (m *Model) Title() string {
	if m.title == "" {
		return "Clusters / Create"
	}
	return m.title
}

func (m *Model) processSrSubmission() tea.Cmd {
	m.state = loading
	details := m.getRegistrationDetails()

	cluster := config.ToCluster(details)
	return func() tea.Msg {
		return m.srConnChecker(cluster.SchemaRegistry)
	}
}

func (m *Model) processClusterSubmission() tea.Cmd {
	m.state = loading
	details := m.getRegistrationDetails()

	cluster := config.ToCluster(details)
	return func() tea.Msg {
		return m.kConnChecker(&cluster)
	}
}

func (m *Model) getRegistrationDetails() config.RegistrationDetails {
	var name string
	var newName *string
	if m.preEditName == nil { // When creating a cluster
		name = m.clusterValues.name
		newName = nil
	} else { // When updating a cluster.
		name = *m.preEditName
		if m.clusterValues.name != *m.preEditName {
			newName = &m.clusterValues.name
		}
	}

	var authMethod config.AuthMethod
	var securityProtocol config.SecurityProtocol
	if m.clusterValues.HasSASLAuthMethodSelected() {
		authMethod = config.SASLAuthMethod
		securityProtocol = m.clusterValues.securityProtocol
	} else {
		authMethod = config.NoneAuthMethod
	}

	details := config.RegistrationDetails{
		Name:                  name,
		NewName:               newName,
		Color:                 m.clusterValues.color,
		Host:                  m.clusterValues.host,
		AuthMethod:            authMethod,
		SecurityProtocol:      securityProtocol,
		SSLEnabled:            m.clusterValues.sslEnabled,
		TLSCertFile:           m.clusterValues.tlsCertFile,
		TLSKeyFile:            m.clusterValues.tlsKeyFile,
		TLSCAFile:             m.clusterValues.tlsCAFile,
		TLSInsecureSkipVerify: m.clusterValues.tlsInsecureSkipVerify,
		Username:              m.clusterValues.username,
		Password:              m.clusterValues.password,
	}
	if m.clusterValues.SrEnabled() {
		details.SchemaRegistry = &config.SchemaRegistryDetails{
			Url:      m.clusterValues.srUrl,
			Username: m.clusterValues.srUsername,
			Password: m.clusterValues.srPassword,
		}
	}

	details.KafkaConnectClusters = m.kcModel.clusterDetails()

	return details
}

func (f *clusterValues) HasSASLAuthMethodSelected() bool {
	return f.authMethod == config.SASLAuthMethod
}

func (f *clusterValues) isSSLEnabled() bool {
	return f.sslEnabled == true
}

func (f *clusterValues) SrEnabled() bool {
	return len(f.srUrl) > 0
}

func (m *Model) NextField(count int) {
	for i := 0; i < count; i++ {
		m.form.NextField()
	}
}

func (m *Model) createCForm() *huh.Form {
	name := huh.NewInput().
		Value(&m.clusterValues.name).
		Title("Name").
		Validate(func(v string) error {
			if v == "" {
				return errors.New("name cannot be empty")
			}
			if m.preEditName != nil {
				// When updating.
				if m.ktx.Config.FindClusterByName(v) != nil && v != *m.preEditName {
					return errors.New("cluster " + v + " already exists, name most be unique")
				}
			} else {
				// When creating a new cluster
				if m.ktx.Config.FindClusterByName(v) != nil {
					return errors.New("cluster " + v + " already exists, name most be unique")
				}
			}
			return nil
		})
	color := huh.NewSelect[string]().
		Value(&m.clusterValues.color).
		Title("Color").
		Options(
			huh.NewOption(styles.Env.Colors.Green.Render("green"), styles.ColorGreen),
			huh.NewOption(styles.Env.Colors.Blue.Render("blue"), styles.ColorBlue),
			huh.NewOption(styles.Env.Colors.Orange.Render("orange"), styles.ColorOrange),
			huh.NewOption(styles.Env.Colors.Purple.Render("purple"), styles.ColorPurple),
			huh.NewOption(styles.Env.Colors.Yellow.Render("yellow"), styles.ColorYellow),
			huh.NewOption(styles.Env.Colors.Red.Render("red"), styles.ColorRed),
		)
	host := huh.NewInput().
		Value(&m.clusterValues.host).
		Title("Host").
		Validate(func(v string) error {
			if v == "" {
				return errors.New("host cannot be empty")
			}
			return nil
		})

	sslEnabled := huh.NewSelect[bool]().
		Value(&m.clusterValues.sslEnabled).
		Title("SSL").
		Options(
			huh.NewOption("Disable SSL", false),
			huh.NewOption("Enable SSL", true),
		)

	var clusterFields []huh.Field
	clusterFields = append(clusterFields, name, color, host, sslEnabled)

	if m.clusterValues.isSSLEnabled() {
		tlsCertFile := huh.NewInput().
			Value(&m.clusterValues.tlsCertFile).
			Title("TLS Certificate File Path")

		tlsKeyFile := huh.NewInput().
			Value(&m.clusterValues.tlsKeyFile).
			Title("TLS Key File Path")

		tlsCAFile := huh.NewInput().
			Value(&m.clusterValues.tlsCAFile).
			Title("TLS CA Certificate File Path")

		tlsInsecureSkipVerify := huh.NewSelect[bool]().
			Value(&m.clusterValues.tlsInsecureSkipVerify).
			Title("Skip TLS Verification").
			Options(
				huh.NewOption("Verify TLS Certificate", false),
				huh.NewOption("Skip TLS Verification (Insecure)", true),
			)

		clusterFields = append(clusterFields, tlsCertFile, tlsKeyFile, tlsCAFile, tlsInsecureSkipVerify)
	}

	auth := huh.NewSelect[config.AuthMethod]().
		Value(&m.clusterValues.authMethod).
		Title("Authentication method").
		Options(
			huh.NewOption("NONE", config.NoneAuthMethod),
			huh.NewOption("SASL", config.SASLAuthMethod),
		)
	clusterFields = append(clusterFields, auth)

	if m.clusterValues.HasSASLAuthMethodSelected() {
		securityProtocol := huh.NewSelect[config.SecurityProtocol]().
			Value(&m.clusterValues.securityProtocol).
			Title("Security Protocol").
			Options(
				huh.NewOption("SASL_PLAINTEXT", config.SASLPlaintextSecurityProtocol),
			)
		username := huh.NewInput().
			Value(&m.clusterValues.username).
			Title("Username")
		pwd := huh.NewInput().
			Value(&m.clusterValues.password).
			EchoMode(huh.EchoModePassword).
			Title("Password")
		clusterFields = append(clusterFields, securityProtocol, username, pwd)
	}

	form := huh.NewForm(
		huh.NewGroup(clusterFields...).
			Title("Cluster").
			WithWidth(m.ktx.WindowWidth - 3),
	)
	form.QuitAfterSubmit = false
	form.Init()
	return form
}

func (m *Model) createSrForm() *huh.Form {
	var fields []huh.Field
	srUrl := huh.NewInput().
		Value(&m.clusterValues.srUrl).
		Title("Schema Registry URL")
	srUsername := huh.NewInput().
		Value(&m.clusterValues.srUsername).
		Title("Schema Registry Username")
	srPwd := huh.NewInput().
		Value(&m.clusterValues.srPassword).
		EchoMode(huh.EchoModePassword).
		Title("Schema Registry Password")
	fields = append(fields, srUrl, srUsername, srPwd)

	form := huh.NewForm(
		huh.NewGroup(fields...).
			Title("Schema Registry").
			WithWidth(m.ktx.WindowWidth - 3),
	)
	form.QuitAfterSubmit = false
	form.Init()

	return form
}

func (m *Model) createNotifierCmdBar() {
	m.notifierCmdBar = cmdbar.NewNotifierCmdBar(notifierCmdbarTag)
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckStartedMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Testing cluster connectivity")
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckSucceededMsg, m *notifier.Model) (bool, tea.Cmd) {
		return true, m.SpinWithLoadingMsg("Connection success creating cluster")
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg kadmin.ConnCheckErrMsg, nm *notifier.Model) (bool, tea.Cmd) {
		m.cForm = m.createCForm()
		m.form = m.cForm
		m.state = none
		nMsg := "Cluster not crated"
		if m.inEditingMode() {
			nMsg = "Cluster not updated"
		}
		return true, nm.ShowErrorMsg(nMsg, msg.Err)
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg config.ClusterRegisteredMsg, nm *notifier.Model) (bool, tea.Cmd) {
		if m.form == m.srForm {
			nm.ShowSuccessMsg("Schema registry registered! <ESC> to go back.")
		} else if m.form == m.cForm {
			if m.inEditingMode() {
				nm.ShowSuccessMsg("Cluster updated!")
			} else {
				nm.ShowSuccessMsg("Cluster registered! <ESC> to go back or <F5> to add a schema registry.")
			}
		} else {
			nm.ShowSuccessMsg("Cluster registered!")
		}
		return true, nm.AutoHideCmd(notifierCmdbarTag)
	})
	cmdbar.WithMsgHandler(m.notifierCmdBar, func(msg sradmin.ConnCheckErrMsg, nm *notifier.Model) (bool, tea.Cmd) {
		m.srForm = m.createSrForm()
		m.form = m.srForm
		m.state = none
		nm.ShowErrorMsg("unable to reach the schema registry", msg.Err)
		return true, nm.AutoHideCmd(notifierCmdbarTag)
	})
}

func (m *Model) inEditingMode() bool {
	return m.clusterToEdit != nil
}

func WithTitle(title string) Option {
	return func(m *Model) {
		m.title = title
	}
}

func initBorder(options ...border.Option) *border.Model {
	return border.New(
		append([]border.Option{
			border.WithTabs(
				border.Tab{Title: "Cluster ≪ F4 »", TabLabel: cTab},
				border.Tab{Title: "Schema Registry ≪ F5 »", TabLabel: srTab},
				border.Tab{Title: "Kafka Connect ≪ F6 »", TabLabel: kcTab},
			),
		}, options...)...)
}

func NewCreateClusterPage(
	NavBack ui.NavBack,
	kConnChecker kadmin.ConnChecker,
	srConnChecker sradmin.ConnChecker,
	registerer config.ClusterRegisterer,
	ktx *kontext.ProgramKtx,
	shortcuts []statusbar.Shortcut,
	options ...Option,
) *Model {
	formValues := &clusterValues{}
	model := Model{
		NavBack:       NavBack,
		clusterValues: formValues,
		kConnChecker:  kConnChecker,
		srConnChecker: srConnChecker,
		shortcuts:     shortcuts,
	}

	model.ktx = ktx

	model.border = initBorder(border.WithInactiveColor(styles.ColorDarkGrey))

	model.cForm = model.createCForm()
	model.srForm = model.createSrForm()
	model.form = model.cForm

	model.createNotifierCmdBar()

	model.kcModel = NewUpsertKcModel(NavBack, ktx, nil, []config.KafkaConnectConfig{}, kcadmin.CheckKafkaConnectClustersConn, model.notifierCmdBar, model.registerCluster)

	model.clusterRegisterer = registerer

	model.authSelectionState = nothingSelected
	model.sslSelectionState = nothingSelected
	model.state = none

	if model.clusterValues.HasSASLAuthMethodSelected() {
		model.authSelectionState = saslSelected
	}

	if model.clusterValues.isSSLEnabled() {
		model.sslSelectionState = sslEnabledState
	} else {
		model.sslSelectionState = sslDisabledState
	}

	for _, option := range options {
		option(&model)
	}

	return &model
}

func NewEditClusterPage(
	back ui.NavBack,
	kConnChecker kadmin.ConnChecker,
	srConnChecker sradmin.ConnChecker,
	registerer config.ClusterRegisterer,
	connectClusterDeleter config.ConnectClusterDeleter,
	ktx *kontext.ProgramKtx,
	cluster config.Cluster,
	options ...Option,
) *Model {
	formValues := &clusterValues{
		name:                  cluster.Name,
		color:                 cluster.Color,
		host:                  cluster.BootstrapServers[0],
		sslEnabled:            cluster.SSLEnabled,
		tlsCertFile:           cluster.TLSCertFile,
		tlsKeyFile:            cluster.TLSKeyFile,
		tlsCAFile:             cluster.TLSCAFile,
		tlsInsecureSkipVerify: cluster.TLSInsecureSkipVerify,
	}
	if cluster.SASLConfig != nil {
		formValues.securityProtocol = cluster.SASLConfig.SecurityProtocol
		formValues.username = cluster.SASLConfig.Username
		formValues.password = cluster.SASLConfig.Password
		formValues.authMethod = config.SASLAuthMethod
	}
	if cluster.SchemaRegistry != nil {
		formValues.srUrl = cluster.SchemaRegistry.Url
		formValues.srUsername = cluster.SchemaRegistry.Username
		formValues.srPassword = cluster.SchemaRegistry.Password
	}
	model := Model{
		NavBack:       back,
		clusterToEdit: &cluster,
		clusterValues: formValues,
		kConnChecker:  kConnChecker,
		srConnChecker: srConnChecker,
		shortcuts: []statusbar.Shortcut{
			{"Confirm", "enter"},
			{"Next Field", "tab"},
			{"Prev. Field", "s-tab"},
			{"Reset Form", "C-r"},
			{"Go Back", "esc"},
		},
	}
	if cluster.Name != "" {
		// copied to prevent model.preEditedName to follow the formValues.Name pointer
		preEditedName := cluster.Name
		model.preEditName = &preEditedName
	}
	model.ktx = ktx

	model.border = initBorder(border.WithInactiveColor(styles.ColorGrey))

	model.cForm = model.createCForm()
	model.srForm = model.createSrForm()
	model.form = model.cForm

	model.createNotifierCmdBar()

	model.kcModel = NewUpsertKcModel(
		back,
		ktx,
		func(name string) tea.Msg {
			return connectClusterDeleter.DeleteKafkaConnectCluster(cluster.Name, name)
		},
		cluster.KafkaConnectClusters,
		kcadmin.CheckKafkaConnectClustersConn,
		model.notifierCmdBar,
		model.registerCluster,
	)

	model.clusterRegisterer = registerer
	model.authSelectionState = nothingSelected
	model.sslSelectionState = nothingSelected
	model.state = none

	if model.clusterValues.HasSASLAuthMethodSelected() {
		model.authSelectionState = saslSelected
	}

	if model.clusterValues.isSSLEnabled() {
		model.sslSelectionState = sslEnabledState
	} else {
		model.sslSelectionState = sslDisabledState
	}

	for _, o := range options {
		o(&model)
	}

	return &model
}
