package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"os"
)

type WidgetConfig struct {
	Type         string
	Label        string
	Placeholder  string   `yaml:",omitempty"`
	SpinnerStyle string   `yaml:",omitempty"`
	Options      []string `yaml:",omitempty"`
	Progress     int      `yaml:",omitempty"`
	TabStop      bool
	// ... other fields depending on the widgets' properties
}

type AppConfig struct {
	Application struct {
		Name string
	}
	Layout struct {
		Type string
	}
	Widgets []WidgetConfig
}

// FocusMsg is dispatched to change focus state of widgets.
type FocusMsg bool

type WidgetModel tea.Model

type AppModel struct {
	widgets      []WidgetModel
	focusedIndex int
}

var _ tea.Model = AppModel{} // Ensure that AppModel implements the tea.Model interface

func NewAppModel(widgets []WidgetModel) AppModel {
	return AppModel{
		widgets:      widgets,
		focusedIndex: 0, // Start with the first widget focused
	}
}

func (m AppModel) Init() tea.Cmd {
	// We could send commands that should run when the program starts, like initializing sub-models.
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Remove focus from the currently focused widget
			_, cmd := m.widgets[m.focusedIndex].Update(FocusMsg(false))
			cmds = append(cmds, cmd)

			// Change the focus to the next widget
			m.focusedIndex = (m.focusedIndex + 1) % len(m.widgets)

			// Apply focus to the new widget
			_, cmd = m.widgets[m.focusedIndex].Update(FocusMsg(true))
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...) // Batch all commands together
}

func (m AppModel) View() string {
	s := ""
	for i, w := range m.widgets {
		view := w.View()
		if i == m.focusedIndex {
			// If the widget is focused, we could render a border or some other indication of focus.
			// This is a simplified example; you'd likely want a more sophisticated way of rendering a focus indicator.
			view = fmt.Sprintf("[ %s ]", view)
		}
		s += view + "\n"
	}
	return s
}

func main() {
	// Read the YAML file
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	// Parse the YAML file
	var config AppConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	runApp(config)
}

type TextInput struct {
	model textinput.Model
}

func (m TextInput) View() string {
	return m.model.View()
}

func (m TextInput) Init() tea.Cmd {
	return nil
}

func (m TextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd

}

type Spinner struct {
	model spinner.Model
}

func (s Spinner) Init() tea.Cmd {
	return nil
}

func (s Spinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	s.model, cmd = s.model.Update(msg)
	return s, cmd
}

func (s Spinner) View() string {
	return s.model.View()
}

type Progress struct {
	model progress.Model
}

func (p Progress) Init() tea.Cmd {
	return p.model.Init()
}

func (p Progress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := p.model.Update(msg)
	p.model = model.(progress.Model)
	return p, cmd
}

func (p Progress) View() string {
	return p.model.View()
}

// For bubbles/timer
type Timer struct {
	model timer.Model
}

func (t Timer) Init() tea.Cmd {
	return t.model.Init()
}

func (t Timer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return t, cmd
}

func (t Timer) View() string {
	return t.model.View()
}

// For bubbles/viewport
type Viewport struct {
	model viewport.Model
}

func (v Viewport) Init() tea.Cmd {
	return v.model.Init()
}

func (v Viewport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	v.model, cmd = v.model.Update(msg)
	return v, cmd
}

func (v Viewport) View() string {
	return v.model.View()
}

// For bubbles/table
type Table struct {
	model table.Model
}

func (t Table) Init() tea.Cmd {
	return nil // Modify if the table.Model has an Init function
}

func (t Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return t, cmd
}

func (t Table) View() string {
	return t.model.View()
}

// For bubbles/list
type List struct {
	model list.Model
}

func (l List) Init() tea.Cmd {
	return nil
}

func (l List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	l.model, cmd = l.model.Update(msg)
	return l, cmd
}

func (l List) View() string {
	return l.model.View()
}

// For bubbles/filepicker
type FilePicker struct {
	model filepicker.Model
}

func (f FilePicker) Init() tea.Cmd {
	return f.model.Init()
}

func (f FilePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	f.model, cmd = f.model.Update(msg)
	return f, cmd
}

func (f FilePicker) View() string {
	return f.model.View()
}

// For bubbles/textarea
type TextArea struct {
	model textarea.Model
}

func (t TextArea) Init() tea.Cmd {
	return nil
}

func (t TextArea) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return t, cmd
}

func (t TextArea) View() string {
	return t.model.View()
}

// For bubbles/stopwatch
type Stopwatch struct {
	model stopwatch.Model
}

func (s Stopwatch) Init() tea.Cmd {
	return s.model.Init()
}

func (s Stopwatch) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	s.model, cmd = s.model.Update(msg)
	return s, cmd
}

func (s Stopwatch) View() string {
	return s.model.View()
}

// For bubbles/paginator
type Paginator struct {
	model paginator.Model
}

func (p Paginator) Init() tea.Cmd {
	return nil
}

func (p Paginator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	p.model, cmd = p.model.Update(msg)
	return p, cmd
}

func (p Paginator) View() string {
	return p.model.View()
}

// For bubbles/help (assuming it follows the pattern of other models)
type Help struct {
	model help.Model // This is theoretical; you'll need to adjust based on the actual 'help' component structure
}

func (h Help) Init() tea.Cmd {
	return nil
}

func (h Help) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	h.model, cmd = h.model.Update(msg)
	return h, cmd
}

func (h Help) View() string {
	// needs to be done through keymap
	//return h.model.View()
	return ""
}

var _ tea.Model = Spinner{}

func createModelFromConfig(widgetConfig WidgetConfig) (WidgetModel, error) {
	switch widgetConfig.Type {
	case "textinput":
		model := TextInput{textinput.New()}
		return model, nil
	case "spinner":
		model := Spinner{spinner.New()}
		return model, nil
	// ... handle other widget types
	default:
		return nil, errors.Errorf("unknown widget type: %s", widgetConfig.Type)
	}
}

func runApp(config AppConfig) {
	// Create models from configuration
	var widgetModels []WidgetModel
	for _, widgetConfig := range config.Widgets {
		model, err := createModelFromConfig(widgetConfig)
		if err != nil {
			panic(err) // Handle error appropriately
		}
		widgetModels = append(widgetModels, model)
	}

	// Initialize the base model with the widget models
	m := NewAppModel(widgetModels)

	// Start the Bubbletea program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("There was an issue starting the program: %v", err)
		os.Exit(1)
	}
}
