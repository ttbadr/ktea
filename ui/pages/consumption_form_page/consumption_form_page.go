package consumption_form_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/ui"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages/nav"
	"strconv"
)

type Model struct {
	form          *huh.Form
	topic         *kadmin.Topic
	formValues    *formValues
	windowResized bool
}

type formValues struct {
	startPoint kadmin.StartPoint
	limit      int
	partitions []int
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	if m.form == nil {
		m.form = m.newForm(m.topic.Partitions, ktx)
	}

	if m.windowResized {
		m.windowResized = false
		m.form = m.newForm(m.topic.Partitions, ktx)
	}

	return renderer.RenderWithStyle(m.form.View(), styles.Form)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m.form == nil {
		return nil
	}

	switch msg.(type) {
	case tea.WindowSizeMsg:
		m.windowResized = true
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return ui.PublishMsg(nav.LoadTopicsPageMsg{})
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		return ui.PublishMsg(nav.LoadConsumptionPageMsg{
			ReadDetails: kadmin.ReadDetails{
				Topic:      m.topic,
				Partitions: m.formValues.partitions,
				StartPoint: m.formValues.startPoint,
				Limit:      m.formValues.limit,
			},
		})
	}
	return cmd
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Next Field", "tab"},
		{"Prev. Field", "s-tab"},
		{"Select Partition", "space"},
		{"Go Back", "esc"},
	}
}

func (m *Model) Title() string {
	return "Consumption details"
}

func (m *Model) newForm(partitions int, ktx *kontext.ProgramKtx) *huh.Form {
	var partOptions []huh.Option[int]
	for i := 0; i < partitions; i++ {
		partOptions = append(partOptions, huh.NewOption[int](strconv.Itoa(i), i))
	}
	optionsHeight := 13 // 12 fixed height of form minus partitions field + padding and margins
	if len(partOptions) < 13 {
		optionsHeight = len(partOptions) + 2 // 2 for field title + padding
	} else {
		optionsHeight = ktx.AvailableHeight - optionsHeight
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[kadmin.StartPoint]().
				Value(&m.formValues.startPoint).
				Title("Start form").
				Options(
					huh.NewOption("Beginning", kadmin.Beginning),
					huh.NewOption("Most Recent", kadmin.MostRecent)),
			huh.NewMultiSelect[int]().
				Value(&m.formValues.partitions).
				Height(optionsHeight).
				Title("Partitions").
				Description("Select none to consume from all available partitions").
				Options(partOptions...),
			huh.NewSelect[int]().
				Value(&m.formValues.limit).
				Title("Limit").
				Options(
					huh.NewOption("50", 50),
					huh.NewOption("500", 500),
					huh.NewOption("5000", 5000)),
		),
	)
	form.Init()
	return form
}

func NewWithDetails(details *kadmin.ReadDetails) *Model {
	return &Model{topic: details.Topic, formValues: &formValues{
		startPoint: details.StartPoint,
		limit:      details.Limit,
		partitions: details.Partitions,
	}}
}

func New(topic *kadmin.Topic) *Model {
	return &Model{topic: topic, formValues: &formValues{}}
}
