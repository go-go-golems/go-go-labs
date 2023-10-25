package main

import (
	"github.com/go-go-golems/go-go-labs/cmd/tests/plugin-test/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"os/exec"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stderr,
		Level:  hclog.Debug,
	})

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "GLAZED_PLUGIN",
			MagicCookieValue: "glazed",
		},
		Plugins: map[string]plugin.Plugin{
			"greeter": &shared.GreeterPlugin{},
			"foobar":  &shared.FoobarPlugin{},
		},
		Cmd:    exec.Command("./plugin/greeter"),
		Logger: logger,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		logger.Error("Failed to create RPC client", "error", err)
		os.Exit(1)
	}

	raw, err := rpcClient.Dispense("greeter")
	if err != nil {
		logger.Error("Failed to dispense plugin", "error", err)
		os.Exit(1)
	}
	rawFoobar, err := rpcClient.Dispense("foobar")
	if err != nil {
		logger.Error("Failed to dispense plugin", "error", err)
		os.Exit(1)
	}

	greeter := raw.(shared.Greeter)
	foobar := rawFoobar.(shared.Foobar)
	logger.Info("Calling plugin")
	greeting := greeter.Greet()
	foobaring := foobar.Foobar()
	logger.Info("Got response from plugin", "greeting", greeting, "foobar", foobaring)
}
