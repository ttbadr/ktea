package error_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
)

type Model struct {
	err error
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return renderer.Render(lipgloss.NewStyle().
		Width(ktx.WindowWidth).
		Height(ktx.AvailableHeight).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("Error: " + m.err.Error() +
			lipgloss.NewStyle().
				AlignHorizontal(lipgloss.Center).
				AlignVertical(lipgloss.Center).
				Render("\nPress "+lipgloss.NewStyle().Bold(true).Render("F5")+" to "+
					lipgloss.NewStyle().Foreground(lipgloss.Color(styles.ColorYellow)).Render("retry"))))
}

func (m *Model) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return nil
}

func (m *Model) Title() string {
	return ""
}

func New(err error) pages.Page {
	return &Model{err: err}
}
