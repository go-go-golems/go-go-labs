package shared

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type TemplateFuncs interface {
	GetFuncNames() []string
	CallFunction(call FunctionCall) (interface{}, error)
}

type FunctionCall struct {
	FuncName string
	Args     []interface{}
}

type TemplateFuncsRPC struct{ client *rpc.Client }

func (g *TemplateFuncsRPC) GetFuncNames() []string {
	var resp []string
	err := g.client.Call("Plugin.GetFuncNames", new(interface{}), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func (g *TemplateFuncsRPC) CallFunction(args FunctionCall) (interface{}, error) {
	var resp interface{}
	err := g.client.Call("Plugin.CallFunction", args, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type TemplateFuncsRPCServer struct {
	Impl TemplateFuncs
}

func (s *TemplateFuncsRPCServer) GetFuncNames(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetFuncNames()
	return nil
}

func (s *TemplateFuncsRPCServer) CallFunction(args FunctionCall, resp *interface{}) error {
	r, err := s.Impl.CallFunction(args)
	if err != nil {
		return err
	}
	*resp = r
	return nil
}

type TemplateFuncsPlugin struct {
	Impl TemplateFuncs
}

func (p *TemplateFuncsPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &TemplateFuncsRPCServer{Impl: p.Impl}, nil
}

func (TemplateFuncsPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &TemplateFuncsRPC{client: c}, nil
}
