package webrtc

import (
	"github.com/pkg/errors"
)

// This will be overridden by the real implementation if the build tags match
var newRealOpusDecoder = func(sampleRate int, channels int) (OpusDecoder, error) {
	return nil, errors.New("no opus implementation available")
}
