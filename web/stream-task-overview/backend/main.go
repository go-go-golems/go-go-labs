package main

import (
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Initialize logger
func initLogger() {
	// Pretty console logging for development
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	multi := zerolog.MultiLevelWriter(consoleWriter)

	// Set global logger
	log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
	
	// Set log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Enable debug level in development
	if os.Getenv("ENV") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Msg("Logger initialized")
}

// Custom echo logger adapter for zerolog
type ZerologAdapter struct {
	log zerolog.Logger
}

func (zl ZerologAdapter) Write(p []byte) (n int, err error) {
	zl.log.Info().Msg(string(p))
	return len(p), nil
}

func (zl ZerologAdapter) Output() io.Writer {
	return zl
}

func main() {
	// Initialize logger
	initLogger()

	// Echo setup
	e := echo.New()
	e.Logger.SetOutput(ZerologAdapter{log: log.With().Str("component", "echo").Logger()})
	e.HideBanner = true

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	
	// Custom logger middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			log.Debug().Str("method", req.Method).Str("path", req.URL.Path).Str("id", c.Response().Header().Get(echo.HeaderXRequestID)).Msg("Request")

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			latency := time.Since(start)
			log.Info().Str("method", req.Method).Str("path", req.URL.Path).Int("status", res.Status).Dur("latency", latency).Msg("Response")

			return err
		}
	})

	// Setup store and handlers
	log.Info().Msg("Initializing data store")
	store := NewStreamStore()

	log.Info().Msg("Creating request handlers")
	h := NewStreamHandler(store)

	// Routes
	log.Info().Msg("Setting up API routes")
	e.GET("/api/stream", h.GetStreamInfo)
	e.PUT("/api/stream", h.UpdateStreamInfo)
	e.GET("/api/stream/steps", h.GetSteps)
	e.PUT("/api/stream/steps/active", h.SetActiveStep)
	e.POST("/api/stream/steps/upcoming", h.AddUpcomingStep)
	e.POST("/api/stream/steps/complete", h.CompleteActiveStep)
	e.PUT("/api/stream/steps/reactivate", h.ReactivateStep)

	// Setup graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		s := <-sig
		log.Info().Str("signal", s.String()).Msg("Shutting down server...")
		e.Close()
	}()

	// Start server
	log.Info().Str("address", ":8080").Msg("Starting server")
	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server startup failed")
	}
}