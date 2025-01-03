package styles

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"ktea/kontext"
)

var Env EnvStyle
var Tab TabStyle
var Statusbar StatusBarStyle
var Table TableStyle
var CmdBar lipgloss.Style
var Form = lipgloss.NewStyle().
	Padding(1)
var Notifier NotifierStyle
var clusterColors map[string]string

const ColorRed = "#FF0000"
const ColorGreen = "#00FF00"
const ColorBlue = "#0000FF"
const ColorOrange = "#FFA500"
const ColorPurple = "#800080"
const ColorYellow = "#FFFF00"
const ColorWhite = "#FFFFFF"
const ColorBlack = "#000000"
const ColorPink = "#FF5F87"

var Bold = lipgloss.NewStyle().
	Bold(true)

type NotifierStyle struct {
	Spinner lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
}

type TableStyle struct {
	Focus  lipgloss.Style
	Blur   lipgloss.Style
	Styles table.Styles
}

type StatusBarStyle struct {
	style       lipgloss.Style
	Indicator   lipgloss.Style
	Text        lipgloss.Style
	BindingName lipgloss.Style
	Shortcuts   lipgloss.Style
	Key         lipgloss.Style
	Spacer      lipgloss.Style
	cluster     lipgloss.Style
}

func FG(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(color)
}

func (s *StatusBarStyle) Cluster(color string) lipgloss.Style {
	return s.cluster.Background(lipgloss.Color(color))
}

type TabStyle struct {
	Tab       lipgloss.Style
	ActiveTab lipgloss.Style
}

type EnvStyle struct {
	Colors struct {
		Red    lipgloss.Style
		Green  lipgloss.Style
		Blue   lipgloss.Style
		Orange lipgloss.Style
		Yellow lipgloss.Style
		Purple lipgloss.Style
	}
}

func (s *StatusBarStyle) Render(c *kontext.ProgramKtx, strs ...string) string {
	return s.style.Render(strs...)
}

func init() {
	{
		clusterColors = map[string]string{
			ColorRed:    ColorWhite,
			ColorGreen:  ColorBlack,
			ColorBlue:   ColorWhite,
			ColorOrange: ColorBlack,
			ColorPurple: ColorWhite,
			ColorYellow: ColorBlack,
		}
	}

	{
		tabStyle := TabStyle{}
		activeTabBorder := lipgloss.Border{
			Top:         "─",
			Bottom:      " ",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┘",
			BottomRight: "└",
		}
		tabBorder := lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┴",
			BottomRight: "┴",
		}
		tabStyle.Tab = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#AAAAAA")).
			Border(tabBorder)
		tabStyle.ActiveTab = lipgloss.NewStyle().
			Padding(0, 1).
			Border(activeTabBorder).
			Foreground(lipgloss.Color(ColorPink)).
			Bold(true)
		Tab = tabStyle
	}

	{
		statusBarStyle := StatusBarStyle{}
		statusBarStyle.style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

		statusBarStyle.Indicator = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			Foreground(lipgloss.Color(ColorBlack)).
			Background(lipgloss.Color(ColorPink)).
			Bold(true).
			Padding(0, 2)

		statusBarStyle.cluster = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(ColorRed)).
			Bold(true).
			Padding(0, 4)

		statusBarStyle.Text = lipgloss.NewStyle().
			MarginTop(1).
			PaddingLeft(0).
			PaddingRight(5)

		statusBarStyle.BindingName = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorYellow))

		statusBarStyle.Key = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite))

		statusBarStyle.Spacer = lipgloss.NewStyle().
			Inherit(statusBarStyle.style).
			PaddingLeft(1)

		statusBarStyle.Shortcuts = lipgloss.NewStyle().
			PaddingLeft(2)

		Statusbar = statusBarStyle
	}

	{
		CmdBar = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(0)
	}

	{
		Notifier.Spinner = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(0).
			MarginLeft(2).
			Height(1)
		Notifier.Success = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(0).
			MarginLeft(2).
			Height(1).
			Foreground(lipgloss.Color(ColorGreen)).
			Bold(true)
		Notifier.Error = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(0).
			MarginLeft(2).
			Height(1).
			Foreground(lipgloss.Color(ColorRed)).
			Bold(true)

	}

	{
		envStyle := EnvStyle{}
		envStyle.Colors.Red = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(ColorRed)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Green = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color(ColorGreen)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Blue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(ColorBlue)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Orange = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(ColorOrange)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Purple = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(ColorPurple)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)
		envStyle.Colors.Yellow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color(ColorYellow)).
			PaddingLeft(1).
			PaddingRight(1).
			Width(10).
			Bold(true)

		Env = envStyle
	}

	{
		blur := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0).
			Margin(0).
			BorderForeground(lipgloss.Color("235"))
		focus := lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("245")).
			Padding(0).
			Margin(0).
			Inherit(blur)
		styles := table.DefaultStyles()
		styles.Header = styles.Header.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		styles.Selected = styles.Selected.
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color(ColorPink)).
			Bold(true)
		Table = TableStyle{
			Focus:  focus,
			Blur:   blur,
			Styles: styles,
		}
	}
}
