package mockbot

import (
	"strings"
	"time"

	"github.com/go-go-golems/go-go-labs/pkg/sse"
)

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents an incoming chat request
type ChatRequest struct {
	Messages []ChatMessage `json:"messages"`
}

// MockBot implements a simple mock chat bot that reverses messages
type MockBot struct {
	eventBus *sse.EventBus
}

// NewMockBot creates a new mock bot instance
func NewMockBot(eventBus *sse.EventBus) *MockBot {
	return &MockBot{
		eventBus: eventBus,
	}
}

// HandleMessage processes a chat message and streams the response
func (b *MockBot) HandleMessage(client *sse.Client, req ChatRequest) {
	// Get the last message and reverse it
	lastMessage := req.Messages[len(req.Messages)-1]
	response := reverseString(lastMessage.Content)

	// Stream the response
	b.streamResponse(client, response)
}

// streamResponse streams a response to the client
func (b *MockBot) streamResponse(client *sse.Client, response string) {
	// Send "thinking" event
	client.Events <- sse.Event{Type: "thinking", Content: ""}
	time.Sleep(1 * time.Second)

	// Split response into words and stream them
	words := strings.Fields(response)
	for i, word := range words {
		select {
		case <-client.Done:
			return
		default:
			client.Events <- sse.Event{Type: "token", Content: word}
			if i < len(words)-1 {
				client.Events <- sse.Event{Type: "token", Content: " "}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	client.Events <- sse.Event{Type: "done", Content: ""}
}

// reverseString reverses a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
