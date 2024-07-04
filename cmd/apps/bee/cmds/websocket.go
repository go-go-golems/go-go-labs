package cmds

import (
	"context"
	"fmt"
	_ "github.com/googollee/go-socket.io"
	socketio "github.com/googollee/go-socket.io"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"time"
)

type Bee struct {
	client *socketio.Client
	apiKey string
}

func NewBee(apiKey string) *Bee {
	return &Bee{apiKey: apiKey}
}

func (b *Bee) Connect(ctx context.Context) error {
	client, err := socketio.NewClient(
		"https://api.bee.computer",
		socketio.WithPath("sdk"),
		socketio.WithHeader("x-api-key", b.apiKey))
	if err != nil {
		return fmt.Errorf("socket.io connection failed: %w", err)
	}
	b.client = client

	// Handle all events
	//b.client.OnEvent(func(event string, data interface{}) {
	//	fmt.Printf("Received event '%s' with data: %v\n", event, data)
	//})

	// Handle connection
	b.client.OnConnect(func(c socketio.Conn) error {
		fmt.Println("Connected to server")
		return nil
	})

	// Handle disconnection
	b.client.OnDisconnect(func(c socketio.Conn, msg string) {
		fmt.Println("Disconnected from server", msg)
	})

	err = b.client.Connect()
	if err != nil {
		return err
	}

	return nil
}

func (b *Bee) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

var WebsocketCmd = &cobra.Command{
	Use:   "websocket",
	Short: "Connect to the Bee API websocket",
	Run: func(cmd *cobra.Command, args []string) {
		websocket_()
	},
}

func websocket_() {
	apiKey := os.Getenv("BEE_API_KEY")
	if apiKey == "" {
		log.Fatal("BEE_API_KEY environment variable is not set")
	}

	bee := NewBee(apiKey)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := bee.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer func(client *socketio.Client) {
		_ = client.Close()
	}(bee.client)

	go bee.Listen(ctx)

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Shutting down...")
}
