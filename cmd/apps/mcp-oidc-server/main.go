package main

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	appserver "github.com/go-go-golems/go-go-labs/cmd/apps/mcp-oidc-server/pkg/server"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
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
		dbPath    string
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
			if dbPath != "" {
				if err := s.EnableSQLite(dbPath); err != nil {
					zlog.Fatal().Err(err).Str("db", dbPath).Msg("failed enabling sqlite persistence")
				}
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
	rootCmd.Flags().StringVar(&dbPath, "db", getenv("DB", ""), "SQLite DB path for client persistence (optional)")

    listCmd := &cobra.Command{
        Use:   "list-clients",
        Short: "List registered OAuth clients from SQLite",
        RunE: func(cmd *cobra.Command, args []string) error {
            if dbPath == "" {
                zlog.Fatal().Msg("--db is required to list clients")
            }
            db, err := sql.Open("sqlite3", dbPath)
            if err != nil { return err }
            defer db.Close()
            rows, err := db.Query("SELECT client_id, redirect_uris FROM oauth_clients")
            if err != nil { return err }
            defer rows.Close()
            for rows.Next() {
                var id, uris string
                if err := rows.Scan(&id, &uris); err != nil { return err }
                zlog.Info().Str("client_id", id).Str("redirect_uris", uris).Msg("client")
            }
            return nil
        },
    }
    rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		zlog.Fatal().Err(err).Msg("command error")
	}
}
