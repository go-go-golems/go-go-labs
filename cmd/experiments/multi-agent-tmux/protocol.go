package main

import (
	"encoding/json"
	"time"
)

// SocketMessage represents a message sent over Unix socket
type SocketMessage struct {
	Type      string    `json:"type"`       // "agent_update", "status_update", "init", "shutdown"
	AgentID   string    `json:"agent_id"`   // Agent identifier
	AgentName string    `json:"agent_name"` // Human-readable agent name
	AgentRole string    `json:"agent_role"` // Agent role description
	Content   string    `json:"content"`    // Message content
	MsgType   string    `json:"msg_type"`   // Message type: "status", "progress", "result", "error"
	Timestamp time.Time `json:"timestamp"`  // When the message was created
}

// Marshal converts SocketMessage to JSON bytes
func (sm *SocketMessage) Marshal() ([]byte, error) {
	return json.Marshal(sm)
}

// UnmarshalSocketMessage converts JSON bytes to SocketMessage
func UnmarshalSocketMessage(data []byte) (*SocketMessage, error) {
	var msg SocketMessage
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

// NewAgentUpdateMessage creates a new agent update message
func NewAgentUpdateMessage(agentID, agentName, agentRole, content, msgType string) *SocketMessage {
	return &SocketMessage{
		Type:      "agent_update",
		AgentID:   agentID,
		AgentName: agentName,
		AgentRole: agentRole,
		Content:   content,
		MsgType:   msgType,
		Timestamp: time.Now(),
	}
}

// NewStatusUpdateMessage creates a new status update message
func NewStatusUpdateMessage(content string) *SocketMessage {
	return &SocketMessage{
		Type:      "status_update",
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewInitMessage creates an initialization message for an agent
func NewInitMessage(agentID, agentName, agentRole string) *SocketMessage {
	return &SocketMessage{
		Type:      "init",
		AgentID:   agentID,
		AgentName: agentName,
		AgentRole: agentRole,
		Content:   "Initializing agent display",
		Timestamp: time.Now(),
	}
}

// NewShutdownMessage creates a shutdown message
func NewShutdownMessage() *SocketMessage {
	return &SocketMessage{
		Type:      "shutdown",
		Content:   "Shutting down display",
		Timestamp: time.Now(),
	}
}
