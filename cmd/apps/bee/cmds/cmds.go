package cmds

import (
	"github.com/go-go-golems/glazed/pkg/settings"
)

// Common function to create Glazed parameter layer
func createGlazedParameterLayer() (*settings.GlazedParameterLayers, error) {
	return settings.NewGlazedParameterLayers()
}
