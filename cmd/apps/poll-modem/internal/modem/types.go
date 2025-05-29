package modem

import "time"

// ModemInfo represents the complete modem information
type ModemInfo struct {
	CableModem     CableModem     `json:"cable_modem"`
	Downstream     []Channel      `json:"downstream"`
	Upstream       []Channel      `json:"upstream"`
	ErrorCodewords []ErrorChannel `json:"error_codewords"`
	LastUpdated    time.Time      `json:"last_updated"`
}

// CableModem represents the cable modem hardware information
type CableModem struct {
	HWVersion       string `json:"hw_version"`
	Vendor          string `json:"vendor"`
	BOOTVersion     string `json:"boot_version"`
	CoreVersion     string `json:"core_version"`
	Model           string `json:"model"`
	ProductType     string `json:"product_type"`
	FlashPart       string `json:"flash_part"`
	DownloadVersion string `json:"download_version"`
}

// Channel represents a downstream or upstream channel
type Channel struct {
	ChannelID    string `json:"channel_id"`
	LockStatus   string `json:"lock_status"`
	Frequency    string `json:"frequency"`
	SNR          string `json:"snr,omitempty"`          // Only for downstream
	PowerLevel   string `json:"power_level"`
	Modulation   string `json:"modulation"`
	SymbolRate   string `json:"symbol_rate,omitempty"`  // Only for upstream
	ChannelType  string `json:"channel_type,omitempty"` // Only for upstream
}

// ErrorChannel represents error codeword information for a channel
type ErrorChannel struct {
	ChannelID              string `json:"channel_id"`
	UnerroredCodewords     string `json:"unerrored_codewords"`
	CorrectableCodewords   string `json:"correctable_codewords"`
	UncorrectableCodewords string `json:"uncorrectable_codewords"`
} 