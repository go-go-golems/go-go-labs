package main

import (
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/funcmap-plugin/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
)

type Plugin struct{}

func (p *Plugin) GetFuncNames() []string {
	return []string{"double"}
}

func (p *Plugin) CallFunction(call shared.FunctionCall) (interface{}, error) {
	if call.FuncName == "double" {
		if len(call.Args) > 0 {
			if val, ok := call.Args[0].(int); ok {
				return val * 2, nil
			}
		}
	}
	return nil, fmt.Errorf("invalid function or arguments")
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Debug,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	tf := &Plugin{}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "MAGIC_COOKIE_KEY",
			MagicCookieValue: "MAGIC_COOKIE_VALUE",
		},
		Plugins: map[string]plugin.Plugin{
			"TemplateFuncs": &shared.TemplateFuncsPlugin{Impl: tf},
		},
		Logger: logger,
	})
}
