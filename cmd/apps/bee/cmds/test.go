package cmds

import "fmt"

import (
	"github.com/googollee/go-socket.io"
	"log"
)

func main() {
	// Connect to the Socket.IO server
	client, err := socketio.NewClient("http://localhost:5000")
	if err != nil {
		log.Fatal(err)
	}

	// Handle connection event
	client.OnEvent("connect", func() {
		fmt.Println("Connected to server")
	})

	// Handle custom events
	client.OnEvent("message", func(msg string) {
		fmt.Println("Received message:", msg)
	})

	// Handle disconnection
	client.OnEvent("disconnect", func() {
		fmt.Println("Disconnected from server")
	})

	// Keep the connection alive
	select {}
}
