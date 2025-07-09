package client

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/protocol"
)

// StreamClient implements the StreamInterface for stream-based communication
type StreamClient struct {
	// Core state
	stream       io.ReadWriteCloser
	parser       *protocol.FrameParser
	builder      *protocol.FrameBuilder
	state        DeviceState
	stateHandler StateHandler

	// Context and synchronization
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex

	// Device information
	myInfo       *pb.MyNodeInfo
	nodes        map[uint32]*pb.NodeInfo
	channels     map[uint32]*pb.Channel
	config       *pb.LocalConfig
	moduleConfig *pb.LocalModuleConfig

	// Message handling
	messageQueue MessageQueue
	packetID     uint32
	responses    map[uint32]chan *pb.AdminMessage

	// Event handlers
	onMessage    func(*pb.MeshPacket)
	onNodeInfo   func(*pb.NodeInfo)
	onPosition   func(*pb.Position)
	onTelemetry  func(*pb.Telemetry)
	onLogLine    func(string)
	onDisconnect func(error)

	// Configuration
	devicePath string
	timeout    time.Duration

	// Statistics
	stats ConnectionStatistics

	// Heartbeat
	heartbeatInterval time.Duration
	lastHeartbeat     time.Time

	// Configuration state
	configReceived bool
	configTimeout  *Timeout
}

// NewStreamClient creates a new stream-based client
func NewStreamClient(stream io.ReadWriteCloser, devicePath string) *StreamClient {
	ctx, cancel := context.WithCancel(context.Background())

	sc := &StreamClient{
		stream:            stream,
		devicePath:        devicePath,
		ctx:               ctx,
		cancel:            cancel,
		state:             StateDisconnected,
		timeout:           30 * time.Second,
		heartbeatInterval: 300 * time.Second,

		// Initialize maps
		nodes:     make(map[uint32]*pb.NodeInfo),
		channels:  make(map[uint32]*pb.Channel),
		responses: make(map[uint32]chan *pb.AdminMessage),

		// Initialize packet ID with timestamp
		packetID: uint32(time.Now().Unix()),

		// Initialize statistics
		stats: ConnectionStatistics{},
	}

	// Create protocol components
	sc.parser = protocol.NewFrameParser(sc.handleFrame, sc.handleLogByte)
	sc.builder = protocol.NewFrameBuilder()

	return sc
}

// Connect implements MeshInterface
func (sc *StreamClient) Connect(ctx context.Context) error {
	sc.changeState(StateConnecting)

	log.Info().Str("device", sc.devicePath).Msg("Connecting to device")

	// Start goroutines
	sc.wg.Add(3)
	go sc.readerLoop()
	go sc.writerLoop()
	go sc.heartbeatLoop()

	// Send wake-up sequence
	if err := sc.sendWakeup(); err != nil {
		sc.changeState(StateError)
		return errors.Wrap(err, "failed to send wake-up sequence")
	}

	// Start configuration phase
	sc.changeState(StateConfiguring)

	// Send want config and wait for response
	if err := sc.SendWantConfig(); err != nil {
		sc.changeState(StateError)
		return errors.Wrap(err, "failed to send want config")
	}

	// Wait for configuration to complete
	if err := sc.WaitForConfig(sc.timeout); err != nil {
		sc.changeState(StateError)
		return errors.Wrap(err, "failed to receive configuration")
	}

	sc.changeState(StateConnected)
	log.Info().Str("device", sc.devicePath).Msg("Successfully connected to device")

	return nil
}

// Disconnect implements MeshInterface
func (sc *StreamClient) Disconnect() error {
	log.Info().Str("device", sc.devicePath).Msg("Disconnecting from device")

	sc.changeState(StateDisconnected)

	// Cancel context to stop goroutines
	sc.cancel()

	// Wait for goroutines to finish
	sc.wg.Wait()

	// Close stream
	if sc.stream != nil {
		if err := sc.stream.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing stream")
		}
	}

	return nil
}

// IsConnected implements MeshInterface
func (sc *StreamClient) IsConnected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.state == StateConnected
}

// GetMyInfo implements MeshInterface
func (sc *StreamClient) GetMyInfo() *pb.MyNodeInfo {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.myInfo
}

// GetNodes implements MeshInterface
func (sc *StreamClient) GetNodes() map[uint32]*pb.NodeInfo {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[uint32]*pb.NodeInfo)
	for k, v := range sc.nodes {
		result[k] = v
	}
	return result
}

// GetChannels implements MeshInterface
func (sc *StreamClient) GetChannels() map[uint32]*pb.Channel {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[uint32]*pb.Channel)
	for k, v := range sc.channels {
		result[k] = v
	}
	return result
}

// GetConfig implements MeshInterface
func (sc *StreamClient) GetConfig() *pb.LocalConfig {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.config
}

// GetModuleConfig implements MeshInterface
func (sc *StreamClient) GetModuleConfig() *pb.LocalModuleConfig {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.moduleConfig
}

// SendMessage implements MeshInterface
func (sc *StreamClient) SendMessage(packet *pb.MeshPacket) error {
	if !sc.IsConnected() {
		return errors.New("device not connected")
	}

	// Enqueue the packet
	if sc.messageQueue != nil {
		return sc.messageQueue.Enqueue(packet)
	}

	// Direct send if no queue
	return sc.sendPacket(packet)
}

// SendText implements MeshInterface
func (sc *StreamClient) SendText(text string, destination uint32) error {
	packet := &pb.MeshPacket{
		To:      destination,
		Id:      sc.nextPacketID(),
		Channel: 0,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_TEXT_MESSAGE_APP,
				Payload: []byte(text),
			},
		},
	}

	return sc.SendMessage(packet)
}

// DevicePath implements MeshInterface
func (sc *StreamClient) DevicePath() string {
	return sc.devicePath
}

// Close implements MeshInterface
func (sc *StreamClient) Close() error {
	return sc.Disconnect()
}

// GetStream implements StreamInterface
func (sc *StreamClient) GetStream() io.ReadWriteCloser {
	return sc.stream
}

// SetStream implements StreamInterface
func (sc *StreamClient) SetStream(stream io.ReadWriteCloser) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.stream = stream
}

// WaitForConfig implements StreamInterface
func (sc *StreamClient) WaitForConfig(timeout time.Duration) error {
	sc.configTimeout = NewTimeout(timeout)

	if sc.configTimeout.WaitForCondition(sc.ctx, func() bool {
		sc.mu.RLock()
		defer sc.mu.RUnlock()
		return sc.configReceived
	}) {
		return nil
	}

	return errors.New("timeout waiting for configuration")
}

// SendWantConfig implements StreamInterface
func (sc *StreamClient) SendWantConfig() error {
	toRadio := &pb.ToRadio{
		PayloadVariant: &pb.ToRadio_WantConfigId{
			WantConfigId: sc.nextPacketID(),
		},
	}

	return sc.sendToRadio(toRadio)
}

// SendAdminMessage implements StreamInterface
func (sc *StreamClient) SendAdminMessage(msg *pb.AdminMessage) (*pb.AdminMessage, error) {
	// Create response channel
	responseID := sc.nextPacketID()
	responseChan := make(chan *pb.AdminMessage, 1)

	sc.mu.Lock()
	sc.responses[responseID] = responseChan
	sc.mu.Unlock()

	// Clean up response channel
	defer func() {
		sc.mu.Lock()
		delete(sc.responses, responseID)
		sc.mu.Unlock()
		close(responseChan)
	}()

	// Send admin message
	packet := &pb.MeshPacket{
		To:      BROADCAST_NUM,
		Id:      responseID,
		Channel: 0,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_ADMIN_APP,
				Payload: mustMarshal(msg),
			},
		},
	}

	if err := sc.SendMessage(packet); err != nil {
		return nil, errors.Wrap(err, "failed to send admin message")
	}

	// Wait for response
	select {
	case response := <-responseChan:
		return response, nil
	case <-time.After(sc.timeout):
		return nil, errors.New("timeout waiting for admin response")
	case <-sc.ctx.Done():
		return nil, errors.New("context cancelled")
	}
}

// GetQueueStatus implements StreamInterface
func (sc *StreamClient) GetQueueStatus() (int, int) {
	if sc.messageQueue == nil {
		return 0, 0
	}
	return sc.messageQueue.Size(), 100 // Assume capacity of 100
}

// Event handler setters
func (sc *StreamClient) SetOnMessage(handler func(*pb.MeshPacket)) {
	sc.onMessage = handler
}

func (sc *StreamClient) SetOnNodeInfo(handler func(*pb.NodeInfo)) {
	sc.onNodeInfo = handler
}

func (sc *StreamClient) SetOnPosition(handler func(*pb.Position)) {
	sc.onPosition = handler
}

func (sc *StreamClient) SetOnTelemetry(handler func(*pb.Telemetry)) {
	sc.onTelemetry = handler
}

func (sc *StreamClient) SetOnLogLine(handler func(string)) {
	sc.onLogLine = handler
}

func (sc *StreamClient) SetOnDisconnect(handler func(error)) {
	sc.onDisconnect = handler
}

// Internal methods

func (sc *StreamClient) changeState(newState DeviceState) {
	sc.mu.Lock()
	oldState := sc.state
	sc.state = newState
	sc.mu.Unlock()

	log.Debug().
		Str("oldState", oldState.String()).
		Str("newState", newState.String()).
		Msg("State change")

	if sc.stateHandler != nil {
		sc.stateHandler.OnStateChange(oldState, newState)
	}
}

func (sc *StreamClient) nextPacketID() uint32 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.packetID++
	return sc.packetID
}

func (sc *StreamClient) sendWakeup() error {
	// Send wake-up sequence (32 bytes of START2)
	wakeup := make([]byte, 32)
	for i := range wakeup {
		wakeup[i] = protocol.START2
	}

	_, err := sc.stream.Write(wakeup)
	if err != nil {
		return errors.Wrap(err, "failed to send wake-up sequence")
	}

	// Wait for device to settle
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (sc *StreamClient) sendPacket(packet *pb.MeshPacket) error {
	toRadio := &pb.ToRadio{
		PayloadVariant: &pb.ToRadio_Packet{
			Packet: packet,
		},
	}

	return sc.sendToRadio(toRadio)
}

func (sc *StreamClient) sendToRadio(toRadio *pb.ToRadio) error {
	frameData, err := sc.builder.BuildFrame(toRadio)
	if err != nil {
		return errors.Wrap(err, "failed to build frame")
	}

	_, err = sc.stream.Write(frameData)
	if err != nil {
		sc.stats.WriteErrors++
		return errors.Wrap(err, "failed to write to stream")
	}

	sc.stats.BytesWritten += uint64(len(frameData))
	sc.stats.FramesSent++

	return nil
}

func (sc *StreamClient) readerLoop() {
	defer sc.wg.Done()

	log.Debug().Msg("Starting reader loop")

	buffer := make([]byte, 1024)

	for {
		select {
		case <-sc.ctx.Done():
			log.Debug().Msg("Reader loop cancelled")
			return
		default:
			n, err := sc.stream.Read(buffer)
			if err != nil {
				if err == io.EOF {
					log.Warn().Msg("Device disconnected (EOF)")
					sc.handleDisconnect(err)
					return
				}

				sc.stats.ReadErrors++
				log.Error().Err(err).Msg("Read error")
				sc.handleDisconnect(err)
				return
			}

			if n > 0 {
				sc.stats.BytesRead += uint64(n)
				sc.parser.ProcessBytes(buffer[:n])
			}
		}
	}
}

func (sc *StreamClient) writerLoop() {
	defer sc.wg.Done()

	log.Debug().Msg("Starting writer loop")

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sc.ctx.Done():
			log.Debug().Msg("Writer loop cancelled")
			return
		case <-ticker.C:
			if sc.messageQueue != nil && !sc.messageQueue.IsEmpty() {
				packet, err := sc.messageQueue.Dequeue()
				if err != nil {
					continue
				}

				if err := sc.sendPacket(packet); err != nil {
					log.Error().Err(err).Msg("Failed to send queued packet")
				}
			}
		}
	}
}

func (sc *StreamClient) heartbeatLoop() {
	defer sc.wg.Done()

	log.Debug().Msg("Starting heartbeat loop")

	ticker := time.NewTicker(sc.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sc.ctx.Done():
			log.Debug().Msg("Heartbeat loop cancelled")
			return
		case <-ticker.C:
			if sc.IsConnected() {
				// Send heartbeat
				if err := sc.sendHeartbeat(); err != nil {
					log.Error().Err(err).Msg("Failed to send heartbeat")
				}
			}
		}
	}
}

func (sc *StreamClient) sendHeartbeat() error {
	// Implementation depends on Meshtastic heartbeat protocol
	// For now, just update the timestamp
	sc.lastHeartbeat = time.Now()
	return nil
}

func (sc *StreamClient) handleFrame(frame *protocol.Frame) {
	sc.stats.FramesReceived++

	fromRadio, err := frame.FromRadioMessage()
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse FromRadio message")
		return
	}

	sc.handleFromRadio(fromRadio)
}

func (sc *StreamClient) handleLogByte(b byte) {
	// Handle device log output
	if sc.onLogLine != nil {
		sc.onLogLine(string(b))
	}
}

func (sc *StreamClient) handleFromRadio(fromRadio *pb.FromRadio) {
	switch payload := fromRadio.PayloadVariant.(type) {
	case *pb.FromRadio_Packet:
		sc.handleMeshPacket(payload.Packet)
	case *pb.FromRadio_MyInfo:
		sc.handleMyInfo(payload.MyInfo)
	case *pb.FromRadio_NodeInfo:
		sc.handleNodeInfo(payload.NodeInfo)
	case *pb.FromRadio_Config:
		sc.handleConfig(payload.Config)
	case *pb.FromRadio_ModuleConfig:
		sc.handleModuleConfig(payload.ModuleConfig)
	case *pb.FromRadio_Channel:
		sc.handleChannel(payload.Channel)
	case *pb.FromRadio_ConfigCompleteId:
		sc.handleConfigComplete(payload.ConfigCompleteId)
	case *pb.FromRadio_Rebooted:
		log.Info().Msg("Device rebooted")
	default:
		log.Debug().Str("type", fmt.Sprintf("%T", payload)).Msg("Unhandled FromRadio message")
	}
}

func (sc *StreamClient) handleMeshPacket(packet *pb.MeshPacket) {
	sc.stats.MessagesReceived++

	if sc.onMessage != nil {
		sc.onMessage(packet)
	}

	decoded := packet.GetDecoded()
	if decoded != nil {
		switch decoded.GetPortnum() {
		case pb.PortNum_POSITION_APP:
			var position pb.Position
			if err := proto.Unmarshal(decoded.GetPayload(), &position); err == nil {
				if sc.onPosition != nil {
					sc.onPosition(&position)
				}
			}
		case pb.PortNum_TELEMETRY_APP:
			var telemetry pb.Telemetry
			if err := proto.Unmarshal(decoded.GetPayload(), &telemetry); err == nil {
				if sc.onTelemetry != nil {
					sc.onTelemetry(&telemetry)
				}
			}
		case pb.PortNum_ADMIN_APP:
			var adminMsg pb.AdminMessage
			if err := proto.Unmarshal(decoded.GetPayload(), &adminMsg); err == nil {
				sc.handleAdminMessage(packet.GetId(), &adminMsg)
			}
		}
	}
}

func (sc *StreamClient) handleMyInfo(myInfo *pb.MyNodeInfo) {
	sc.mu.Lock()
	sc.myInfo = myInfo
	sc.mu.Unlock()

	log.Info().
		Uint32("nodeNum", myInfo.GetMyNodeNum()).
		Uint32("rebootCount", myInfo.GetRebootCount()).
		Msg("My node info received")
}

func (sc *StreamClient) handleNodeInfo(nodeInfo *pb.NodeInfo) {
	sc.mu.Lock()
	sc.nodes[nodeInfo.Num] = nodeInfo
	sc.mu.Unlock()

	log.Info().
		Uint32("nodeNum", nodeInfo.Num).
		Str("longName", nodeInfo.GetUser().GetLongName()).
		Str("shortName", nodeInfo.GetUser().GetShortName()).
		Msg("Node info updated")

	if sc.onNodeInfo != nil {
		sc.onNodeInfo(nodeInfo)
	}
}

func (sc *StreamClient) handleConfig(config *pb.Config) {
	sc.mu.Lock()
	if sc.config == nil {
		sc.config = &pb.LocalConfig{}
	}

	// Update the appropriate config section
	switch payload := config.PayloadVariant.(type) {
	case *pb.Config_Device:
		sc.config.Device = payload.Device
	case *pb.Config_Position:
		sc.config.Position = payload.Position
	case *pb.Config_Power:
		sc.config.Power = payload.Power
	case *pb.Config_Network:
		sc.config.Network = payload.Network
	case *pb.Config_Display:
		sc.config.Display = payload.Display
	case *pb.Config_Lora:
		sc.config.Lora = payload.Lora
	case *pb.Config_Bluetooth:
		sc.config.Bluetooth = payload.Bluetooth
	}
	sc.mu.Unlock()

	log.Debug().Str("type", fmt.Sprintf("%T", config.PayloadVariant)).Msg("Config received")
}

func (sc *StreamClient) handleModuleConfig(moduleConfig *pb.ModuleConfig) {
	sc.mu.Lock()
	if sc.moduleConfig == nil {
		sc.moduleConfig = &pb.LocalModuleConfig{}
	}

	// Update the appropriate module config section
	switch payload := moduleConfig.PayloadVariant.(type) {
	case *pb.ModuleConfig_Mqtt:
		sc.moduleConfig.Mqtt = payload.Mqtt
	case *pb.ModuleConfig_Serial:
		sc.moduleConfig.Serial = payload.Serial
	case *pb.ModuleConfig_ExternalNotification:
		sc.moduleConfig.ExternalNotification = payload.ExternalNotification
	case *pb.ModuleConfig_StoreForward:
		sc.moduleConfig.StoreForward = payload.StoreForward
	case *pb.ModuleConfig_RangeTest:
		sc.moduleConfig.RangeTest = payload.RangeTest
	case *pb.ModuleConfig_Telemetry:
		sc.moduleConfig.Telemetry = payload.Telemetry
	case *pb.ModuleConfig_CannedMessage:
		sc.moduleConfig.CannedMessage = payload.CannedMessage
	case *pb.ModuleConfig_Audio:
		sc.moduleConfig.Audio = payload.Audio
	case *pb.ModuleConfig_RemoteHardware:
		sc.moduleConfig.RemoteHardware = payload.RemoteHardware
	case *pb.ModuleConfig_NeighborInfo:
		sc.moduleConfig.NeighborInfo = payload.NeighborInfo
	case *pb.ModuleConfig_AmbientLighting:
		sc.moduleConfig.AmbientLighting = payload.AmbientLighting
	case *pb.ModuleConfig_DetectionSensor:
		sc.moduleConfig.DetectionSensor = payload.DetectionSensor
	case *pb.ModuleConfig_Paxcounter:
		sc.moduleConfig.Paxcounter = payload.Paxcounter
	}
	sc.mu.Unlock()

	log.Debug().Str("type", fmt.Sprintf("%T", moduleConfig.PayloadVariant)).Msg("Module config received")
}

func (sc *StreamClient) handleChannel(channel *pb.Channel) {
	sc.mu.Lock()
	sc.channels[uint32(channel.GetIndex())] = channel
	sc.mu.Unlock()

	log.Debug().Int32("index", channel.GetIndex()).Str("name", channel.GetSettings().GetName()).Msg("Channel received")
}

func (sc *StreamClient) handleConfigComplete(configID uint32) {
	sc.mu.Lock()
	sc.configReceived = true
	sc.mu.Unlock()

	log.Debug().Uint32("configID", configID).Msg("Config complete")
}

func (sc *StreamClient) handleAdminMessage(packetID uint32, adminMsg *pb.AdminMessage) {
	sc.mu.RLock()
	responseChan, exists := sc.responses[packetID]
	sc.mu.RUnlock()

	if exists {
		select {
		case responseChan <- adminMsg:
		default:
			log.Warn().Msg("Admin response channel full")
		}
	}
}

func (sc *StreamClient) handleDisconnect(err error) {
	sc.changeState(StateDisconnected)

	if sc.onDisconnect != nil {
		sc.onDisconnect(err)
	}
}

// Helper function to marshal protobuf messages
func mustMarshal(msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return data
}
