package consumption_form_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/tests/keys"
	"ktea/ui"
	"ktea/ui/pages/nav"
	"testing"
)

func TestConsumeForm_Navigation(t *testing.T) {

	t.Run("esc goes back to topic list page", func(t *testing.T) {
		m := New(kadmin.Topic{
			Name:       "topic1",
			Replicas:   1,
			Isr:        1,
			Partitions: 10,
		})
		// make sure form has been initialized
		m.View(ui.TestKontext, ui.TestRenderer)

		cmd := m.Update(keys.Key(tea.KeyEsc))

		assert.IsType(t, nav.LoadTopicsPageMsg{}, cmd())
	})

	t.Run("renders all available partitions when there is height enough", func(t *testing.T) {
		m := New(kadmin.Topic{
			Name:       "topic1",
			Replicas:   1,
			Isr:        1,
			Partitions: 10,
		})
		// make sure form has been initialized
		render := m.View(&kontext.ProgramKtx{
			Config:          nil,
			WindowWidth:     100,
			WindowHeight:    100,
			AvailableHeight: 100,
		}, ui.TestRenderer)

		assert.Contains(t, render, `
   Partitions                                                                    
   Select none to consume from all available partitions                          
   > • 0                                                                         
     • 1                                                                         
     • 2                                                                         
     • 3                                                                         
     • 4                                                                         
     • 5                                                                         
     • 6                                                                         
     • 7                                                                         
     • 8                                                                         
     • 9`)
		assert.NotContains(t, render, "• 10")
	})

	t.Run("renders subset of partitions when there is not enough height", func(t *testing.T) {
		m := New(kadmin.Topic{
			Name:       "topic1",
			Replicas:   1,
			Isr:        1,
			Partitions: 100,
		})
		// make sure form has been initialized
		render := m.View(&kontext.ProgramKtx{
			Config:          nil,
			WindowWidth:     100,
			WindowHeight:    20,
			AvailableHeight: 20,
		}, ui.TestRenderer)

		assert.Contains(t, render, `
   Partitions                                                                    
   Select none to consume from all available partitions                          
   > • 0                                                                         
     • 1                                                                         
     • 2                                                                         
     • 3                                                                         
     • 4                                                                         
                                                                                 
   Limit`)
		assert.NotContains(t, render, "• 5")
	})

	t.Run("submitting form loads consumption page with consumption information", func(t *testing.T) {
		m := New(kadmin.Topic{
			Name:       "topic1",
			Partitions: 10,
			Replicas:   1,
			Isr:        1,
		})
		// make sure form has been initialized
		m.View(ui.TestKontext, ui.TestRenderer)

		// select start from most recent
		cmd := m.Update(keys.Key(tea.KeyDown))
		cmd = m.Update(keys.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		// select partition 3 and 5
		m.Update(keys.Key(tea.KeyDown))
		m.Update(keys.Key(tea.KeyDown))
		m.Update(keys.Key(tea.KeyDown))
		cmd = m.Update(keys.Key(tea.KeySpace))
		m.Update(keys.Key(tea.KeyDown))
		m.Update(keys.Key(tea.KeyDown))
		m.Update(keys.Key(tea.KeySpace))
		cmd = m.Update(keys.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// select limit 500
		m.Update(keys.Key(tea.KeyDown))
		cmd = m.Update(keys.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		cmd = m.Update(cmd())

		assert.Equal(t, nav.LoadConsumptionPageMsg{
			ReadDetails: kadmin.ReadDetails{
				Topic: kadmin.Topic{
					Name:       "topic1",
					Partitions: 10,
					Replicas:   1,
					Isr:        1,
				},
				Limit:      500,
				Partitions: []int{3, 5},
				StartPoint: kadmin.MostRecent,
			},
		}, cmd())
	})

	t.Run("selecting partitions is optional", func(t *testing.T) {
		m := New(kadmin.Topic{
			Name:       "topic1",
			Partitions: 10,
			Replicas:   1,
			Isr:        1,
		})
		// make sure form has been initialized
		m.View(ui.TestKontext, ui.TestRenderer)

		// select start from most recent
		cmd := m.Update(keys.Key(tea.KeyDown))
		cmd = m.Update(keys.Key(tea.KeyEnter))
		// next field
		m.Update(cmd())
		// select no partitions
		cmd = m.Update(keys.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// select limit 500
		m.Update(keys.Key(tea.KeyDown))
		cmd = m.Update(keys.Key(tea.KeyEnter))
		// next field
		cmd = m.Update(cmd())
		// next group
		cmd = m.Update(cmd())

		assert.Equal(t, nav.LoadConsumptionPageMsg{
			ReadDetails: kadmin.ReadDetails{
				Topic: kadmin.Topic{
					Name:       "topic1",
					Partitions: 10,
					Replicas:   1,
					Isr:        1,
				},
				Limit:      500,
				Partitions: []int{},
				StartPoint: kadmin.MostRecent,
			},
		}, cmd())
	})

}
