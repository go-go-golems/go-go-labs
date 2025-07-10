package generated

import (
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestMeshPacketSerialization(t *testing.T) {
	// Create a simple test packet
	packet := &MeshPacket{
		Id:       123,
		From:     456,
		To:       789,
		Priority: MeshPacket_UNSET,
		Channel:  1,
	}

	// Marshal the packet
	data, err := proto.Marshal(packet)
	if err != nil {
		t.Fatalf("Failed to marshal packet: %v", err)
	}

	// Unmarshal the packet
	var newPacket MeshPacket
	err = proto.Unmarshal(data, &newPacket)
	if err != nil {
		t.Fatalf("Failed to unmarshal packet: %v", err)
	}

	// Check the values
	if newPacket.Id != 123 {
		t.Errorf("Expected ID 123, got %d", newPacket.Id)
	}
	if newPacket.From != 456 {
		t.Errorf("Expected From 456, got %d", newPacket.From)
	}
	if newPacket.To != 789 {
		t.Errorf("Expected To 789, got %d", newPacket.To)
	}
}

func TestConfigSerialization(t *testing.T) {
	// Create a device config
	deviceConfig := &Config_DeviceConfig{
		Role:          Config_DeviceConfig_CLIENT,
		SerialEnabled: true,
		ButtonGpio:    17,
		BuzzerGpio:    18,
	}

	config := &Config{
		PayloadVariant: &Config_Device{
			Device: deviceConfig,
		},
	}

	// Marshal the config
	data, err := proto.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal the config
	var newConfig Config
	err = proto.Unmarshal(data, &newConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Check the values
	if device := newConfig.GetDevice(); device != nil {
		if device.Role != Config_DeviceConfig_CLIENT {
			t.Errorf("Expected CLIENT role, got %v", device.Role)
		}
		if !device.SerialEnabled {
			t.Error("Expected SerialEnabled to be true")
		}
		if device.ButtonGpio != 17 {
			t.Errorf("Expected ButtonGpio 17, got %d", device.ButtonGpio)
		}
	} else {
		t.Error("Expected device config to be present")
	}
}

func TestPortNums(t *testing.T) {
	// Test that port numbers are accessible
	if PortNum_TEXT_MESSAGE_APP == 0 {
		t.Error("Expected TEXT_MESSAGE_APP to be non-zero")
	}
	if PortNum_NODEINFO_APP == 0 {
		t.Error("Expected NODEINFO_APP to be non-zero")
	}
	if PortNum_ROUTING_APP == 0 {
		t.Error("Expected ROUTING_APP to be non-zero")
	}
	if PortNum_ADMIN_APP == 0 {
		t.Error("Expected ADMIN_APP to be non-zero")
	}
}

func TestAdminMessage(t *testing.T) {
	// Create an admin message
	adminMsg := &AdminMessage{
		SessionPasskey: []byte("test-session-key"),
		PayloadVariant: &AdminMessage_GetDeviceMetadataRequest{
			GetDeviceMetadataRequest: true,
		},
	}

	// Marshal the admin message
	data, err := proto.Marshal(adminMsg)
	if err != nil {
		t.Fatalf("Failed to marshal admin message: %v", err)
	}

	// Unmarshal the admin message
	var newAdminMsg AdminMessage
	err = proto.Unmarshal(data, &newAdminMsg)
	if err != nil {
		t.Fatalf("Failed to unmarshal admin message: %v", err)
	}

	// Check the values
	if string(newAdminMsg.SessionPasskey) != "test-session-key" {
		t.Errorf("Expected session passkey 'test-session-key', got %s", string(newAdminMsg.SessionPasskey))
	}
	if !newAdminMsg.GetGetDeviceMetadataRequest() {
		t.Error("Expected GetDeviceMetadataRequest to be true")
	}
}
