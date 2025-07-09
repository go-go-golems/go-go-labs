package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/protocol"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(1)

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			Margin(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("36")).
			Padding(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

func RenderTitle(title string) string {
	return titleStyle.Render(title)
}

func RenderMessage(msg protocol.Message) string {
	content := fmt.Sprintf("From: %s\nTo: %s\nContent: %s", msg.From, msg.To, msg.Content)
	return messageStyle.Render(content)
}

func RenderMessages(messages []protocol.Message, width, height int) string {
	if len(messages) == 0 {
		return "No messages"
	}

	var rendered []string
	for _, msg := range messages {
		rendered = append(rendered, RenderMessage(msg))
	}

	return strings.Join(rendered, "\n")
}

func RenderInput(input string, width int) string {
	return inputStyle.Width(width - 4).Render(input)
}

func RenderError(err error) string {
	return errorStyle.Render(fmt.Sprintf("Error: %v", err))
}

func RenderLoadingScreen() string {
	return titleStyle.Render("Loading Meshtastic TUI...")
}

func RenderMainScreen(messages []protocol.Message, input string, width, height int) string {
	title := RenderTitle("Meshtastic TUI")

	messagesHeight := height - 8 // Reserve space for title, input, and padding
	messagesView := RenderMessages(messages, width, messagesHeight)

	inputView := RenderInput(input, width)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		messagesView,
		inputView,
	)
}
