package main

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	appserver "github.com/go-go-golems/go-go-labs/cmd/apps/mcp-oidc-server/pkg/server"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	var (
		addr      string
		issuer    string
		logFormat string
		logLevel  string
	)

	rootCmd := &cobra.Command{
		Use:   "mcp-oidc-server",
		Short: "MCP + OIDC discovery stub server",
		RunE: func(cmd *cobra.Command, args []string) error {
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
			switch logFormat {
			case "json":
				// default JSON
			default:
				zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
			}
			switch logLevel {
			case "trace":
				zerolog.SetGlobalLevel(zerolog.TraceLevel)
			case "debug":
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			case "info":
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			case "warn":
				zerolog.SetGlobalLevel(zerolog.WarnLevel)
			case "error":
				zerolog.SetGlobalLevel(zerolog.ErrorLevel)
			default:
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}

			s, err := appserver.New(issuer)
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed generating RSA key")
			}

			mux := http.NewServeMux()
			s.Routes(mux)
			wrapped := s.LoggingMiddleware(mux)
			zlog.Info().Str("addr", addr).Str("issuer", issuer).Msg("mcp-oidc-server listening")
			if err := http.ListenAndServe(addr, wrapped); err != nil {
				zlog.Fatal().Err(err).Msg("server exited")
			}
			return nil
		},
	}

	rootCmd.Flags().StringVar(&addr, "addr", getenv("ADDR", ":8080"), "HTTP listen address")
	rootCmd.Flags().StringVar(&issuer, "issuer", getenv("ISSUER", "http://localhost:8080"), "Issuer/base URL")
	rootCmd.Flags().StringVar(&logFormat, "log-format", getenv("LOG_FORMAT", "console"), "Log format: console|json")
	rootCmd.Flags().StringVar(&logLevel, "log-level", getenv("LOG_LEVEL", "info"), "Log level: trace|debug|info|warn|error")

	if err := rootCmd.Execute(); err != nil {
		zlog.Fatal().Err(err).Msg("command error")
	}
}
