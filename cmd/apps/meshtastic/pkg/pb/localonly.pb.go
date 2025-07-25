// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: meshtastic/localonly.proto

package generated

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LocalConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The part of the config that is specific to the Device
	Device *Config_DeviceConfig `protobuf:"bytes,1,opt,name=device,proto3" json:"device,omitempty"`
	// The part of the config that is specific to the GPS Position
	Position *Config_PositionConfig `protobuf:"bytes,2,opt,name=position,proto3" json:"position,omitempty"`
	// The part of the config that is specific to the Power settings
	Power *Config_PowerConfig `protobuf:"bytes,3,opt,name=power,proto3" json:"power,omitempty"`
	// The part of the config that is specific to the Wifi Settings
	Network *Config_NetworkConfig `protobuf:"bytes,4,opt,name=network,proto3" json:"network,omitempty"`
	// The part of the config that is specific to the Display
	Display *Config_DisplayConfig `protobuf:"bytes,5,opt,name=display,proto3" json:"display,omitempty"`
	// The part of the config that is specific to the Lora Radio
	Lora *Config_LoRaConfig `protobuf:"bytes,6,opt,name=lora,proto3" json:"lora,omitempty"`
	// The part of the config that is specific to the Bluetooth settings
	Bluetooth *Config_BluetoothConfig `protobuf:"bytes,7,opt,name=bluetooth,proto3" json:"bluetooth,omitempty"`
	// A version integer used to invalidate old save files when we make
	// incompatible changes This integer is set at build time and is private to
	// NodeDB.cpp in the device code.
	Version uint32 `protobuf:"varint,8,opt,name=version,proto3" json:"version,omitempty"`
	// The part of the config that is specific to Security settings
	Security      *Config_SecurityConfig `protobuf:"bytes,9,opt,name=security,proto3" json:"security,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LocalConfig) Reset() {
	*x = LocalConfig{}
	mi := &file_meshtastic_localonly_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LocalConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LocalConfig) ProtoMessage() {}

func (x *LocalConfig) ProtoReflect() protoreflect.Message {
	mi := &file_meshtastic_localonly_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LocalConfig.ProtoReflect.Descriptor instead.
func (*LocalConfig) Descriptor() ([]byte, []int) {
	return file_meshtastic_localonly_proto_rawDescGZIP(), []int{0}
}

func (x *LocalConfig) GetDevice() *Config_DeviceConfig {
	if x != nil {
		return x.Device
	}
	return nil
}

func (x *LocalConfig) GetPosition() *Config_PositionConfig {
	if x != nil {
		return x.Position
	}
	return nil
}

func (x *LocalConfig) GetPower() *Config_PowerConfig {
	if x != nil {
		return x.Power
	}
	return nil
}

func (x *LocalConfig) GetNetwork() *Config_NetworkConfig {
	if x != nil {
		return x.Network
	}
	return nil
}

func (x *LocalConfig) GetDisplay() *Config_DisplayConfig {
	if x != nil {
		return x.Display
	}
	return nil
}

func (x *LocalConfig) GetLora() *Config_LoRaConfig {
	if x != nil {
		return x.Lora
	}
	return nil
}

func (x *LocalConfig) GetBluetooth() *Config_BluetoothConfig {
	if x != nil {
		return x.Bluetooth
	}
	return nil
}

func (x *LocalConfig) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *LocalConfig) GetSecurity() *Config_SecurityConfig {
	if x != nil {
		return x.Security
	}
	return nil
}

type LocalModuleConfig struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The part of the config that is specific to the MQTT module
	Mqtt *ModuleConfig_MQTTConfig `protobuf:"bytes,1,opt,name=mqtt,proto3" json:"mqtt,omitempty"`
	// The part of the config that is specific to the Serial module
	Serial *ModuleConfig_SerialConfig `protobuf:"bytes,2,opt,name=serial,proto3" json:"serial,omitempty"`
	// The part of the config that is specific to the ExternalNotification module
	ExternalNotification *ModuleConfig_ExternalNotificationConfig `protobuf:"bytes,3,opt,name=external_notification,json=externalNotification,proto3" json:"external_notification,omitempty"`
	// The part of the config that is specific to the Store & Forward module
	StoreForward *ModuleConfig_StoreForwardConfig `protobuf:"bytes,4,opt,name=store_forward,json=storeForward,proto3" json:"store_forward,omitempty"`
	// The part of the config that is specific to the RangeTest module
	RangeTest *ModuleConfig_RangeTestConfig `protobuf:"bytes,5,opt,name=range_test,json=rangeTest,proto3" json:"range_test,omitempty"`
	// The part of the config that is specific to the Telemetry module
	Telemetry *ModuleConfig_TelemetryConfig `protobuf:"bytes,6,opt,name=telemetry,proto3" json:"telemetry,omitempty"`
	// The part of the config that is specific to the Canned Message module
	CannedMessage *ModuleConfig_CannedMessageConfig `protobuf:"bytes,7,opt,name=canned_message,json=cannedMessage,proto3" json:"canned_message,omitempty"`
	// The part of the config that is specific to the Audio module
	Audio *ModuleConfig_AudioConfig `protobuf:"bytes,9,opt,name=audio,proto3" json:"audio,omitempty"`
	// The part of the config that is specific to the Remote Hardware module
	RemoteHardware *ModuleConfig_RemoteHardwareConfig `protobuf:"bytes,10,opt,name=remote_hardware,json=remoteHardware,proto3" json:"remote_hardware,omitempty"`
	// The part of the config that is specific to the Neighbor Info module
	NeighborInfo *ModuleConfig_NeighborInfoConfig `protobuf:"bytes,11,opt,name=neighbor_info,json=neighborInfo,proto3" json:"neighbor_info,omitempty"`
	// The part of the config that is specific to the Ambient Lighting module
	AmbientLighting *ModuleConfig_AmbientLightingConfig `protobuf:"bytes,12,opt,name=ambient_lighting,json=ambientLighting,proto3" json:"ambient_lighting,omitempty"`
	// The part of the config that is specific to the Detection Sensor module
	DetectionSensor *ModuleConfig_DetectionSensorConfig `protobuf:"bytes,13,opt,name=detection_sensor,json=detectionSensor,proto3" json:"detection_sensor,omitempty"`
	// Paxcounter Config
	Paxcounter *ModuleConfig_PaxcounterConfig `protobuf:"bytes,14,opt,name=paxcounter,proto3" json:"paxcounter,omitempty"`
	// A version integer used to invalidate old save files when we make
	// incompatible changes This integer is set at build time and is private to
	// NodeDB.cpp in the device code.
	Version       uint32 `protobuf:"varint,8,opt,name=version,proto3" json:"version,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LocalModuleConfig) Reset() {
	*x = LocalModuleConfig{}
	mi := &file_meshtastic_localonly_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LocalModuleConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LocalModuleConfig) ProtoMessage() {}

func (x *LocalModuleConfig) ProtoReflect() protoreflect.Message {
	mi := &file_meshtastic_localonly_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LocalModuleConfig.ProtoReflect.Descriptor instead.
func (*LocalModuleConfig) Descriptor() ([]byte, []int) {
	return file_meshtastic_localonly_proto_rawDescGZIP(), []int{1}
}

func (x *LocalModuleConfig) GetMqtt() *ModuleConfig_MQTTConfig {
	if x != nil {
		return x.Mqtt
	}
	return nil
}

func (x *LocalModuleConfig) GetSerial() *ModuleConfig_SerialConfig {
	if x != nil {
		return x.Serial
	}
	return nil
}

func (x *LocalModuleConfig) GetExternalNotification() *ModuleConfig_ExternalNotificationConfig {
	if x != nil {
		return x.ExternalNotification
	}
	return nil
}

func (x *LocalModuleConfig) GetStoreForward() *ModuleConfig_StoreForwardConfig {
	if x != nil {
		return x.StoreForward
	}
	return nil
}

func (x *LocalModuleConfig) GetRangeTest() *ModuleConfig_RangeTestConfig {
	if x != nil {
		return x.RangeTest
	}
	return nil
}

func (x *LocalModuleConfig) GetTelemetry() *ModuleConfig_TelemetryConfig {
	if x != nil {
		return x.Telemetry
	}
	return nil
}

func (x *LocalModuleConfig) GetCannedMessage() *ModuleConfig_CannedMessageConfig {
	if x != nil {
		return x.CannedMessage
	}
	return nil
}

func (x *LocalModuleConfig) GetAudio() *ModuleConfig_AudioConfig {
	if x != nil {
		return x.Audio
	}
	return nil
}

func (x *LocalModuleConfig) GetRemoteHardware() *ModuleConfig_RemoteHardwareConfig {
	if x != nil {
		return x.RemoteHardware
	}
	return nil
}

func (x *LocalModuleConfig) GetNeighborInfo() *ModuleConfig_NeighborInfoConfig {
	if x != nil {
		return x.NeighborInfo
	}
	return nil
}

func (x *LocalModuleConfig) GetAmbientLighting() *ModuleConfig_AmbientLightingConfig {
	if x != nil {
		return x.AmbientLighting
	}
	return nil
}

func (x *LocalModuleConfig) GetDetectionSensor() *ModuleConfig_DetectionSensorConfig {
	if x != nil {
		return x.DetectionSensor
	}
	return nil
}

func (x *LocalModuleConfig) GetPaxcounter() *ModuleConfig_PaxcounterConfig {
	if x != nil {
		return x.Paxcounter
	}
	return nil
}

func (x *LocalModuleConfig) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

var File_meshtastic_localonly_proto protoreflect.FileDescriptor

const file_meshtastic_localonly_proto_rawDesc = "" +
	"\n" +
	"\x1ameshtastic/localonly.proto\x12\n" +
	"meshtastic\x1a\x17meshtastic/config.proto\x1a\x1emeshtastic/module_config.proto\"\x81\x04\n" +
	"\vLocalConfig\x127\n" +
	"\x06device\x18\x01 \x01(\v2\x1f.meshtastic.Config.DeviceConfigR\x06device\x12=\n" +
	"\bposition\x18\x02 \x01(\v2!.meshtastic.Config.PositionConfigR\bposition\x124\n" +
	"\x05power\x18\x03 \x01(\v2\x1e.meshtastic.Config.PowerConfigR\x05power\x12:\n" +
	"\anetwork\x18\x04 \x01(\v2 .meshtastic.Config.NetworkConfigR\anetwork\x12:\n" +
	"\adisplay\x18\x05 \x01(\v2 .meshtastic.Config.DisplayConfigR\adisplay\x121\n" +
	"\x04lora\x18\x06 \x01(\v2\x1d.meshtastic.Config.LoRaConfigR\x04lora\x12@\n" +
	"\tbluetooth\x18\a \x01(\v2\".meshtastic.Config.BluetoothConfigR\tbluetooth\x12\x18\n" +
	"\aversion\x18\b \x01(\rR\aversion\x12=\n" +
	"\bsecurity\x18\t \x01(\v2!.meshtastic.Config.SecurityConfigR\bsecurity\"\xae\b\n" +
	"\x11LocalModuleConfig\x127\n" +
	"\x04mqtt\x18\x01 \x01(\v2#.meshtastic.ModuleConfig.MQTTConfigR\x04mqtt\x12=\n" +
	"\x06serial\x18\x02 \x01(\v2%.meshtastic.ModuleConfig.SerialConfigR\x06serial\x12h\n" +
	"\x15external_notification\x18\x03 \x01(\v23.meshtastic.ModuleConfig.ExternalNotificationConfigR\x14externalNotification\x12P\n" +
	"\rstore_forward\x18\x04 \x01(\v2+.meshtastic.ModuleConfig.StoreForwardConfigR\fstoreForward\x12G\n" +
	"\n" +
	"range_test\x18\x05 \x01(\v2(.meshtastic.ModuleConfig.RangeTestConfigR\trangeTest\x12F\n" +
	"\ttelemetry\x18\x06 \x01(\v2(.meshtastic.ModuleConfig.TelemetryConfigR\ttelemetry\x12S\n" +
	"\x0ecanned_message\x18\a \x01(\v2,.meshtastic.ModuleConfig.CannedMessageConfigR\rcannedMessage\x12:\n" +
	"\x05audio\x18\t \x01(\v2$.meshtastic.ModuleConfig.AudioConfigR\x05audio\x12V\n" +
	"\x0fremote_hardware\x18\n" +
	" \x01(\v2-.meshtastic.ModuleConfig.RemoteHardwareConfigR\x0eremoteHardware\x12P\n" +
	"\rneighbor_info\x18\v \x01(\v2+.meshtastic.ModuleConfig.NeighborInfoConfigR\fneighborInfo\x12Y\n" +
	"\x10ambient_lighting\x18\f \x01(\v2..meshtastic.ModuleConfig.AmbientLightingConfigR\x0fambientLighting\x12Y\n" +
	"\x10detection_sensor\x18\r \x01(\v2..meshtastic.ModuleConfig.DetectionSensorConfigR\x0fdetectionSensor\x12I\n" +
	"\n" +
	"paxcounter\x18\x0e \x01(\v2).meshtastic.ModuleConfig.PaxcounterConfigR\n" +
	"paxcounter\x12\x18\n" +
	"\aversion\x18\b \x01(\rR\aversionBd\n" +
	"\x13com.geeksville.meshB\x0fLocalOnlyProtosZ\"github.com/meshtastic/go/generated\xaa\x02\x14Meshtastic.Protobufs\xba\x02\x00b\x06proto3"

var (
	file_meshtastic_localonly_proto_rawDescOnce sync.Once
	file_meshtastic_localonly_proto_rawDescData []byte
)

func file_meshtastic_localonly_proto_rawDescGZIP() []byte {
	file_meshtastic_localonly_proto_rawDescOnce.Do(func() {
		file_meshtastic_localonly_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_meshtastic_localonly_proto_rawDesc), len(file_meshtastic_localonly_proto_rawDesc)))
	})
	return file_meshtastic_localonly_proto_rawDescData
}

var file_meshtastic_localonly_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_meshtastic_localonly_proto_goTypes = []any{
	(*LocalConfig)(nil),                             // 0: meshtastic.LocalConfig
	(*LocalModuleConfig)(nil),                       // 1: meshtastic.LocalModuleConfig
	(*Config_DeviceConfig)(nil),                     // 2: meshtastic.Config.DeviceConfig
	(*Config_PositionConfig)(nil),                   // 3: meshtastic.Config.PositionConfig
	(*Config_PowerConfig)(nil),                      // 4: meshtastic.Config.PowerConfig
	(*Config_NetworkConfig)(nil),                    // 5: meshtastic.Config.NetworkConfig
	(*Config_DisplayConfig)(nil),                    // 6: meshtastic.Config.DisplayConfig
	(*Config_LoRaConfig)(nil),                       // 7: meshtastic.Config.LoRaConfig
	(*Config_BluetoothConfig)(nil),                  // 8: meshtastic.Config.BluetoothConfig
	(*Config_SecurityConfig)(nil),                   // 9: meshtastic.Config.SecurityConfig
	(*ModuleConfig_MQTTConfig)(nil),                 // 10: meshtastic.ModuleConfig.MQTTConfig
	(*ModuleConfig_SerialConfig)(nil),               // 11: meshtastic.ModuleConfig.SerialConfig
	(*ModuleConfig_ExternalNotificationConfig)(nil), // 12: meshtastic.ModuleConfig.ExternalNotificationConfig
	(*ModuleConfig_StoreForwardConfig)(nil),         // 13: meshtastic.ModuleConfig.StoreForwardConfig
	(*ModuleConfig_RangeTestConfig)(nil),            // 14: meshtastic.ModuleConfig.RangeTestConfig
	(*ModuleConfig_TelemetryConfig)(nil),            // 15: meshtastic.ModuleConfig.TelemetryConfig
	(*ModuleConfig_CannedMessageConfig)(nil),        // 16: meshtastic.ModuleConfig.CannedMessageConfig
	(*ModuleConfig_AudioConfig)(nil),                // 17: meshtastic.ModuleConfig.AudioConfig
	(*ModuleConfig_RemoteHardwareConfig)(nil),       // 18: meshtastic.ModuleConfig.RemoteHardwareConfig
	(*ModuleConfig_NeighborInfoConfig)(nil),         // 19: meshtastic.ModuleConfig.NeighborInfoConfig
	(*ModuleConfig_AmbientLightingConfig)(nil),      // 20: meshtastic.ModuleConfig.AmbientLightingConfig
	(*ModuleConfig_DetectionSensorConfig)(nil),      // 21: meshtastic.ModuleConfig.DetectionSensorConfig
	(*ModuleConfig_PaxcounterConfig)(nil),           // 22: meshtastic.ModuleConfig.PaxcounterConfig
}
var file_meshtastic_localonly_proto_depIdxs = []int32{
	2,  // 0: meshtastic.LocalConfig.device:type_name -> meshtastic.Config.DeviceConfig
	3,  // 1: meshtastic.LocalConfig.position:type_name -> meshtastic.Config.PositionConfig
	4,  // 2: meshtastic.LocalConfig.power:type_name -> meshtastic.Config.PowerConfig
	5,  // 3: meshtastic.LocalConfig.network:type_name -> meshtastic.Config.NetworkConfig
	6,  // 4: meshtastic.LocalConfig.display:type_name -> meshtastic.Config.DisplayConfig
	7,  // 5: meshtastic.LocalConfig.lora:type_name -> meshtastic.Config.LoRaConfig
	8,  // 6: meshtastic.LocalConfig.bluetooth:type_name -> meshtastic.Config.BluetoothConfig
	9,  // 7: meshtastic.LocalConfig.security:type_name -> meshtastic.Config.SecurityConfig
	10, // 8: meshtastic.LocalModuleConfig.mqtt:type_name -> meshtastic.ModuleConfig.MQTTConfig
	11, // 9: meshtastic.LocalModuleConfig.serial:type_name -> meshtastic.ModuleConfig.SerialConfig
	12, // 10: meshtastic.LocalModuleConfig.external_notification:type_name -> meshtastic.ModuleConfig.ExternalNotificationConfig
	13, // 11: meshtastic.LocalModuleConfig.store_forward:type_name -> meshtastic.ModuleConfig.StoreForwardConfig
	14, // 12: meshtastic.LocalModuleConfig.range_test:type_name -> meshtastic.ModuleConfig.RangeTestConfig
	15, // 13: meshtastic.LocalModuleConfig.telemetry:type_name -> meshtastic.ModuleConfig.TelemetryConfig
	16, // 14: meshtastic.LocalModuleConfig.canned_message:type_name -> meshtastic.ModuleConfig.CannedMessageConfig
	17, // 15: meshtastic.LocalModuleConfig.audio:type_name -> meshtastic.ModuleConfig.AudioConfig
	18, // 16: meshtastic.LocalModuleConfig.remote_hardware:type_name -> meshtastic.ModuleConfig.RemoteHardwareConfig
	19, // 17: meshtastic.LocalModuleConfig.neighbor_info:type_name -> meshtastic.ModuleConfig.NeighborInfoConfig
	20, // 18: meshtastic.LocalModuleConfig.ambient_lighting:type_name -> meshtastic.ModuleConfig.AmbientLightingConfig
	21, // 19: meshtastic.LocalModuleConfig.detection_sensor:type_name -> meshtastic.ModuleConfig.DetectionSensorConfig
	22, // 20: meshtastic.LocalModuleConfig.paxcounter:type_name -> meshtastic.ModuleConfig.PaxcounterConfig
	21, // [21:21] is the sub-list for method output_type
	21, // [21:21] is the sub-list for method input_type
	21, // [21:21] is the sub-list for extension type_name
	21, // [21:21] is the sub-list for extension extendee
	0,  // [0:21] is the sub-list for field type_name
}

func init() { file_meshtastic_localonly_proto_init() }
func file_meshtastic_localonly_proto_init() {
	if File_meshtastic_localonly_proto != nil {
		return
	}
	file_meshtastic_config_proto_init()
	file_meshtastic_module_config_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_meshtastic_localonly_proto_rawDesc), len(file_meshtastic_localonly_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_meshtastic_localonly_proto_goTypes,
		DependencyIndexes: file_meshtastic_localonly_proto_depIdxs,
		MessageInfos:      file_meshtastic_localonly_proto_msgTypes,
	}.Build()
	File_meshtastic_localonly_proto = out.File
	file_meshtastic_localonly_proto_goTypes = nil
	file_meshtastic_localonly_proto_depIdxs = nil
}
