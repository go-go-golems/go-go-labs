package serial

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tarm/serial"
)

type Connection struct {
	port   *serial.Port
	config *serial.Config
}

type SerialConfig struct {
	Device   string
	Baudrate int
	Timeout  time.Duration
}

func NewConnection(config SerialConfig) (*Connection, error) {
	serialConfig := &serial.Config{
		Name:        config.Device,
		Baud:        config.Baudrate,
		ReadTimeout: config.Timeout,
	}

	port, err := serial.OpenPort(serialConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open serial port")
	}

	log.Info().Str("device", config.Device).Int("baud", config.Baudrate).Msg("Serial connection established")

	return &Connection{
		port:   port,
		config: serialConfig,
	}, nil
}

func (c *Connection) Close() error {
	if c.port != nil {
		return c.port.Close()
	}
	return nil
}

func (c *Connection) Write(data []byte) (int, error) {
	return c.port.Write(data)
}

func (c *Connection) Read(buf []byte) (int, error) {
	return c.port.Read(buf)
}

func (c *Connection) ReadWithContext(ctx context.Context, buf []byte) (int, error) {
	// TODO: Implement context-aware reading
	return c.port.Read(buf)
}

func (c *Connection) WriteWithContext(ctx context.Context, data []byte) (int, error) {
	// TODO: Implement context-aware writing
	return c.port.Write(data)
}

var _ io.ReadWriteCloser = (*Connection)(nil)
