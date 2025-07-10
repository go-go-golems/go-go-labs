package serial

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tarm/serial"
)

// Common VID/PID combinations for Meshtastic devices
var (
	// Whitelisted vendor IDs (these are likely Meshtastic devices)
	WhitelistedVIDs = []uint16{
		0x239a, // Adafruit (RAK4631)
		0x303a, // Espressif (ESP32-based devices)
		0x10c4, // Silicon Labs (CP2102)
		0x0403, // FTDI
	}

	// Blacklisted vendor IDs (these are definitely not Meshtastic devices)
	BlacklistedVIDs = []uint16{
		0x1366, // Segger J-Link
		0x0483, // STMicroelectronics
		0x1915, // Nordic Semiconductor
		0x0925, // Lakeview Research
		0x04b4, // Cypress Semiconductor
	}

	// Common device paths for different platforms
	DefaultDevicePaths = []string{
		"/dev/ttyACM0",
		"/dev/ttyACM1",
		"/dev/ttyUSB0",
		"/dev/ttyUSB1",
		"/dev/cu.usbmodem*",
		"/dev/cu.usbserial*",
	}
)

// DeviceInfo represents information about a discovered device
type DeviceInfo struct {
	Path        string
	VID         uint16
	PID         uint16
	SerialNum   string
	Description string
}

// String returns a string representation of the device info
func (d DeviceInfo) String() string {
	return fmt.Sprintf("%s (VID:0x%04x PID:0x%04x Serial:%s) - %s",
		d.Path, d.VID, d.PID, d.SerialNum, d.Description)
}

// FindMeshtasticPorts discovers potential Meshtastic devices
func FindMeshtasticPorts() ([]DeviceInfo, error) {
	log.Debug().Msg("Starting device discovery")

	// Get list of available serial ports
	ports, err := listSerialPorts()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list serial ports")
	}

	if len(ports) == 0 {
		log.Debug().Msg("No serial ports found")
		return nil, nil
	}

	log.Debug().Int("count", len(ports)).Msg("Found serial ports")

	var devices []DeviceInfo
	var whitelistedDevices []DeviceInfo

	// Filter ports based on VID/PID
	for _, port := range ports {
		device := DeviceInfo{
			Path:        port,
			Description: "Unknown device",
		}

		// Try to get device info (this is platform-specific)
		if info, err := getDeviceInfo(port); err == nil {
			device.VID = info.VID
			device.PID = info.PID
			device.SerialNum = info.SerialNum
			device.Description = info.Description
		}

		// Check if device is whitelisted
		if isWhitelisted(device.VID) {
			log.Debug().Str("port", port).Uint16("vid", device.VID).Msg("Found whitelisted device")
			whitelistedDevices = append(whitelistedDevices, device)
		} else if !isBlacklisted(device.VID) {
			log.Debug().Str("port", port).Uint16("vid", device.VID).Msg("Found non-blacklisted device")
			devices = append(devices, device)
		} else {
			log.Debug().Str("port", port).Uint16("vid", device.VID).Msg("Skipping blacklisted device")
		}
	}

	// Prefer whitelisted devices
	if len(whitelistedDevices) > 0 {
		log.Info().Int("count", len(whitelistedDevices)).Msg("Using whitelisted devices")
		return whitelistedDevices, nil
	}

	if len(devices) > 0 {
		log.Info().Int("count", len(devices)).Msg("Using non-blacklisted devices")
		return devices, nil
	}

	log.Debug().Msg("No suitable devices found")
	return nil, nil
}

// FindBestMeshtasticPort returns the best candidate for a Meshtastic device
func FindBestMeshtasticPort() (string, error) {
	devices, err := FindMeshtasticPorts()
	if err != nil {
		return "", err
	}

	if len(devices) == 0 {
		return "", errors.New("no Meshtastic devices found")
	}

	// Return the first device (they're already sorted by preference)
	return devices[0].Path, nil
}

// isWhitelisted checks if a VID is in the whitelist
func isWhitelisted(vid uint16) bool {
	for _, whitelistedVID := range WhitelistedVIDs {
		if vid == whitelistedVID {
			return true
		}
	}
	return false
}

// isBlacklisted checks if a VID is in the blacklist
func isBlacklisted(vid uint16) bool {
	for _, blacklistedVID := range BlacklistedVIDs {
		if vid == blacklistedVID {
			return true
		}
	}
	return false
}

// listSerialPorts returns a list of available serial ports
func listSerialPorts() ([]string, error) {
	// For now, use a simple approach - check default paths
	// In a real implementation, you'd want to use platform-specific APIs
	var ports []string

	for _, path := range DefaultDevicePaths {
		if !strings.Contains(path, "*") {
			// Simple path - check if it exists
			if _, err := serial.OpenPort(&serial.Config{Name: path, Baud: 115200}); err == nil {
				ports = append(ports, path)
			}
		}
		// TODO: Handle wildcard paths for macOS
	}

	return ports, nil
}

// DeviceInformation represents device USB information
type DeviceInformation struct {
	VID         uint16
	PID         uint16
	SerialNum   string
	Description string
}

// getDeviceInfo tries to get USB device information for a port
func getDeviceInfo(port string) (*DeviceInformation, error) {
	// This is a simplified implementation
	// In a real implementation, you'd use platform-specific APIs
	// to get USB device information
	return &DeviceInformation{
		VID:         0x0000,
		PID:         0x0000,
		SerialNum:   "unknown",
		Description: "Serial device",
	}, nil
}
