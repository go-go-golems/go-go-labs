package model

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/protocol"
)

type Model struct {
	client   *client.MeshtasticClient
	messages []protocol.Message
	input    string
	width    int
	height   int
	ready    bool
	err      error
}

type State int

const (
	StateLoading State = iota
	StateReady
	StateError
)

func New(client *client.MeshtasticClient) Model {
	return Model{
		client:   client,
		messages: make([]protocol.Message, 0),
		ready:    false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	return "Meshtastic TUI - Press 'q' to quit"
}

func (m *Model) AddMessage(message protocol.Message) {
	m.messages = append(m.messages, message)
}

func (m *Model) GetMessages() []protocol.Message {
	return m.messages
}

func (m *Model) SetError(err error) {
	m.err = err
}

func (m *Model) GetError() error {
	return m.err
}
