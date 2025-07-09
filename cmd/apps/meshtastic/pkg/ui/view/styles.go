package view

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors defines the color scheme for the TUI
var Colors = struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	Info      lipgloss.Color
	Subtle    lipgloss.Color
	Muted     lipgloss.Color
	Border    lipgloss.Color
	Selected  lipgloss.Color
}{
	Primary:   lipgloss.Color("#7C3AED"),
	Secondary: lipgloss.Color("#6B7280"),
	Accent:    lipgloss.Color("#3B82F6"),
	Success:   lipgloss.Color("#10B981"),
	Warning:   lipgloss.Color("#F59E0B"),
	Error:     lipgloss.Color("#EF4444"),
	Info:      lipgloss.Color("#06B6D4"),
	Subtle:    lipgloss.Color("#F3F4F6"),
	Muted:     lipgloss.Color("#9CA3AF"),
	Border:    lipgloss.Color("#E5E7EB"),
	Selected:  lipgloss.Color("#DDD6FE"),
}

// Styles contains all the styling for the TUI
type Styles struct {
	// Layout
	App       lipgloss.Style
	Header    lipgloss.Style
	Footer    lipgloss.Style
	StatusBar lipgloss.Style
	TabBar    lipgloss.Style
	Content   lipgloss.Style
	Sidebar   lipgloss.Style

	// Tabs
	Tab         lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	// Lists
	List             lipgloss.Style
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListTitle        lipgloss.Style

	// Messages
	Message         lipgloss.Style
	MessageSent     lipgloss.Style
	MessageReceived lipgloss.Style
	MessageHeader   lipgloss.Style
	MessageBody     lipgloss.Style
	MessageTime     lipgloss.Style

	// Nodes
	Node        lipgloss.Style
	NodeOnline  lipgloss.Style
	NodeOffline lipgloss.Style
	NodeId      lipgloss.Style
	NodeName    lipgloss.Style
	NodeStatus  lipgloss.Style

	// Status
	StatusItem  lipgloss.Style
	StatusKey   lipgloss.Style
	StatusValue lipgloss.Style
	StatusText  lipgloss.Style

	// Compose
	ComposeBox   lipgloss.Style
	ComposeInput lipgloss.Style
	ComposeTitle lipgloss.Style

	// Common
	Border    lipgloss.Style
	Focused   lipgloss.Style
	Unfocused lipgloss.Style
	Selected  lipgloss.Style
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Text      lipgloss.Style
	Muted     lipgloss.Style
	Error     lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Info      lipgloss.Style

	// Help
	Help     lipgloss.Style
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
}

// DefaultStyles returns the default styling for the TUI
func DefaultStyles() *Styles {
	s := &Styles{}

	// Layout
	s.App = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border)

	s.Header = lipgloss.NewStyle().
		Background(Colors.Primary).
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Padding(0, 1)

	s.Footer = lipgloss.NewStyle().
		Background(Colors.Secondary).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1)

	s.StatusBar = lipgloss.NewStyle().
		Background(Colors.Subtle).
		Foreground(Colors.Secondary).
		Padding(0, 1)

	s.TabBar = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(Colors.Border).
		Padding(0, 1)

	s.Content = lipgloss.NewStyle().
		Padding(1, 2)

	s.Sidebar = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(Colors.Border).
		Padding(1, 2)

	// Tabs
	s.Tab = lipgloss.NewStyle().
		Padding(0, 2)

	s.TabActive = s.Tab.Copy().
		Background(Colors.Primary).
		Foreground(lipgloss.Color("255")).
		Bold(true)

	s.TabInactive = s.Tab.Copy().
		Background(Colors.Subtle).
		Foreground(Colors.Secondary)

	// Lists
	s.List = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 2)

	s.ListItem = lipgloss.NewStyle().
		Padding(0, 1)

	s.ListItemSelected = s.ListItem.Copy().
		Background(Colors.Selected).
		Foreground(Colors.Primary).
		Bold(true)

	s.ListTitle = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		Padding(0, 0, 1, 0)

	// Messages
	s.Message = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 2).
		Margin(0, 0, 1, 0)

	s.MessageSent = s.Message.Copy().
		BorderForeground(Colors.Primary).
		Align(lipgloss.Right)

	s.MessageReceived = s.Message.Copy().
		BorderForeground(Colors.Info).
		Align(lipgloss.Left)

	s.MessageHeader = lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		Bold(true)

	s.MessageBody = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Padding(1, 0, 0, 0)

	s.MessageTime = lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Italic(true)

	// Nodes
	s.Node = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 2).
		Margin(0, 0, 1, 0)

	s.NodeOnline = s.Node.Copy().
		BorderForeground(Colors.Success)

	s.NodeOffline = s.Node.Copy().
		BorderForeground(Colors.Error)

	s.NodeId = lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		Bold(true)

	s.NodeName = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)

	s.NodeStatus = lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Italic(true)

	// Status
	s.StatusItem = lipgloss.NewStyle().
		Padding(0, 0, 1, 0)

	s.StatusKey = lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		Bold(true)

	s.StatusValue = lipgloss.NewStyle().
		Foreground(Colors.Primary)

	s.StatusText = lipgloss.NewStyle().
		Foreground(Colors.Secondary)

	// Compose
	s.ComposeBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Padding(1, 2)

	s.ComposeInput = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 2)

	s.ComposeTitle = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		Padding(0, 0, 1, 0)

	// Common
	s.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border)

	s.Focused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary)

	s.Unfocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border)

	s.Selected = lipgloss.NewStyle().
		Background(Colors.Selected).
		Foreground(Colors.Primary).
		Bold(true)

	s.Title = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		Padding(0, 0, 1, 0)

	s.Subtitle = lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		Bold(true)

	s.Text = lipgloss.NewStyle().
		Foreground(Colors.Primary)

	s.Muted = lipgloss.NewStyle().
		Foreground(Colors.Muted)

	s.Error = lipgloss.NewStyle().
		Foreground(Colors.Error).
		Bold(true)

	s.Success = lipgloss.NewStyle().
		Foreground(Colors.Success).
		Bold(true)

	s.Warning = lipgloss.NewStyle().
		Foreground(Colors.Warning).
		Bold(true)

	s.Info = lipgloss.NewStyle().
		Foreground(Colors.Info).
		Bold(true)

	// Help
	s.Help = lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		Padding(1, 2)

	s.HelpKey = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)

	s.HelpDesc = lipgloss.NewStyle().
		Foreground(Colors.Secondary)

	return s
}
