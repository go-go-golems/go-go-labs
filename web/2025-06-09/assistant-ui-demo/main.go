package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type TodoItem struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

type DropdownOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type UIWidget struct {
	Type  string      `json:"type"`
	ID    string      `json:"id"`
	Title string      `json:"title"`
	Data  interface{} `json:"data"`
}

func main() {
	var logLevel string
	var port int

	rootCmd := &cobra.Command{
		Use:   "assistant-ui-demo",
		Short: "Demo application with streaming chat and generative UI",
		Run: func(cmd *cobra.Command, args []string) {
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				log.Fatal(err)
			}
			zerolog.SetGlobalLevel(level)
			logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

			logger.Info().Msg("Starting assistant UI demo server")

			r := mux.NewRouter()

			// Serve static files
			r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

			// API routes
			r.HandleFunc("/ws", handleWebSocket).Methods("GET")
			r.HandleFunc("/", handleIndex).Methods("GET")

			addr := fmt.Sprintf(":%d", port)
			srv := &http.Server{
				Addr:    addr,
				Handler: r,
			}

			go func() {
				logger.Info().Str("addr", srv.Addr).Msg("Server starting")
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal().Err(err).Msg("Server failed to start")
				}
			}()

			// Wait for interrupt signal
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			logger.Info().Msg("Shutting down server...")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				logger.Fatal().Err(err).Msg("Server forced to shutdown")
			}

			logger.Info().Msg("Server exited")
		},
	}

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().IntVar(&port, "port", 8080, "Port to listen on")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Assistant UI Demo</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="/static/style.css" rel="stylesheet">
</head>
<body>
    <div id="root"></div>
    <script src="/static/main.js"></script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upgrade connection")
		return
	}
	defer conn.Close()

	logger.Info().Msg("WebSocket connection established")

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to read message")
			break
		}

		logger.Info().Str("type", msg.Type).Msg("Received message")

		switch msg.Type {
		case "chat":
			handleChatMessage(conn, msg.Content)
		default:
			logger.Warn().Str("type", msg.Type).Msg("Unknown message type")
		}
	}
}

func handleChatMessage(conn *websocket.Conn, content interface{}) {
	contentBytes, _ := json.Marshal(content)
	var chatMsg ChatMessage
	json.Unmarshal(contentBytes, &chatMsg)

	// Echo user message
	conn.WriteJSON(Message{
		Type:    "message",
		Content: chatMsg,
	})

	// Simulate thinking time
	time.Sleep(500 * time.Millisecond)

	// Generate response with potential UI widgets
	response := generateResponse(chatMsg.Content)

	// Stream response
	for _, chunk := range strings.Split(response.Content, " ") {
		conn.WriteJSON(Message{
			Type: "chunk",
			Content: map[string]string{
				"id":    response.ID,
				"chunk": chunk + " ",
			},
		})
		time.Sleep(50 * time.Millisecond)
	}

	// Send complete message
	conn.WriteJSON(Message{
		Type:    "message",
		Content: response,
	})

	// Maybe generate UI widgets
	if shouldGenerateWidget(chatMsg.Content) {
		widget := generateWidget(chatMsg.Content)
		conn.WriteJSON(Message{
			Type: "widget",
			Content: map[string]interface{}{
				"messageId": response.ID,
				"widget":    widget,
			},
		})
	}
}

func generateResponse(userMessage string) ChatMessage {
	responses := []string{
		"That's an interesting question! Let me think about that.",
		"I can help you with that. Here's what I think:",
		"Great point! From my perspective:",
		"Let me break that down for you:",
		"I understand what you're asking. Here's my take:",
	}

	response := responses[rand.Intn(len(responses))]

	// Add more context based on user message
	if strings.Contains(strings.ToLower(userMessage), "todo") {
		response += " It sounds like you need help organizing tasks."
	} else if strings.Contains(strings.ToLower(userMessage), "help") {
		response += " I'm here to assist you with various tasks and questions."
	} else {
		response += " That's quite a thoughtful question that deserves a detailed answer."
	}

	return ChatMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
	}
}

func shouldGenerateWidget(userMessage string) bool {
	keywords := []string{"todo", "list", "tasks", "organize", "dropdown", "options", "choose"}
	msg := strings.ToLower(userMessage)

	for _, keyword := range keywords {
		if strings.Contains(msg, keyword) {
			return true
		}
	}

	return rand.Float32() < 0.3 // 30% chance for demo purposes
}

func generateWidget(userMessage string) UIWidget {
	msg := strings.ToLower(userMessage)

	if strings.Contains(msg, "todo") || strings.Contains(msg, "task") {
		return UIWidget{
			Type:  "todo",
			ID:    fmt.Sprintf("widget_%d", time.Now().UnixNano()),
			Title: "Your Tasks",
			Data: []TodoItem{
				{ID: "1", Text: "Complete project documentation", Completed: false},
				{ID: "2", Text: "Review code changes", Completed: true},
				{ID: "3", Text: "Schedule team meeting", Completed: false},
			},
		}
	}

	// Default to dropdown
	return UIWidget{
		Type:  "dropdown",
		ID:    fmt.Sprintf("widget_%d", time.Now().UnixNano()),
		Title: "Choose an option",
		Data: []DropdownOption{
			{Value: "option1", Label: "First Option"},
			{Value: "option2", Label: "Second Option"},
			{Value: "option3", Label: "Third Option"},
		},
	}
}
