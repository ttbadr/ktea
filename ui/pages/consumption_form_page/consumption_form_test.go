package consumption_form_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/tests/keys"
	"ktea/ui/pages/navigation"
	"testing"
)

func TestConsumeForm_Navigation(t *testing.T) {

	t.Run("esc goes back to topic list page", func(t *testing.T) {
		m := New(kadmin.Topic{"topic1", 0, 1, 1})

		cmd := m.Update(keys.Key(tea.KeyEsc))

		assert.IsType(t, navigation.LoadTopicsPageMsg{}, cmd())
	})

}

func submit(cmd tea.Cmd, m *Model) tea.Cmd {
	// next field
	cmd = m.Update(cmd())
	// next group
	cmd = m.Update(cmd())
	return cmd
}
