package shared

import (
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

type Counter interface {
	Add(x int) int
}

type CounterRPC struct {
	client *rpc.Client
}

func (c *CounterRPC) Add(x int) int {
	var resp int
	err := c.client.Call("Plugin.Add", x, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

type CounterRPCServer struct {
	Impl Counter
}

func (s *CounterRPCServer) Add(args int, resp *int) error {
	*resp = s.Impl.Add(args)
	return nil
}

type CounterPlugin struct {
	Impl Counter
}

func (p *CounterPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return &CounterRPCServer{Impl: p.Impl}, nil
}

func (p *CounterPlugin) Client(broker *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &CounterRPC{client: c}, nil
}
