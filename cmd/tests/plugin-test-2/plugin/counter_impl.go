// counter-plugin/main.go
package main

import (
	"github.com/go-go-golems/go-go-labs/cmd/tests/plugin-test-2/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
)

type CounterImpl struct {
	current int
}

func (c *CounterImpl) Add(x int) int {
	c.current += x
	return c.current
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Debug,
		Output: os.Stderr,
	})

	counter := &CounterImpl{}

	handshakeConfig := plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "MAGIC_COOKIE_KEY",
		MagicCookieValue: "MAGIC_COOKIE_VALUE",
	}

	pluginMap := map[string]plugin.Plugin{
		"counter": &shared.CounterPlugin{Impl: counter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Logger:          logger,
	})
}
