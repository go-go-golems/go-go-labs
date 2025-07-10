package meshbus

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ZerologAdapter adapts zerolog to watermill logger
type ZerologAdapter struct {
	logger zerolog.Logger
}

// NewZerologAdapter creates a new zerolog adapter
func NewZerologAdapter() *ZerologAdapter {
	return &ZerologAdapter{
		logger: log.With().Str("component", "watermill").Logger(),
	}
}

// Error logs an error message
func (z *ZerologAdapter) Error(msg string, err error, fields watermill.LogFields) {
	event := z.logger.Error().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Info logs an info message
func (z *ZerologAdapter) Info(msg string, fields watermill.LogFields) {
	event := z.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Debug logs a debug message
func (z *ZerologAdapter) Debug(msg string, fields watermill.LogFields) {
	event := z.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Trace logs a trace message
func (z *ZerologAdapter) Trace(msg string, fields watermill.LogFields) {
	event := z.logger.Trace()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// With creates a new logger with additional fields
func (z *ZerologAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	event := z.logger.With()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	return &ZerologAdapter{logger: event.Logger()}
}

// NewCorrelationMiddleware creates middleware for correlation IDs
func NewCorrelationMiddleware() message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			// Add correlation ID if not present
			if msg.Metadata.Get("correlation_id") == "" {
				msg.Metadata.Set("correlation_id", uuid.New().String())
			}

			// Add processing timestamp
			msg.Metadata.Set("processed_at", time.Now().UTC().Format(time.RFC3339))

			return h(msg)
		}
	}
}

// NewLoggingMiddleware creates middleware for logging
func NewLoggingMiddleware() message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			start := time.Now()

			correlationID := msg.Metadata.Get("correlation_id")
			eventType := msg.Metadata.Get("event_type")
			deviceID := msg.Metadata.Get("device_id")

			logger := log.With().
				Str("correlation_id", correlationID).
				Str("event_type", eventType).
				Str("device_id", deviceID).
				Str("message_uuid", msg.UUID).
				Logger()

			logger.Debug().
				Str("payload_size", formatBytes(len(msg.Payload))).
				Msg("Processing message")

			// Process the message
			results, err := h(msg)

			duration := time.Since(start)

			if err != nil {
				logger.Error().
					Err(err).
					Dur("duration", duration).
					Msg("Message processing failed")
			} else {
				logger.Debug().
					Dur("duration", duration).
					Int("results_count", len(results)).
					Msg("Message processed successfully")
			}

			return results, err
		}
	}
}

// NewRecoveryMiddleware creates middleware for panic recovery
func NewRecoveryMiddleware() message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (results []*message.Message, err error) {
			defer func() {
				if r := recover(); r != nil {
					correlationID := msg.Metadata.Get("correlation_id")
					eventType := msg.Metadata.Get("event_type")

					log.Error().
						Str("correlation_id", correlationID).
						Str("event_type", eventType).
						Str("message_uuid", msg.UUID).
						Interface("panic", r).
						Msg("Panic recovered in message handler")

					err = errors.Errorf("panic recovered: %v", r)
					results = nil
				}
			}()

			return h(msg)
		}
	}
}

// NewRetryMiddleware creates middleware for retrying failed messages
func NewRetryMiddleware(maxRetries int, backoffDuration time.Duration) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			var err error
			var results []*message.Message

			for attempt := 0; attempt <= maxRetries; attempt++ {
				results, err = h(msg)
				if err == nil {
					return results, nil
				}

				if attempt < maxRetries {
					correlationID := msg.Metadata.Get("correlation_id")
					log.Warn().
						Str("correlation_id", correlationID).
						Err(err).
						Int("attempt", attempt+1).
						Int("max_retries", maxRetries).
						Msg("Message processing failed, retrying")

					// Add retry metadata
					msg.Metadata.Set("retry_count", "1")   // Simplified
					msg.Metadata.Set("retry_attempt", "1") // Simplified

					// Wait before retry
					time.Sleep(backoffDuration * time.Duration(attempt+1))
				}
			}

			return results, errors.Wrapf(err, "failed after %d retries", maxRetries)
		}
	}
}

// NewTimeoutMiddleware creates middleware for message processing timeout
func NewTimeoutMiddleware(timeout time.Duration) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			resultsChan := make(chan []*message.Message, 1)
			errorChan := make(chan error, 1)

			go func() {
				results, err := h(msg)
				if err != nil {
					errorChan <- err
				} else {
					resultsChan <- results
				}
			}()

			select {
			case results := <-resultsChan:
				return results, nil
			case err := <-errorChan:
				return nil, err
			case <-ctx.Done():
				correlationID := msg.Metadata.Get("correlation_id")
				log.Warn().
					Str("correlation_id", correlationID).
					Dur("timeout", timeout).
					Msg("Message processing timed out")
				return nil, errors.New("message processing timed out")
			}
		}
	}
}

// formatBytes formats byte count as human readable string
func formatBytes(bytes int) string {
	if bytes < 1024 {
		return "< 1KB"
	} else if bytes < 1024*1024 {
		return "< 1MB"
	}
	return "> 1MB"
}
