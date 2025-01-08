package cmdbar

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type DeleteMsgFunc[T any] func(T) string

type DeleteFunc[T any] func(T) tea.Cmd

type DeleteCmdBarModel[T any] struct {
	active        bool
	deleteConfirm *huh.Confirm
	msg           string
	deleteValue   T
	deleteMsgFunc DeleteMsgFunc[T]
	deleteFunc    DeleteFunc[T]
}

func (s *DeleteCmdBarModel[any]) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	s.deleteConfirm.Title(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Render(lipgloss.NewStyle().MarginRight(2).Render("🗑️  " + s.deleteMsgFunc(s.deleteValue)))).
		WithButtonAlignment(lipgloss.Left).
		Focus()

	s.deleteConfirm.WithTheme(huh.ThemeCharm())

	return renderer.RenderWithStyle(s.deleteConfirm.View(), styles.CmdBar)
}

func (s *DeleteCmdBarModel[any]) IsFocussed() bool {
	return false
}

func (s *DeleteCmdBarModel[any]) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Select Cancel", "c"},
		{"Select Delete", "d"},
		{"Quit", "esc"},
	}
}

func (s *DeleteCmdBarModel[any]) Update(msg tea.Msg) (bool, tea.Msg, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			s.active = true
			return s.active, nil, nil
		case "esc":
			s.active = false
			return s.active, nil, nil
		case "enter":
			if s.deleteConfirm.GetValue().(bool) {
				return s.active, nil, s.deleteFunc(s.deleteValue)
			}
		}
	}
	confirm, cmd := s.deleteConfirm.Update(msg)
	if cmd != nil {
		// if msg has been handled do not propagate it
		msg = nil
	}
	if c, ok := confirm.(*huh.Confirm); ok {
		s.deleteConfirm = c
	}

	return s.active, msg, nil
}

func (s *DeleteCmdBarModel[any]) DeleteMsg(msg string) {
	s.msg = msg
}

func (s *DeleteCmdBarModel[T]) Delete(d T) {
	s.deleteValue = d
}

func newDeleteConfirm() *huh.Confirm {
	return huh.NewConfirm().
		Inline(true).
		Affirmative("Delete!").
		Negative("Cancel.").
		WithKeyMap(&huh.KeyMap{
			Confirm: huh.ConfirmKeyMap{
				Submit: key.NewBinding(key.WithKeys("enter")),
				Toggle: key.NewBinding(key.WithKeys("h", "l", "right", "left")),
				Accept: key.NewBinding(key.WithKeys("d")),
				Reject: key.NewBinding(key.WithKeys("c")),
			},
		}).(*huh.Confirm)
}

func NewDeleteCmdBar[T any](deleteMsgFunc DeleteMsgFunc[T], deleteFunc DeleteFunc[T]) *DeleteCmdBarModel[T] {
	return &DeleteCmdBarModel[T]{
		deleteFunc:    deleteFunc,
		deleteMsgFunc: deleteMsgFunc,
		deleteConfirm: newDeleteConfirm(),
	}
}
