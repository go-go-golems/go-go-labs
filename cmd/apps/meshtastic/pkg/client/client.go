package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/serial"
)

// Constants
const (
	BROADCAST_ADDR = 0xFFFFFFFF
	BROADCAST_NUM  = 0xFFFFFFFF
)

// MeshtasticClient provides a high-level interface to a Meshtastic device
type MeshtasticClient struct {
	iface  *serial.SerialInterface
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex

	// Device state
	myInfo       *pb.MyNodeInfo
	nodes        map[uint32]*pb.NodeInfo
	channels     map[uint32]*pb.Channel
	config       *pb.LocalConfig
	moduleConfig *pb.LocalModuleConfig

	// Message handling
	packetID  uint32
	responses map[uint32]chan *pb.AdminMessage

	// Event handlers
	onMessage   func(*pb.MeshPacket)
	onNodeInfo  func(*pb.NodeInfo)
	onPosition  func(*pb.Position)
	onTelemetry func(*pb.Telemetry)
	onLogLine   func(string)
}

// Config represents client configuration
type Config struct {
	DevicePath  string
	Timeout     time.Duration
	DebugOutput interface{}
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout: 10 * time.Second,
	}
}

// NewMeshtasticClient creates a new Meshtastic client
func NewMeshtasticClient(config *Config) (*MeshtasticClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create serial interface
	serialConfig := serial.DefaultConfig()
	serialConfig.DevicePath = config.DevicePath
	if config.DebugOutput != nil {
		if writer, ok := config.DebugOutput.(interface{ Write([]byte) (int, error) }); ok {
			serialConfig.DebugOutput = writer
		}
	}

	iface, err := serial.NewSerialInterface(serialConfig)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "failed to create serial interface")
	}

	client := &MeshtasticClient{
		iface:     iface,
		ctx:       ctx,
		cancel:    cancel,
		nodes:     make(map[uint32]*pb.NodeInfo),
		channels:  make(map[uint32]*pb.Channel),
		responses: make(map[uint32]chan *pb.AdminMessage),
		packetID:  uint32(time.Now().Unix()),
	}

	// Set up event handlers
	iface.SetOnFromRadio(client.handleFromRadio)
	iface.SetOnLogOutput(client.handleLogOutput)
	iface.SetOnDisconnect(client.handleDisconnect)

	// Initialize the device connection
	if err := client.initialize(); err != nil {
		client.Close()
		return nil, errors.Wrap(err, "failed to initialize device")
	}

	return client, nil
}

// initialize initializes the device connection
func (c *MeshtasticClient) initialize() error {
	log.Info().Msg("Initializing Meshtastic device")

	// Send want config to get device info
	if err := c.sendWantConfig(); err != nil {
		return errors.Wrap(err, "failed to send want config")
	}

	// Wait for device to respond with initial data
	time.Sleep(2 * time.Second)

	log.Info().Msg("Meshtastic device initialized")
	return nil
}

// sendWantConfig sends a want config message to get device info
func (c *MeshtasticClient) sendWantConfig() error {
	toRadio := &pb.ToRadio{
		PayloadVariant: &pb.ToRadio_WantConfigId{
			WantConfigId: c.nextPacketID(),
		},
	}

	return c.iface.SendToRadio(toRadio)
}

// handleFromRadio handles messages from the device
func (c *MeshtasticClient) handleFromRadio(fromRadio *pb.FromRadio) {
	switch payload := fromRadio.PayloadVariant.(type) {
	case *pb.FromRadio_Packet:
		c.handleMeshPacket(payload.Packet)
	case *pb.FromRadio_MyInfo:
		c.handleMyInfo(payload.MyInfo)
	case *pb.FromRadio_NodeInfo:
		c.handleNodeInfo(payload.NodeInfo)
	case *pb.FromRadio_Config:
		c.handleConfig(payload.Config)
	case *pb.FromRadio_ModuleConfig:
		c.handleModuleConfig(payload.ModuleConfig)
	case *pb.FromRadio_Channel:
		c.handleChannel(payload.Channel)
	case *pb.FromRadio_ConfigCompleteId:
		log.Debug().Uint32("id", payload.ConfigCompleteId).Msg("Config complete")
	case *pb.FromRadio_Rebooted:
		log.Info().Msg("Device rebooted")
	default:
		log.Debug().Str("type", fmt.Sprintf("%T", payload)).Msg("Unhandled FromRadio message")
	}
}

// handleMeshPacket handles mesh packets
func (c *MeshtasticClient) handleMeshPacket(packet *pb.MeshPacket) {
	decoded := packet.GetDecoded()
	if decoded != nil {
		switch decoded.GetPortnum() {
		case pb.PortNum_TEXT_MESSAGE_APP:
			log.Info().
				Uint32("from", packet.GetFrom()).
				Uint32("to", packet.GetTo()).
				Str("message", string(decoded.GetPayload())).
				Msg("Text message received")
		case pb.PortNum_NODEINFO_APP:
			var nodeInfo pb.User
			if err := proto.Unmarshal(decoded.GetPayload(), &nodeInfo); err == nil {
				log.Info().
					Uint32("from", packet.GetFrom()).
					Str("longName", nodeInfo.GetLongName()).
					Str("shortName", nodeInfo.GetShortName()).
					Msg("Node info received")
			}
		case pb.PortNum_POSITION_APP:
			var position pb.Position
			if err := proto.Unmarshal(decoded.GetPayload(), &position); err == nil {
				lat := float64(position.GetLatitudeI()) * 1e-7
				lon := float64(position.GetLongitudeI()) * 1e-7
				log.Info().
					Uint32("from", packet.GetFrom()).
					Float64("lat", lat).
					Float64("lon", lon).
					Msg("Position received")

				if c.onPosition != nil {
					c.onPosition(&position)
				}
			}
		case pb.PortNum_TELEMETRY_APP:
			var telemetry pb.Telemetry
			if err := proto.Unmarshal(decoded.GetPayload(), &telemetry); err == nil {
				log.Info().
					Uint32("from", packet.GetFrom()).
					Msg("Telemetry received")

				if c.onTelemetry != nil {
					c.onTelemetry(&telemetry)
				}
			}
		}
	}

	if c.onMessage != nil {
		c.onMessage(packet)
	}
}

// handleMyInfo handles my node info
func (c *MeshtasticClient) handleMyInfo(myInfo *pb.MyNodeInfo) {
	c.mu.Lock()
	c.myInfo = myInfo
	c.mu.Unlock()

	log.Info().
		Uint32("nodeNum", myInfo.GetMyNodeNum()).
		Uint32("rebootCount", myInfo.GetRebootCount()).
		Msg("My node info received")
}

// handleNodeInfo handles node info updates
func (c *MeshtasticClient) handleNodeInfo(nodeInfo *pb.NodeInfo) {
	c.mu.Lock()
	c.nodes[nodeInfo.Num] = nodeInfo
	c.mu.Unlock()

	log.Info().
		Uint32("nodeNum", nodeInfo.Num).
		Str("longName", nodeInfo.GetUser().GetLongName()).
		Str("shortName", nodeInfo.GetUser().GetShortName()).
		Msg("Node info updated")

	if c.onNodeInfo != nil {
		c.onNodeInfo(nodeInfo)
	}
}

// handleConfig handles device configuration
func (c *MeshtasticClient) handleConfig(config *pb.Config) {
	c.mu.Lock()
	if c.config == nil {
		c.config = &pb.LocalConfig{}
	}

	// Update the appropriate config section
	switch payload := config.PayloadVariant.(type) {
	case *pb.Config_Device:
		c.config.Device = payload.Device
	case *pb.Config_Position:
		c.config.Position = payload.Position
	case *pb.Config_Power:
		c.config.Power = payload.Power
	case *pb.Config_Network:
		c.config.Network = payload.Network
	case *pb.Config_Display:
		c.config.Display = payload.Display
	case *pb.Config_Lora:
		c.config.Lora = payload.Lora
	case *pb.Config_Bluetooth:
		c.config.Bluetooth = payload.Bluetooth
	}
	c.mu.Unlock()

	log.Debug().Str("type", fmt.Sprintf("%T", config.PayloadVariant)).Msg("Config received")
}

// handleModuleConfig handles module configuration
func (c *MeshtasticClient) handleModuleConfig(moduleConfig *pb.ModuleConfig) {
	c.mu.Lock()
	if c.moduleConfig == nil {
		c.moduleConfig = &pb.LocalModuleConfig{}
	}

	// Update the appropriate module config section
	switch payload := moduleConfig.PayloadVariant.(type) {
	case *pb.ModuleConfig_Mqtt:
		c.moduleConfig.Mqtt = payload.Mqtt
	case *pb.ModuleConfig_Serial:
		c.moduleConfig.Serial = payload.Serial
	case *pb.ModuleConfig_ExternalNotification:
		c.moduleConfig.ExternalNotification = payload.ExternalNotification
	case *pb.ModuleConfig_StoreForward:
		c.moduleConfig.StoreForward = payload.StoreForward
	case *pb.ModuleConfig_RangeTest:
		c.moduleConfig.RangeTest = payload.RangeTest
	case *pb.ModuleConfig_Telemetry:
		c.moduleConfig.Telemetry = payload.Telemetry
	case *pb.ModuleConfig_CannedMessage:
		c.moduleConfig.CannedMessage = payload.CannedMessage
	case *pb.ModuleConfig_Audio:
		c.moduleConfig.Audio = payload.Audio
	case *pb.ModuleConfig_RemoteHardware:
		c.moduleConfig.RemoteHardware = payload.RemoteHardware
	case *pb.ModuleConfig_NeighborInfo:
		c.moduleConfig.NeighborInfo = payload.NeighborInfo
	case *pb.ModuleConfig_AmbientLighting:
		c.moduleConfig.AmbientLighting = payload.AmbientLighting
	case *pb.ModuleConfig_DetectionSensor:
		c.moduleConfig.DetectionSensor = payload.DetectionSensor
	case *pb.ModuleConfig_Paxcounter:
		c.moduleConfig.Paxcounter = payload.Paxcounter
	}
	c.mu.Unlock()

	log.Debug().Str("type", fmt.Sprintf("%T", moduleConfig.PayloadVariant)).Msg("Module config received")
}

// handleChannel handles channel configuration
func (c *MeshtasticClient) handleChannel(channel *pb.Channel) {
	c.mu.Lock()
	c.channels[uint32(channel.GetIndex())] = channel
	c.mu.Unlock()

	log.Debug().Int32("index", channel.GetIndex()).Str("name", channel.GetSettings().GetName()).Msg("Channel received")
}

// handleLogOutput handles device log output
func (c *MeshtasticClient) handleLogOutput(line string) {
	log.Debug().Str("device", line).Msg("Device log")

	if c.onLogLine != nil {
		c.onLogLine(line)
	}
}

// handleDisconnect handles device disconnection
func (c *MeshtasticClient) handleDisconnect(err error) {
	log.Error().Err(err).Msg("Device disconnected")
}

// nextPacketID returns the next packet ID
func (c *MeshtasticClient) nextPacketID() uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.packetID++
	return c.packetID
}

// SendText sends a text message
func (c *MeshtasticClient) SendText(text string, destination uint32) error {
	packet := &pb.MeshPacket{
		To:      destination,
		Id:      c.nextPacketID(),
		Channel: 0,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_TEXT_MESSAGE_APP,
				Payload: []byte(text),
			},
		},
	}

	return c.sendPacket(packet)
}

// sendPacket sends a mesh packet
func (c *MeshtasticClient) sendPacket(packet *pb.MeshPacket) error {
	toRadio := &pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: packet,
		},
	}

	return c.iface.SendToRadio(toRadio)
}

// GetMyInfo returns my node information
func (c *MeshtasticClient) GetMyInfo() *pb.MyNodeInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.myInfo
}

// GetNodes returns all known nodes
func (c *MeshtasticClient) GetNodes() map[uint32]*pb.NodeInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[uint32]*pb.NodeInfo)
	for k, v := range c.nodes {
		result[k] = v
	}
	return result
}

// GetChannels returns all channels
func (c *MeshtasticClient) GetChannels() map[uint32]*pb.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[uint32]*pb.Channel)
	for k, v := range c.channels {
		result[k] = v
	}
	return result
}

// IsConnected returns true if the device is connected
func (c *MeshtasticClient) IsConnected() bool {
	return c.iface.IsConnected()
}

// DevicePath returns the device path
func (c *MeshtasticClient) DevicePath() string {
	return c.iface.DevicePath()
}

// Event handler setters
func (c *MeshtasticClient) SetOnMessage(handler func(*pb.MeshPacket)) {
	c.onMessage = handler
}

func (c *MeshtasticClient) SetOnNodeInfo(handler func(*pb.NodeInfo)) {
	c.onNodeInfo = handler
}

func (c *MeshtasticClient) SetOnPosition(handler func(*pb.Position)) {
	c.onPosition = handler
}

func (c *MeshtasticClient) SetOnTelemetry(handler func(*pb.Telemetry)) {
	c.onTelemetry = handler
}

func (c *MeshtasticClient) SetOnLogLine(handler func(string)) {
	c.onLogLine = handler
}

// Close closes the client
func (c *MeshtasticClient) Close() error {
	log.Info().Msg("Closing Meshtastic client")

	c.cancel()

	if c.iface != nil {
		return c.iface.Close()
	}

	return nil
}
