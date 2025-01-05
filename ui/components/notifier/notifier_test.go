package notifier

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNotifier(t *testing.T) {
	t.Run("Update TickMsg", func(t *testing.T) {

		notifier := New()

		t.Run("when spinning continue ticking", func(t *testing.T) {

			notifier.SpinWithLoadingMsg("Loading")

			cmd := notifier.Update(spinner.TickMsg{
				Time: time.Now(),
				ID:   1,
			})

			assert.IsType(t, spinner.TickMsg{}, cmd())
		})

		t.Run("when not spinning stop ticking", func(t *testing.T) {

			notifier.Idle()

			cmd := notifier.Update(spinner.TickMsg{
				Time: time.Now(),
				ID:   1,
			})

			assert.Nil(t, cmd)
		})
	})
}
