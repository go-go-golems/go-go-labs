//go:build !noopus
// +build !noopus

package webrtc

import (
	"github.com/hraban/opus"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RealOpusDecoder implements the OpusDecoder interface using libopus
type RealOpusDecoder struct {
	decoder *opus.Decoder
	logger  zerolog.Logger
}

// init is called when the package is initialized
// It overrides the default mock implementation with the real one
func init() {
	// Override the newRealOpusDecoder variable with our real implementation
	newRealOpusDecoder = func(sampleRate int, channels int) (OpusDecoder, error) {
		logger := log.With().
			Str("component", "RealOpusDecoder").
			Int("sampleRate", sampleRate).
			Int("channels", channels).
			Logger()

		logger.Info().Msg("Creating real Opus decoder using libopus")

		decoder, err := opus.NewDecoder(sampleRate, channels)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Opus decoder")
		}

		logger.Info().Msg("Created real Opus decoder")

		return &RealOpusDecoder{
			decoder: decoder,
			logger:  logger,
		}, nil
	}
}

// Decode decodes an Opus packet to PCM data
func (d *RealOpusDecoder) Decode(data []byte, pcm []int16) (int, error) {
	if len(data) == 0 {
		d.logger.Debug().Msg("Empty Opus packet, skipping")
		return 0, nil
	}

	d.logger.Debug().
		Int("opusBytes", len(data)).
		Int("pcmCapacity", len(pcm)).
		Msg("Decoding Opus packet with real decoder")

	n, err := d.decoder.Decode(data, pcm)
	if err != nil {
		return 0, errors.Wrap(err, "failed to decode Opus packet")
	}

	d.logger.Debug().
		Int("decodedSamples", n).
		Msg("Successfully decoded Opus packet")

	return n, nil
}

// Close releases resources used by the decoder
func (d *RealOpusDecoder) Close() error {
	d.logger.Debug().Msg("Closing real Opus decoder")
	return nil
}
