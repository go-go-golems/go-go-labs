package main

import (
	"github.com/go-go-golems/go-go-labs/cmd/tests/plugin-test/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"net/rpc"
	"os"
)

type GreeterHello struct {
	logger hclog.Logger
}

func (g *GreeterHello) Foobar(i int, f float64, s string) string {
	return "Foobar"
}

func (g *GreeterHello) Greet() string {
	g.logger.Info("Hello!")
	return "Hello!"
}

func (g *GreeterHello) Foobar2() string {
	g.logger.Info("Foobar!")
	return "Foobar!"
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GLAZED_PLUGIN",
	MagicCookieValue: "glazed",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Debug,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	greeter := &GreeterHello{
		logger: logger,
	}

	plugin_ := &BothPlugin{
		Impl:       greeter,
		FoobarImpl: greeter,
	}

	var pluginMap = map[string]plugin.Plugin{
		// we could do a meld of the plugins here
		"greeter": plugin_,
		"foobar":  plugin_,
	}
	logger.Info("message from plugin")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}

type BothPlugin struct {
	Impl       shared.Greeter
	FoobarImpl shared.Foobar
}

func (p *BothPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return &shared.BothRPCServer{Impl: p.Impl, FoobarImpl: p.FoobarImpl}, nil
}

func (p *BothPlugin) Client(broker *plugin.MuxBroker, client *rpc.Client) (interface{}, error) {
	return nil, errors.New("not implemented")
}
