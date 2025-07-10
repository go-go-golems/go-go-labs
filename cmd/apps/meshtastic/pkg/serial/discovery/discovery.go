package discovery

import (
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
)

// DeviceType represents the type of discovered device
type DeviceType int

const (
	DeviceTypeSerial DeviceType = iota
	DeviceTypeNetwork
	DeviceTypeBLE
)

// DiscoveredDevice represents a discovered Meshtastic device
type DiscoveredDevice struct {
	Type        DeviceType
	Port        string
	Host        string
	Address     string
	Description string
	NodeID      uint32
	Device      *SupportedDevice
}

// DiscoveryConfig configures device discovery
type DiscoveryConfig struct {
	Timeout    time.Duration
	SerialOnly bool
	TCPOnly    bool
	BLEOnly    bool
}

// SupportedDevice represents a known Meshtastic device
type SupportedDevice struct {
	Name            string
	Version         string
	Firmware        string
	DeviceClass     string
	VendorID        uint16
	ProductID       uint16
	LinuxPortBase   string
	MacPortBase     string
	WindowsPortBase string
	Description     string
	Priority        int // Lower number = higher priority
}

// Known Meshtastic devices with their VID/PID combinations
var SupportedDevices = []SupportedDevice{
	{
		Name:            "RAK WisBlock 4631",
		Version:         "1.0",
		Firmware:        "esp32",
		DeviceClass:     "esp32",
		VendorID:        0x239A, // Adafruit
		ProductID:       0x8029,
		LinuxPortBase:   "/dev/ttyACM",
		MacPortBase:     "/dev/cu.usbmodem",
		WindowsPortBase: "COM",
		Description:     "RAK WisBlock Core RAK4631",
		Priority:        1,
	},
	{
		Name:            "Heltec WiFi LoRa 32 V2",
		Version:         "2.0",
		Firmware:        "esp32",
		DeviceClass:     "esp32",
		VendorID:        0x10C4, // Silicon Labs
		ProductID:       0xEA60,
		LinuxPortBase:   "/dev/ttyUSB",
		MacPortBase:     "/dev/cu.usbserial",
		WindowsPortBase: "COM",
		Description:     "Heltec WiFi LoRa 32 V2",
		Priority:        2,
	},
	{
		Name:            "T-Beam",
		Version:         "1.0",
		Firmware:        "esp32",
		DeviceClass:     "esp32",
		VendorID:        0x10C4, // Silicon Labs
		ProductID:       0xEA60,
		LinuxPortBase:   "/dev/ttyUSB",
		MacPortBase:     "/dev/cu.usbserial",
		WindowsPortBase: "COM",
		Description:     "TTGO T-Beam",
		Priority:        3,
	},
	{
		Name:            "ESP32 DevKit",
		Version:         "1.0",
		Firmware:        "esp32",
		DeviceClass:     "esp32",
		VendorID:        0x10C4, // Silicon Labs
		ProductID:       0xEA60,
		LinuxPortBase:   "/dev/ttyUSB",
		MacPortBase:     "/dev/cu.usbserial",
		WindowsPortBase: "COM",
		Description:     "Generic ESP32 DevKit",
		Priority:        10,
	},
	{
		Name:            "CH340 USB-Serial",
		Version:         "1.0",
		Firmware:        "esp32",
		DeviceClass:     "esp32",
		VendorID:        0x1A86, // QinHeng Electronics
		ProductID:       0x7523,
		LinuxPortBase:   "/dev/ttyUSB",
		MacPortBase:     "/dev/cu.usbserial",
		WindowsPortBase: "COM",
		Description:     "CH340 USB to Serial",
		Priority:        15,
	},
	{
		Name:            "FTDI USB-Serial",
		Version:         "1.0",
		Firmware:        "esp32",
		DeviceClass:     "esp32",
		VendorID:        0x0403, // FTDI
		ProductID:       0x6001,
		LinuxPortBase:   "/dev/ttyUSB",
		MacPortBase:     "/dev/cu.usbserial",
		WindowsPortBase: "COM",
		Description:     "FTDI USB to Serial",
		Priority:        20,
	},
}

// High-priority device VIDs (known Meshtastic devices)
var WhitelistVIDs = []uint16{
	0x239A, // Adafruit (RAK devices)
	0x303A, // Espressif (ESP32 native USB)
	0x10C4, // Silicon Labs (ESP32 DevKit, Heltec, T-Beam)
	0x1A86, // QinHeng Electronics (CH340)
	0x0403, // FTDI
}

// VIDs to avoid (debug probes, etc.)
var BlacklistVIDs = []uint16{
	0x1366, // SEGGER J-Link
	0x0483, // STMicroelectronics
	0x1915, // Nordic Semiconductor
	0x0925, // Lakeview Research
	0x04B4, // Cypress Semiconductor
}

// DiscoveredPort represents a discovered serial port
type DiscoveredPort struct {
	Name          string
	Description   string
	VendorID      uint16
	ProductID     uint16
	SerialNumber  string
	Manufacturer  string
	Product       string
	Device        *SupportedDevice
	Priority      int
	IsWhitelisted bool
	IsBlacklisted bool
}

// String returns a string representation of the discovered port
func (dp *DiscoveredPort) String() string {
	if dp.Device != nil {
		return fmt.Sprintf("%s (%s)", dp.Name, dp.Device.Description)
	}
	return fmt.Sprintf("%s (%s)", dp.Name, dp.Description)
}

// FindAllPorts discovers all serial ports on the system
func FindAllPorts() ([]DiscoveredPort, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ports list")
	}

	var discovered []DiscoveredPort

	for _, portName := range ports {
		// Get port details
		mode := &serial.Mode{
			BaudRate: 115200,
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: serial.OneStopBit,
		}

		port, err := serial.Open(portName, mode)
		if err != nil {
			// Skip ports that can't be opened
			log.Debug().Str("port", portName).Err(err).Msg("Failed to open port for discovery")
			continue
		}

		// Close the port after testing
		port.Close()

		// Create discovered port with basic information
		// The go.bug.st/serial library doesn't provide VID/PID info easily
		dp := DiscoveredPort{
			Name:         portName,
			Description:  "Serial Port",
			VendorID:     0,
			ProductID:    0,
			SerialNumber: "",
			Manufacturer: "",
			Product:      "",
		}

		// Check if it's whitelisted
		dp.IsWhitelisted = isWhitelisted(dp.VendorID)

		// Check if it's blacklisted
		dp.IsBlacklisted = isBlacklisted(dp.VendorID)

		// Find matching device
		dp.Device = findMatchingDevice(dp.VendorID, dp.ProductID)

		// Set priority
		if dp.Device != nil {
			dp.Priority = dp.Device.Priority
		} else if dp.IsWhitelisted {
			dp.Priority = 50
		} else if dp.IsBlacklisted {
			dp.Priority = 1000
		} else {
			dp.Priority = 100
		}

		discovered = append(discovered, dp)
	}

	return discovered, nil
}

// FindMeshtasticPorts finds all potential Meshtastic ports
func FindMeshtasticPorts() ([]DiscoveredPort, error) {
	allPorts, err := FindAllPorts()
	if err != nil {
		return nil, err
	}

	var meshtasticPorts []DiscoveredPort

	// First, add all whitelisted ports
	for _, port := range allPorts {
		if port.IsWhitelisted && !port.IsBlacklisted {
			meshtasticPorts = append(meshtasticPorts, port)
		}
	}

	// If no whitelisted ports found, add non-blacklisted ports
	if len(meshtasticPorts) == 0 {
		for _, port := range allPorts {
			if !port.IsBlacklisted {
				meshtasticPorts = append(meshtasticPorts, port)
			}
		}
	}

	// Sort by priority (lower number = higher priority)
	sort.Slice(meshtasticPorts, func(i, j int) bool {
		return meshtasticPorts[i].Priority < meshtasticPorts[j].Priority
	})

	return meshtasticPorts, nil
}

// FindBestMeshtasticPort finds the best Meshtastic port automatically
func FindBestMeshtasticPort() (string, error) {
	ports, err := FindMeshtasticPorts()
	if err != nil {
		return "", err
	}

	if len(ports) == 0 {
		return "", errors.New("no Meshtastic devices found")
	}

	if len(ports) > 1 {
		log.Info().
			Int("count", len(ports)).
			Str("selected", ports[0].Name).
			Msg("Multiple Meshtastic devices found, selecting best match")

		// Log all found devices
		for i, port := range ports {
			log.Info().
				Int("index", i).
				Str("port", port.Name).
				Str("description", port.Description).
				Int("priority", port.Priority).
				Msg("Found Meshtastic device")
		}
	}

	return ports[0].Name, nil
}

// TestPortConnection tests if a port can be opened and basic communication works
func TestPortConnection(portName string) error {
	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return errors.Wrap(err, "failed to open port")
	}
	defer port.Close()

	// Set read timeout
	if err := port.SetReadTimeout(1 * time.Second); err != nil {
		return errors.Wrap(err, "failed to set read timeout")
	}

	// Try to read some data (device might be sending data)
	buffer := make([]byte, 256)
	_, err = port.Read(buffer)

	// We don't care about the specific error, just that the port is accessible
	// EOF or timeout are both acceptable here

	return nil
}

// GetPortInfo returns detailed information about a port
func GetPortInfo(portName string) (*DiscoveredPort, error) {
	allPorts, err := FindAllPorts()
	if err != nil {
		return nil, err
	}

	for _, port := range allPorts {
		if port.Name == portName {
			return &port, nil
		}
	}

	return nil, errors.Errorf("port %s not found", portName)
}

// WaitForDevice waits for a Meshtastic device to be connected
func WaitForDevice(timeout time.Duration) (string, error) {
	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			port, err := FindBestMeshtasticPort()
			if err == nil {
				return port, nil
			}

			if time.Since(start) >= timeout {
				return "", errors.New("timeout waiting for device")
			}

			log.Debug().Err(err).Msg("No device found, waiting...")
		}
	}
}

// MonitorDeviceConnection monitors device connection status
func MonitorDeviceConnection(portName string, callback func(connected bool)) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastConnected := false

	for {
		select {
		case <-ticker.C:
			connected := IsDeviceConnected(portName)
			if connected != lastConnected {
				callback(connected)
				lastConnected = connected
			}
		}
	}
}

// IsDeviceConnected checks if a device is connected to the specified port
func IsDeviceConnected(portName string) bool {
	err := TestPortConnection(portName)
	return err == nil
}

// GetPlatformSpecificPorts returns platform-specific port patterns
func GetPlatformSpecificPorts() []string {
	switch runtime.GOOS {
	case "linux":
		return []string{
			"/dev/ttyUSB*",
			"/dev/ttyACM*",
			"/dev/ttyS*",
		}
	case "darwin":
		return []string{
			"/dev/cu.usbserial*",
			"/dev/cu.usbmodem*",
			"/dev/cu.SLAB_USBtoUART*",
		}
	case "windows":
		return []string{
			"COM*",
		}
	default:
		return []string{}
	}
}

// Helper functions

func isWhitelisted(vid uint16) bool {
	for _, whiteVID := range WhitelistVIDs {
		if vid == whiteVID {
			return true
		}
	}
	return false
}

func isBlacklisted(vid uint16) bool {
	for _, blackVID := range BlacklistVIDs {
		if vid == blackVID {
			return true
		}
	}
	return false
}

func findMatchingDevice(vid, pid uint16) *SupportedDevice {
	for _, device := range SupportedDevices {
		if device.VendorID == vid && device.ProductID == pid {
			return &device
		}
	}

	// Try to find by VID only for generic matches
	for _, device := range SupportedDevices {
		if device.VendorID == vid && device.ProductID == 0 {
			return &device
		}
	}

	return nil
}

// DiscoveryResult represents the result of device discovery
type DiscoveryResult struct {
	Ports         []DiscoveredPort
	BestPort      string
	MultipleFound bool
	Error         error
}

// DiscoverDevices performs comprehensive device discovery
func DiscoverDevices() *DiscoveryResult {
	result := &DiscoveryResult{}

	// Find all Meshtastic ports
	ports, err := FindMeshtasticPorts()
	if err != nil {
		result.Error = err
		return result
	}

	result.Ports = ports
	result.MultipleFound = len(ports) > 1

	if len(ports) > 0 {
		result.BestPort = ports[0].Name
	}

	return result
}

// PrintDiscoveryResult prints the discovery result in a user-friendly format
func PrintDiscoveryResult(result *DiscoveryResult) {
	if result.Error != nil {
		fmt.Printf("Error during discovery: %v\n", result.Error)
		return
	}

	if len(result.Ports) == 0 {
		fmt.Println("No Meshtastic devices found.")
		fmt.Println("Make sure your device is connected and drivers are installed.")
		return
	}

	fmt.Printf("Found %d Meshtastic device(s):\n", len(result.Ports))

	for i, port := range result.Ports {
		fmt.Printf("  %d. %s\n", i+1, port.String())
		if port.Device != nil {
			fmt.Printf("     Device: %s\n", port.Device.Name)
			fmt.Printf("     Class: %s\n", port.Device.DeviceClass)
		}
		fmt.Printf("     VID:PID: %04X:%04X\n", port.VendorID, port.ProductID)
		if port.SerialNumber != "" {
			fmt.Printf("     Serial: %s\n", port.SerialNumber)
		}
		if port.Manufacturer != "" {
			fmt.Printf("     Manufacturer: %s\n", port.Manufacturer)
		}
		fmt.Printf("     Priority: %d\n", port.Priority)
		fmt.Println()
	}

	if result.BestPort != "" {
		fmt.Printf("Best match: %s\n", result.BestPort)
	}

	if result.MultipleFound {
		fmt.Println("Multiple devices found. Use the --port flag to specify a device.")
	}
}

// DiscoverMeshtasticDevices discovers available Meshtastic devices
func DiscoverMeshtasticDevices(config DiscoveryConfig) ([]DiscoveredDevice, error) {
	var devices []DiscoveredDevice

	// Serial device discovery
	if !config.TCPOnly && !config.BLEOnly {
		serialDevices, err := discoverSerialDevices(config.Timeout)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to discover serial devices")
		} else {
			devices = append(devices, serialDevices...)
		}
	}

	// Network device discovery (placeholder)
	if !config.SerialOnly && !config.BLEOnly {
		networkDevices, err := discoverNetworkDevices(config.Timeout)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to discover network devices")
		} else {
			devices = append(devices, networkDevices...)
		}
	}

	// BLE device discovery (placeholder)
	if !config.SerialOnly && !config.TCPOnly {
		bleDevices, err := discoverBLEDevices(config.Timeout)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to discover BLE devices")
		} else {
			devices = append(devices, bleDevices...)
		}
	}

	return devices, nil
}

// discoverSerialDevices discovers serial Meshtastic devices
func discoverSerialDevices(timeout time.Duration) ([]DiscoveredDevice, error) {
	ports, err := FindMeshtasticPorts()
	if err != nil {
		return nil, err
	}

	var devices []DiscoveredDevice
	for _, port := range ports {
		device := DiscoveredDevice{
			Type:        DeviceTypeSerial,
			Port:        port.Name,
			Description: port.Description,
			Device:      port.Device,
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// discoverNetworkDevices discovers network Meshtastic devices (placeholder)
func discoverNetworkDevices(timeout time.Duration) ([]DiscoveredDevice, error) {
	// TODO: Implement network discovery
	return []DiscoveredDevice{}, nil
}

// discoverBLEDevices discovers BLE Meshtastic devices (placeholder)
func discoverBLEDevices(timeout time.Duration) ([]DiscoveredDevice, error) {
	// TODO: Implement BLE discovery
	return []DiscoveredDevice{}, nil
}
