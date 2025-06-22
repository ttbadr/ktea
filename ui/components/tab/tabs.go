package tab

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/tabs"
	"regexp"
	"strconv"
	"strings"
)

type Model struct {
	elements []string
	// zero indexed
	activeTab int
}

func (m *Model) View(ctx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if len(m.elements) == 0 {
		return ""
	}
	tabsToRender := make([]string, len(m.elements))
	if len(m.elements) == 1 {
		tabsToRender = append(tabsToRender, styles.Tab.ActiveTab.Render(m.elements[0]))
	} else {
		for i, e := range m.elements {
			tabName := fmt.Sprintf("%s (Meta-%d)", e, i+1)
			if i == m.activeTab {
				tabsToRender = append(tabsToRender, styles.Tab.ActiveTab.Render(tabName))
			} else {
				tabsToRender = append(tabsToRender, styles.Tab.Tab.Render(tabName))
			}
		}
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
		keyPattern, _ := regexp.Compile("alt\\+(\\d)")
		if keyPattern.MatchString(msg.String()) {
			subMatch := keyPattern.FindAllStringSubmatch(msg.String(), -1)[0]
			tabNr, _ := strconv.Atoi(subMatch[1])
			m.GoToTab(tabNr - 1)
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
	return len(m.elements)
}

func (m *Model) GoToTab(tab interface{}) {
	switch tab := tab.(type) {
	case tabs.TabName:
		if int(tab) <= m.numberOfTabs()-1 {
			m.activeTab = int(tab)
		}
	case int:
		if tab <= m.numberOfTabs()-1 {
			m.activeTab = tab
		}
	}
}

func (m *Model) ActiveTab() int {
	return m.activeTab
}

func New(tabs ...string) Model {
	return Model{
		elements: tabs,
	}
}
