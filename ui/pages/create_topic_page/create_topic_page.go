package create_topic_page

import (
	"errors"
	"fmt"
	bsp "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"ktea/ui/components/notifier"
	"ktea/ui/components/statusbar"
	"ktea/ui/pages"
	"regexp"
	"strconv"
	"strings"
)

type formState int

const (
	initial       formState = 0
	created       formState = 0
	configEntered formState = 1
	loading       formState = 2
)

type Model struct {
	shortcuts    []statusbar.Shortcut
	form         *huh.Form
	notifier     *notifier.Model
	topicCreator kadmin.TopicCreator
	formValues   topicFormValues
	formState    formState
}

type config struct {
	key   string
	value string
}

type topicFormValues struct {
	name          string
	numPartitions string
	config        string
	configs       []config
	cleanupPolicy string
}

func (m *Model) View(ktx *kontext.ProgramKtx, renderer *ui.Renderer) string {
	var views []string
	notifierView := m.notifier.View(ktx, renderer)
	views = append(views, notifierView, m.form.View())

	if len(m.formValues.configs) > 0 {
		views = append(views, renderer.Render("Custom Topic configurations:\n\n"))
		for _, c := range m.formValues.configs {
			views = append(views, renderer.Render(fmt.Sprintf("%s: %s\n", c.key, c.value)))
		}
	}

	return ui.JoinVerticalSkipEmptyViews(views...)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case kadmin.TopicCreationStartedMsg:
		return msg.AwaitCompletion
	case bsp.TickMsg:
		cmd := m.notifier.Update(msg)
		return cmd
	case tea.KeyMsg:
		if msg.String() == "esc" && m.formState != loading {
			return ui.PublishMsg(pages.LoadTopicsPageMsg{})
		} else if msg.String() == "ctrl+r" {
			m.formValues.name = ""
			m.formValues.cleanupPolicy = ""
			m.formValues.config = ""
			m.formValues.numPartitions = ""
			m.formValues.configs = []config{}
			m.initForm(initial)
			return propagateMsgToForm(m, msg)
		} else {
			return propagateMsgToForm(m, msg)
		}
	//case kadmin.TopicCreationErrorMsg:
	//	log.Debug("TopicCreationError", msg.Err)
	//	m.aTopicCreated = false
	//	return nil
	case kadmin.TopicCreatedMsg:
		m.notifier.ShowSuccessMsg("Topic created!")
		m.formValues.name = ""
		m.formValues.cleanupPolicy = ""
		m.formValues.config = ""
		m.formValues.numPartitions = ""
		m.formValues.configs = []config{}
		m.initForm(initial)
		return nil
	default:
		return propagateMsgToForm(m, msg)
	}
}

func propagateMsgToForm(m *Model, msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	form, c := m.form.Update(msg)
	cmd = c
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted && m.formState != loading {
		if m.formValues.config == "" {
			m.formState = loading
			return tea.Batch(
				m.notifier.SpinWithRocketMsg("Creating topic"),
				func() tea.Msg {
					numPartitions, _ := strconv.Atoi(m.formValues.numPartitions)
					return m.topicCreator.CreateTopic(
						kadmin.TopicCreationDetails{
							Name:          m.formValues.name,
							NumPartitions: numPartitions,
							Properties: map[string]string{
								"cleanup.policy": m.formValues.cleanupPolicy,
							},
						})
				})
		} else {
			m.formState = configEntered
			split := strings.Split(m.formValues.config, "=")
			m.formValues.configs = append(m.formValues.configs, config{split[0], split[1]})
			m.formValues.config = ""
			m.initForm(0)
			return cmd
		}
	}
	return cmd
}

func (m *Model) Shortcuts() []statusbar.Shortcut {
	return m.shortcuts
}

func (m *Model) Title() string {
	return "Topics / Create"
}

func (m *Model) initForm(fs formState) {
	topicNameInput := huh.NewInput().
		Title("Topic name").
		Value(&m.formValues.name).
		Validate(func(str string) error {
			if str == "" {
				return errors.New("Topic Name cannot be empty.")
			}
			return nil
		})

	validateInput := huh.NewInput().
		Title("Number of Partitions").
		Value(&m.formValues.numPartitions).
		Validate(func(str string) error {
			if str == "" {
				return errors.New("Number of Partitions cannot be empty.")
			}
			if n, e := strconv.Atoi(str); e != nil {
				return errors.New(fmt.Sprintf("'%s' is not a valid numeric partition count value", str))
			} else if n <= 0 {
				return errors.New("Value must be greater than zero")
			}
			return nil
		})
	cleanupPolicySelect := huh.NewSelect[string]().
		Title("Cleanup Policy").
		Value(&m.formValues.cleanupPolicy).
		Options(
			huh.NewOption("Retention (delete)", "delete"),
			huh.NewOption("Compaction (compact)", "compact"),
			huh.NewOption("Retention + Compaction", "delete-compact"))

	configInput := huh.NewInput().
		Description("Enter custom topic configurations in the format config=value. Leave blank to create the topic.").
		Title("Config").
		Key("configKey").
		Suggestions([]string{
			"cleanup.policy",
			"compression.type",
			"delete.retention.ms",
			"file.delete.delay.ms",
			"flush.messages",
			"flush.ms",
			"follower.replication.throttled.replicas",
			"index.interval.bytes",
			"leader.replication.throttled.replicas",
			"local.retention.bytes",
			"local.retention.ms",
			"max.compaction.lag.ms",
			"max.message.bytes",
			"message.format.version",
			"message.timestamp.difference.max.ms",
			"message.timestamp.type",
			"min.cleanable.dirty.ratio",
			"min.compaction.lag.ms",
			"min.insync.replicas",
			"preallocate",
			"remote.storage.enable",
			"retention.bytes",
			"retention.ms",
			"segment.bytes",
			"segment.index.bytes",
			"segment.jitter.ms",
			"segment.ms",
			"unclean.leader.election.enable",
			"message.downconversion.enable",
		}).
		Validate(func(str string) error {
			if str == "" {
				return nil
			}

			r, _ := regexp.Compile(`^[\w.]+=\w+$`)
			if r.MatchString(str) {
				return nil
			}

			return errors.New("please enter configurations in the format \"config=value\"")
		}).
		Value(&m.formValues.config)

	form := huh.NewForm(huh.NewGroup(topicNameInput, validateInput, cleanupPolicySelect, configInput))
	form.QuitAfterSubmit = false
	if m.formState == configEntered {
		form.NextField()
		form.NextField()
		form.NextField()
	}
	form.Init()
	m.formState = fs
	m.form = form
}

func New(tc kadmin.TopicCreator) *Model {
	var t = Model{}
	t.topicCreator = tc
	t.shortcuts = []statusbar.Shortcut{
		{"Confirm", "enter"},
		{"Next Field", "tab"},
		{"Prev. Field", "s-tab"},
		{"Reset Form", "C-r"},
		{"Go Back", "esc"},
	}
	t.initForm(initial)
	t.notifier = notifier.New()
	return &t
}
