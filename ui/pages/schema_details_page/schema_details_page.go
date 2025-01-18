package schema_details_page

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/sradmin"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/chips"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"slices"
	"sort"
	"strconv"
)

type Model struct {
	cmdbar        *CmdBar
	schemas       []sradmin.Schema
	vp            *viewport.Model
	subject       sradmin.Subject
	versionChips  *chips.Model
	schemaLister  sradmin.VersionLister
	activeVersion int
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string

	views = append(views, m.cmdbar.View(ktx, renderer))

	if m.schemas != nil {
		if m.vp == nil {
			var versions []string
			for _, schema := range m.schemas {
				versions = append(versions, strconv.Itoa(schema.Version))
			}
			m.versionChips = chips.New("Versions", versions...)
			m.versionChips.ActivateByLabel(strconv.Itoa(m.activeVersion))
			vp := viewport.New(ktx.WindowWidth-3, ktx.AvailableHeight-4)
			m.vp = &vp
		}
		if m.vp != nil {
			m.vp.Height = ktx.AvailableHeight - 4
			m.vp.Width = ktx.WindowWidth - 3
			views = append(views, lipgloss.NewStyle().
				PaddingTop(1).
				PaddingLeft(1).
				Render(m.versionChips.View(ktx, renderer)))

			m.vp.SetContent(ui.PrettyPrintJson(m.activeSchema()))
			views = append(views, renderer.RenderWithStyle(m.vp.View(), styles.TextViewPort))
		}
	}

	return ui.JoinVerticalSkipEmptyViews(lipgloss.Top, views...)
}

func (m *Model) activeSchema() string {
	var schema string
	for _, s := range m.schemas {
		if m.activeVersion == s.Version {
			schema = s.Schema
		}
	}
	return schema
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	msg, cmd := m.cmdbar.Update(msg)
	cmds = append(cmds, cmd)

	if m.versionChips != nil {
		cmd := m.versionChips.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return ui.PublishMsg(nav.LoadSubjectsPageMsg{})
		case "enter":
			version, _ := strconv.Atoi(m.versionChips.SelectedLabel())
			m.activeVersion = version
		}
	case sradmin.SchemasListed:
		m.schemas = msg.Schemas
		sort.Slice(m.schemas, func(i int, j int) bool {
			return m.schemas[i].Version < m.schemas[j].Version
		})
		m.activeVersion = m.latestSchema().Version
	case sradmin.SchemaListingStarted:
		return msg.AwaitCompletion
	}

	if m.vp != nil {
		vp, cmd := m.vp.Update(msg)
		m.vp = &vp
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{
			Name:       "Prev Version",
			Keybinding: "h/←",
		},
		{
			Name:       "Next Version",
			Keybinding: "l/→",
		},
		{
			Name:       "Select Version",
			Keybinding: "enter",
		},
		{
			Name:       "Delete Version",
			Keybinding: "F2",
		},
		{
			Name:       "Copy Version",
			Keybinding: "h/←",
		},
		{
			Name:       "Go Back",
			Keybinding: "esc",
		},
	}
}

func (m *Model) Title() string {
	schema := m.latestSchema()
	if schema != nil {
		// wait until schemas have been loaded
		return "Subjects / " + m.subject.Name + " / Versions / " + strconv.Itoa(schema.Version)
	}
	return ""
}

func (m *Model) latestSchema() *sradmin.Schema {
	if m.schemas == nil {
		return nil
	}
	var latest = slices.MaxFunc(m.schemas, func(a sradmin.Schema, b sradmin.Schema) int {
		if a.Version >= b.Version {
			return a.Version
		}
		return b.Version
	})
	return &latest
}

func New(
	schemaLister sradmin.VersionLister,
	subject sradmin.Subject,
) (*Model, tea.Cmd) {
	model := &Model{
		cmdbar:       NewCmdBar(),
		subject:      subject,
		schemaLister: schemaLister,
	}
	return model, func() tea.Msg {
		return schemaLister.ListVersions(subject.Name, subject.Versions)
	}
}
