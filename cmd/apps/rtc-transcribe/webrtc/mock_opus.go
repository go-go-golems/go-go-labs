//go:build noopus
// +build noopus

package webrtc

import (
	"github.com/pkg/errors"
)

// This file is only compiled when the noopus build tag is set
// It provides a mock implementation by making newRealOpusDecoder return an error

func init() {
	// Override the newRealOpusDecoder variable with our mock implementation
	newRealOpusDecoder = func(sampleRate int, channels int) (OpusDecoder, error) {
		return nil, errors.New("real opus decoder not available (built with noopus tag)")
	}
}