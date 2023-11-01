package main

import (
	"github.com/go-go-golems/go-go-labs/cmd/experiments/funcmap-plugin/shared"
	"github.com/hashicorp/go-plugin"
	"os"
	"os/exec"
	"text/template"
)

// createFuncStubs creates stub functions that forward the function calls to the plugin.
func createFuncStubs(plugin shared.TemplateFuncs) map[string]interface{} {
	stubs := map[string]interface{}{}
	for _, funcName := range plugin.GetFuncNames() {
		stubs[funcName] = func(args ...interface{}) (interface{}, error) {
			return plugin.CallFunction(shared.FunctionCall{
				FuncName: funcName,
				Args:     args,
			})
		}
	}
	return stubs
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MAGIC_COOKIE_KEY",
	MagicCookieValue: "MAGIC_COOKIE_VALUE",
}

var pluginMap = map[string]plugin.Plugin{
	"TemplateFuncs": &shared.TemplateFuncsPlugin{},
}

func main() {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command("./plugin/funcmap-plugin"), // change to the path of the compiled funcmap-plugin binary
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		panic(err)
	}

	raw, err := rpcClient.Dispense("TemplateFuncs")
	if err != nil {
		panic(err)
	}

	templateFuncsPlugin := raw.(shared.TemplateFuncs)

	// Create stub functions.
	funcStubs := createFuncStubs(templateFuncsPlugin)

	// Initialize the template with the stub functions.
	tmpl := template.New("test").Funcs(funcStubs)

	// Parse the template.
	tmpl, err = tmpl.Parse(`
YOYOYO {{ double 5 }} YOYOYO

`)
	if err != nil {
		panic(err)
	}

	// Execute the template.
	err = tmpl.Execute(os.Stdout, nil)
	if err != nil {
		panic(err)
	}
}
