package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin in development
	},
}

// WebSocket clients
type WebClient struct {
	conn *websocket.Conn
	send chan models.EventOutput
}

type WebHub struct {
	clients    map[*WebClient]bool
	register   chan *WebClient
	unregister chan *WebClient
	broadcast  chan models.EventOutput
	verbose    bool
}

func NewWebHub(verbose bool) *WebHub {
	return &WebHub{
		clients:    make(map[*WebClient]bool),
		register:   make(chan *WebClient),
		unregister: make(chan *WebClient),
		broadcast:  make(chan models.EventOutput),
		verbose:    verbose,
	}
}

func (h *WebHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if h.verbose {
				log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if h.verbose {
					log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))
				}
			}

		case event := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- event:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *WebHub) Broadcast(event models.EventOutput) {
	select {
	case h.broadcast <- event:
	default:
		// Non-blocking send
	}
}

func HandleWebSocket(hub *WebHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}

		client := &WebClient{
			conn: conn,
			send: make(chan models.EventOutput, 256),
		}

		hub.register <- client

		go func() {
			defer func() {
				hub.unregister <- client
				conn.Close()
			}()

			for {
				select {
				case event, ok := <-client.send:
					if !ok {
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}

					if err := conn.WriteJSON(event); err != nil {
						log.Printf("WebSocket write error: %v", err)
						return
					}
				}
			}
		}()

		// Keep connection alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}

func StartServer(port int, styleCSS, appJS []byte, indexHandler http.HandlerFunc) *WebHub {
	hub := NewWebHub(false)
	go hub.Run()

	// Serve embedded static files
	http.HandleFunc("/static/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(styleCSS)
	})
	http.HandleFunc("/static/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(appJS)
	})
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ws", HandleWebSocket(hub))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Web UI available at http://localhost%s\n", addr)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	return hub
}
