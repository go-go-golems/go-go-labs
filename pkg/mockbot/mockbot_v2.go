package mockbot

import (
	"strings"
	"sync"
	"time"

	"github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/go-go-labs/pkg/sse"
	"github.com/google/uuid"
)

// BotV2 implements a mock chat bot with conversation management
type BotV2 struct {
	eventBus *sse.EventBus
	// Map of client IDs to conversation managers
	conversations sync.Map
}

// NewBotV2 creates a new mock bot instance with conversation management
func NewBotV2(eventBus *sse.EventBus) *BotV2 {
	return &BotV2{
		eventBus: eventBus,
	}
}

// getOrCreateManager gets an existing conversation manager or creates a new one
func (b *BotV2) getOrCreateManager(clientID string) conversation.Manager {
	if manager, ok := b.conversations.Load(clientID); ok {
		return manager.(conversation.Manager)
	}

	// Create new manager with system prompt
	manager := conversation.NewManager(
		conversation.WithManagerConversationID(uuid.New()),
	)

	// Add system prompt
	manager.AppendMessages(conversation.NewChatMessage(
		conversation.RoleSystem,
		"I am a mock chat bot that reverses messages for testing purposes.",
	))

	b.conversations.Store(clientID, manager)
	return manager
}

// HandleMessage processes a chat message and streams the response
func (b *BotV2) HandleMessage(client *sse.Client, req ChatRequest) {
	manager := b.getOrCreateManager(client.ID)

	// Convert the last message to a conversation message
	lastMessage := req.Messages[len(req.Messages)-1]
	userMsg := conversation.NewChatMessage(
		conversation.RoleUser,
		lastMessage.Content,
	)

	// Add user message to conversation
	manager.AppendMessages(userMsg)

	// Generate and stream response
	response := reverseString(lastMessage.Content)
	b.streamResponse(client, response, manager)
}

// streamResponse streams a response to the client and updates the conversation
func (b *BotV2) streamResponse(client *sse.Client, response string, manager conversation.Manager) {
	// Send "thinking" event
	client.Events <- sse.Event{Type: "thinking", Content: ""}
	time.Sleep(1 * time.Second)

	// Prepare assistant message
	assistantMsg := conversation.NewChatMessage(
		conversation.RoleAssistant,
		response,
	)

	// Add message to conversation before streaming
	manager.AppendMessages(assistantMsg)

	// Stream the response word by word
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

// GetConversation returns the conversation for a client
func (b *BotV2) GetConversation(clientID string) (conversation.Conversation, bool) {
	if manager, ok := b.conversations.Load(clientID); ok {
		return manager.(conversation.Manager).GetConversation(), true
	}
	return nil, false
}

// SaveConversation saves the conversation for a client to a file
func (b *BotV2) SaveConversation(clientID string, filename string) error {
	if manager, ok := b.conversations.Load(clientID); ok {
		return manager.(conversation.Manager).SaveToFile(filename)
	}
	return nil
}

// LoadConversation loads a conversation from a file for a client
func (b *BotV2) LoadConversation(clientID string, filename string) error {
	manager := conversation.NewManager(
		conversation.WithManagerConversationID(uuid.New()),
	)

	if err := manager.Tree.LoadFromFile(filename); err != nil {
		return err
	}

	b.conversations.Store(clientID, manager)
	return nil
}
