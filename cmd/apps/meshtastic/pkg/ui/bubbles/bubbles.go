package bubbles

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/keys"
)

type Bubbles struct {
	TextInput textinput.Model
	Help      help.Model
	Keys      *keys.KeyMap
}

func New() Bubbles {
	ti := textinput.New()
	ti.Placeholder = "Type your message..."
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 50

	return Bubbles{
		TextInput: ti,
		Help:      help.New(),
		Keys:      keys.DefaultKeyMap(),
	}
}

func (b *Bubbles) UpdateTextInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	b.TextInput, cmd = b.TextInput.Update(msg)
	return cmd
}

func (b *Bubbles) UpdateHelp(msg tea.Msg) {
	b.Help, _ = b.Help.Update(msg)
}

func (b *Bubbles) GetTextInputValue() string {
	return b.TextInput.Value()
}

func (b *Bubbles) ClearTextInput() {
	b.TextInput.SetValue("")
}

func (b *Bubbles) FocusTextInput() {
	b.TextInput.Focus()
}

func (b *Bubbles) BlurTextInput() {
	b.TextInput.Blur()
}

func (b *Bubbles) SetTextInputWidth(width int) {
	b.TextInput.Width = width
}

func (b *Bubbles) RenderTextInput() string {
	return b.TextInput.View()
}

func (b *Bubbles) RenderHelp() string {
	return b.Help.View(b.Keys)
}
