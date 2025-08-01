package notifier

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"strings"
	"sync/atomic"
	"time"
)

type state int

const (
	idle     state = 0
	Err      state = 1
	success  state = 2
	Spinning state = 3
)

type Model struct {
	spinner    spinner.Model
	successMsg string
	msg        string
	State      state
	autoHide   atomic.Bool
}

type HideNotificationMsg struct {
	Tag string
}

type NotificationHiddenMsg struct{}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.State == Spinning {
		return renderer.RenderWithStyle(
			lg.JoinHorizontal(lg.Top, m.spinner.View(), m.msg),
			styles.Notifier.Spinner,
		)
	} else if m.State == success {
		return renderer.RenderWithStyle(
			wordwrap.String(m.msg, ktx.WindowWidth),
			styles.Notifier.Success,
		)
	} else if m.State == Err {
		return renderer.RenderWithStyle(
			wordwrap.String(m.msg, ktx.WindowWidth),
			styles.Notifier.Error,
		)
	}
	return renderer.Render("")
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.State != Spinning {
			return nil
		}
		s, cmd := m.spinner.Update(msg)
		m.spinner = s
		return cmd
	case HideNotificationMsg:
		m.Idle()
		return func() tea.Msg {
			return NotificationHiddenMsg{}
		}
	}
	return nil
}

func (m *Model) SpinWithLoadingMsg(msg string) tea.Cmd {
	m.autoHide.Store(false)
	m.State = Spinning
	m.msg = "⏳ " + msg
	return m.spinner.Tick
}

func (m *Model) SpinWithRocketMsg(msg string) tea.Cmd {
	m.autoHide.Store(false)
	m.State = Spinning
	m.msg = "🚀 " + msg
	return m.spinner.Tick
}

func (m *Model) ShowErrorMsg(msg string, error error) tea.Cmd {
	m.autoHide.Store(false)
	m.State = Err
	s := ": "
	if msg == "" {
		s = ""
	}
	m.msg = "🚨 " + styles.FG(styles.ColorRed).Render(msg+s) +
		styles.FG(styles.ColorWhite).Render(strings.TrimSuffix(error.Error(), "\n"))
	return nil
}

func (m *Model) ShowError(error error) tea.Cmd {
	m.autoHide.Store(false)
	m.State = Err
	msg := error.Error()
	split := strings.SplitN(msg, ":", 2)
	if len(split) > 1 {
		m.msg = "🚨 " + styles.FG(styles.ColorRed).Render(split[0]) + ": " +
			styles.FG(styles.ColorWhite).Render(strings.TrimSuffix(split[1], "\n"))
	} else {
		m.msg = "🚨 " + styles.FG(styles.ColorRed).Render(msg)
	}
	return nil
}

func (m *Model) ShowSuccessMsg(msg string) tea.Cmd {
	m.autoHide.Store(false)
	m.State = success
	m.msg = "🎉 " + msg
	return nil
}

func (m *Model) Idle() {
	m.autoHide.Store(false)
	m.State = idle
	m.msg = ""
}

func (m *Model) AutoHideCmd(tag string) tea.Cmd {
	m.autoHide.Store(true)
	return func() tea.Msg {
		time.Sleep(5 * time.Second)
		if m.autoHide.Load() {
			return HideNotificationMsg{Tag: tag}
		}
		return nil
	}
}

func (m *Model) HasPriority() bool {
	return m.State == Spinning
}

func (m *Model) IsIdle() bool {
	return m.State == idle
}

func New() *Model {
	l := Model{}
	l.State = idle
	l.spinner = spinner.New()
	l.spinner.Spinner = spinner.Dot
	return &l
}
