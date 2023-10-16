package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ws", handleConnections)

	go handleBroadcast()

	go func() {
		counter := 0
		for {
			counter++
			broadcast <- strconv.Itoa(counter)
			time.Sleep(5 * time.Second)
		}
	}()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "cmd/ws-htmx/index.html")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func(ws *websocket.Conn) {
		_ = ws.Close()
	}(ws)

	clients[ws] = true

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(clients, ws)
			break
		}
	}
}

func handleBroadcast() {
	for {
		msg := <-broadcast
		html := fmt.Sprintf("<div>%s</div>", msg)
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(html))
			if err != nil {
				err := client.Close()
				if err != nil {
					return
				}
				delete(clients, client)
			}
		}
	}
}
