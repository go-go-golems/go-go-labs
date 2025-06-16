package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
)

// SocketServer manages Unix socket connections for TUI clients
type SocketServer struct {
	socketDir   string
	listeners   map[string]net.Listener
	connections map[string][]net.Conn
	mu          sync.RWMutex
}

// NewSocketServer creates a new socket server
func NewSocketServer() (*SocketServer, error) {
	// Create a temporary directory for sockets
	socketDir, err := os.MkdirTemp("", "multi-agent-sockets-")
	if err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	return &SocketServer{
		socketDir:   socketDir,
		listeners:   make(map[string]net.Listener),
		connections: make(map[string][]net.Conn),
	}, nil
}

// CreateAgentSocket creates a Unix socket for an agent
func (s *SocketServer) CreateAgentSocket(agentID string) (string, error) {
	socketPath := filepath.Join(s.socketDir, fmt.Sprintf("%s.sock", agentID))

	// Remove any existing socket file
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return "", fmt.Errorf("failed to create socket for agent %s: %w", agentID, err)
	}

	s.mu.Lock()
	s.listeners[agentID] = listener
	s.connections[agentID] = make([]net.Conn, 0)
	s.mu.Unlock()

	// Start accepting connections for this agent
	go s.acceptConnections(agentID, listener)

	log.Debug().
		Str("agent_id", agentID).
		Str("socket_path", socketPath).
		Msg("Created agent socket")

	return socketPath, nil
}

// CreateStatusSocket creates a Unix socket for orchestrator status
func (s *SocketServer) CreateStatusSocket() (string, error) {
	socketPath := filepath.Join(s.socketDir, "orchestrator.sock")

	// Remove any existing socket file
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return "", fmt.Errorf("failed to create status socket: %w", err)
	}

	s.mu.Lock()
	s.listeners["orchestrator"] = listener
	s.connections["orchestrator"] = make([]net.Conn, 0)
	s.mu.Unlock()

	// Start accepting connections for status
	go s.acceptConnections("orchestrator", listener)

	log.Debug().
		Str("socket_path", socketPath).
		Msg("Created status socket")

	return socketPath, nil
}

// acceptConnections accepts incoming connections for a socket
func (s *SocketServer) acceptConnections(id string, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Debug().
				Str("id", id).
				Err(err).
				Msg("Socket accept error (likely shutdown)")
			return
		}

		s.mu.Lock()
		s.connections[id] = append(s.connections[id], conn)
		s.mu.Unlock()

		log.Debug().
			Str("id", id).
			Msg("New TUI client connected")

		// Handle connection cleanup when client disconnects
		go func(conn net.Conn, id string) {
			// Wait for connection to close
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				// Just consume any input from client (typically none)
			}

			// Remove connection when it closes
			s.mu.Lock()
			connections := s.connections[id]
			for i, c := range connections {
				if c == conn {
					s.connections[id] = append(connections[:i], connections[i+1:]...)
					break
				}
			}
			s.mu.Unlock()

			conn.Close()
			log.Debug().
				Str("id", id).
				Msg("TUI client disconnected")
		}(conn, id)
	}
}

// SendToAgent sends a message to all TUI clients connected to an agent
func (s *SocketServer) SendToAgent(agentID string, msg *SocketMessage) error {
	s.mu.RLock()
	connections := s.connections[agentID]
	s.mu.RUnlock()

	if len(connections) == 0 {
		// No clients connected, silently ignore
		return nil
	}

	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Add newline for line-based protocol
	data = append(data, '\n')

	s.mu.RLock()
	defer s.mu.RUnlock()

	var failed []int
	for i, conn := range s.connections[agentID] {
		if _, err := conn.Write(data); err != nil {
			log.Warn().
				Str("agent_id", agentID).
				Err(err).
				Msg("Failed to send message to TUI client")
			failed = append(failed, i)
		}
	}

	// Remove failed connections
	if len(failed) > 0 {
		newConnections := make([]net.Conn, 0, len(s.connections[agentID])-len(failed))
		for i, conn := range s.connections[agentID] {
			shouldRemove := false
			for _, failedIdx := range failed {
				if i == failedIdx {
					shouldRemove = true
					conn.Close()
					break
				}
			}
			if !shouldRemove {
				newConnections = append(newConnections, conn)
			}
		}
		s.connections[agentID] = newConnections
	}

	return nil
}

// SendToStatus sends a message to all TUI clients connected to orchestrator status
func (s *SocketServer) SendToStatus(msg *SocketMessage) error {
	return s.SendToAgent("orchestrator", msg)
}

// InitializeAgent sends initialization message to agent TUI clients
func (s *SocketServer) InitializeAgent(agentID, agentName, agentRole string) error {
	msg := NewInitMessage(agentID, agentName, agentRole)
	return s.SendToAgent(agentID, msg)
}

// Shutdown closes all sockets and cleans up
func (s *SocketServer) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Send shutdown messages to all connections
	shutdownMsg := NewShutdownMessage()
	for id := range s.connections {
		for _, conn := range s.connections[id] {
			if data, err := shutdownMsg.Marshal(); err == nil {
				data = append(data, '\n')
				conn.Write(data)
			}
			conn.Close()
		}
	}

	// Close all listeners
	for id, listener := range s.listeners {
		listener.Close()
		log.Debug().Str("id", id).Msg("Closed socket listener")
	}

	// Clean up socket directory
	if s.socketDir != "" {
		os.RemoveAll(s.socketDir)
		log.Debug().Str("socketDir", s.socketDir).Msg("Cleaned up socket directory")
	}

	return nil
}

// GetSocketDir returns the socket directory path
func (s *SocketServer) GetSocketDir() string {
	return s.socketDir
}
