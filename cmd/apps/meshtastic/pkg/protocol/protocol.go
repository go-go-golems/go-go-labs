package protocol

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Message struct {
	From    string
	To      string
	Content string
	Type    MessageType
}

type MessageType int

const (
	MessageTypeText MessageType = iota
	MessageTypePosition
	MessageTypeNodeInfo
	MessageTypeConfig
)

type DeviceInfo struct {
	NodeID     string
	Firmware   string
	Hardware   string
	MacAddress string
	Channel    string
}

type Protocol struct {
	// TODO: Add protocol-specific fields
}

func New() *Protocol {
	return &Protocol{}
}

func (p *Protocol) GetDeviceInfo(ctx context.Context) (*DeviceInfo, error) {
	log.Debug().Msg("Getting device info")

	// TODO: Implement device info retrieval
	return &DeviceInfo{
		NodeID:     "placeholder",
		Firmware:   "0.0.0",
		Hardware:   "unknown",
		MacAddress: "00:00:00:00:00:00",
		Channel:    "default",
	}, nil
}

func (p *Protocol) SendMessage(ctx context.Context, message *Message) error {
	log.Debug().Str("to", message.To).Str("content", message.Content).Msg("Sending message")

	// TODO: Implement message sending
	return errors.New("not implemented")
}

func (p *Protocol) ReceiveMessage(ctx context.Context) (*Message, error) {
	log.Debug().Msg("Receiving message")

	// TODO: Implement message receiving
	return nil, errors.New("not implemented")
}

func (p *Protocol) StartListening(ctx context.Context, messageHandler func(*Message)) error {
	log.Debug().Msg("Starting to listen for messages")

	// TODO: Implement message listening
	return errors.New("not implemented")
}
