package machinery

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type Machine struct {
	Router *message.Router
	PubSub *gochannel.GoChannel
}

func NewMachine() (*Machine, error) {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewStdLogger(false, false))
	r, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	if err != nil {
		return nil, err
	}

	return &Machine{
		Router: r,
		PubSub: pubSub,
	}, nil
}

func (r *Machine) Publish(topic string, messages ...*message.Message) error {
	return r.PubSub.Publish(topic, messages...)
}
