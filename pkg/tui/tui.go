// Package tui provides Terminal User Interface components for the Redis monitor
package tui

// Re-export types for convenience
import (
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

// Re-export commonly used types
type (
	StreamData   = models.StreamData
	GroupData    = models.GroupData
	ConsumerData = models.ConsumerData
	ServerData   = models.ServerData
	Styles       = styles.Styles
)

// NewStyles creates default TUI styles
func NewStyles() Styles {
	return styles.NewStyles()
}
