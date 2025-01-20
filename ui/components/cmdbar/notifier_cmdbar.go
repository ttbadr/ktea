package cmdbar

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"reflect"
)

// Notification triggers a specific notification and
// returns if the CmdBar is still active or not along with an optional
// tea.Cmd to execute.
type Notification[T any] func(T, *notifier.Model) (bool, tea.Cmd)

type NotifierCmdBar struct {
	active            bool
	Notifier          *notifier.Model
	msgByNotification map[reflect.Type]Notification[any]
}

func (n *NotifierCmdBar) IsFocussed() bool {
	return n.active && n.Notifier.HasPriority()
}

func (n *NotifierCmdBar) Shortcuts() []statusbar.Shortcut {
	return nil
}

func (n *NotifierCmdBar) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	return n.Notifier.View(ktx, renderer)
}

func (n *NotifierCmdBar) Update(msg tea.Msg) (bool, tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		cmd := n.Notifier.Update(msg)
		return n.active, nil, cmd
	case notifier.HideNotificationMsg:
		cmd := n.Notifier.Update(msg)
		return n.active, nil, cmd
	}

	msgType := reflect.TypeOf(msg)
	if notification, ok := n.msgByNotification[msgType]; ok {
		active, cmd := notification(msg, n.Notifier)
		n.active = active
		return n.active, nil, cmd
	}
	return n.active, msg, nil
}

func WithMsgHandler[T any](bar *NotifierCmdBar, notification Notification[T]) *NotifierCmdBar {
	msgType := reflect.TypeOf((*T)(nil)).Elem()
	bar.msgByNotification[msgType] = WrapNotification(notification)
	return bar
}

func WrapNotification[T any](n Notification[T]) Notification[any] {
	return func(msg any, m *notifier.Model) (bool, tea.Cmd) {
		typedMsg, ok := msg.(T)
		if !ok {
			return false, nil
		}
		return n(typedMsg, m)
	}
}

func NewNotifierCmdBar() *NotifierCmdBar {
	return &NotifierCmdBar{msgByNotification: make(map[reflect.Type]Notification[any]), Notifier: notifier.New()}
}
