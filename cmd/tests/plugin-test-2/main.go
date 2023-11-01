// main.go
package main

import (
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/tests/plugin-test-2/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "MAGIC_COOKIE_KEY",
			MagicCookieValue: "MAGIC_COOKIE_VALUE",
		},
		Plugins: map[string]plugin.Plugin{
			"counter": &shared.CounterPlugin{},
		},
		Cmd:    exec.Command("plugin/counter"), // change to the path of the compiled counter-plugin binary
		Logger: logger,
	})

	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		log.Fatalf("Error establishing RPC client: %s", err)
	}

	raw, err := rpcClient.Dispense("counter")
	if err != nil {
		log.Fatalf("Error dispensing plugin: %s", err)
	}

	counter := raw.(shared.Counter)
	for {
		fmt.Println(counter.Add(5))  // Outputs: 5
		fmt.Println(counter.Add(10)) // Outputs: 15
		time.Sleep(1 * time.Second)

	}
}
