package tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"strings"
)

type Label string

type Tab struct {
	Title string
	Label
}

type Model struct {
	tabs []Tab
	// zero indexed
	activeTab int
}

func (m *Model) View(ctx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if len(m.tabs) == 0 {
		return ""
	}
	tabsToRender := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		var tab string
		if i == m.activeTab {
			tab = styles.Tab.ActiveTab.Render(t.Title)
		} else {
			tab = styles.Tab.Tab.Render(t.Title)
		}
		tabsToRender = append(tabsToRender, tab)
	}
	renderedTabs := lipgloss.JoinHorizontal(lipgloss.Top, tabsToRender...)
	tabLine := strings.Builder{}
	leftOverSpace := ctx.WindowWidth - lipgloss.Width(renderedTabs)
	for i := 0; i < leftOverSpace; i++ {
		tabLine.WriteString("─")
	}
	s := renderedTabs + tabLine.String()
	return renderer.Render(s)
}

func (m *Model) Update(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlLeft, tea.KeyCtrlH:
			m.Prev()
		case tea.KeyCtrlRight, tea.KeyCtrlL:
			m.Next()
		}
	}
}

func (m *Model) Next() {
	if m.activeTab < m.numberOfTabs()-1 {
		m.activeTab++
	}
}

func (m *Model) Prev() {
	if m.activeTab > 0 {
		m.activeTab--
	}
}

func (m *Model) numberOfTabs() int {
	return len(m.tabs)
}

func (m *Model) GoToTab(label Label) {
	for i, t := range m.tabs {
		if t.Label == label {
			m.activeTab = i
		}
	}
}

func (m *Model) ActiveTab() Tab {
	if m.tabs == nil {
		return Tab{}
	}
	return m.tabs[m.activeTab]
}

func New(tabs ...Tab) Model {
	return Model{
		tabs: tabs,
	}
}
