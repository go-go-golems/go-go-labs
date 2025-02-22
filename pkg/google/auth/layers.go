package auth

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

const (
	AuthSlug   = "google-auth"
	DBAuthSlug = "db-auth"
)

type AuthSettings struct {
	CredentialsFile string `glazed.parameter:"credentials-file"`
	TokenStoreType  string `glazed.parameter:"token-store-type"`
	TokenStorePath  string `glazed.parameter:"token-store-path"`
	TokenStorePerms int    `glazed.parameter:"token-store-perms"`
	ServerPort      int    `glazed.parameter:"server-port"`
	CallbackPath    string `glazed.parameter:"callback-path"`
	Timeout         int    `glazed.parameter:"timeout"`
}

type DBAuthSettings struct {
	UserID    int      `glazed.parameter:"user-id"`
	Provider  string   `glazed.parameter:"provider"`
	Scopes    []string `glazed.parameter:"scopes"`
	TeamID    string   `glazed.parameter:"team-id"`
	UserAppID string   `glazed.parameter:"user-app-id"`
	AppID     string   `glazed.parameter:"app-id"`
}

func NewDBAuthParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		DBAuthSlug,
		"Database Auth Settings",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"user-id",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("User ID for database token store"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"provider",
				parameters.ParameterTypeString,
				parameters.WithHelp("OAuth provider name for database token store"),
				parameters.WithDefault("google"),
			),
			parameters.NewParameterDefinition(
				"scopes",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("OAuth scopes for the token"),
			),
			parameters.NewParameterDefinition(
				"team-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Slack team ID for the token"),
			),
			parameters.NewParameterDefinition(
				"user-app-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Slack user ID for the token"),
			),
			parameters.NewParameterDefinition(
				"app-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Slack app installation ID for the token"),
			),
		),
	)
}

func NewAuthParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		AuthSlug,
		"Google Auth Settings",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"credentials-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to Google OAuth2 credentials file"),
				parameters.WithDefault("~/.gcal/credentials.json"),
			),
			parameters.NewParameterDefinition(
				"token-store-type",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Type of token store to use (file or database)"),
				parameters.WithDefault("file"),
				parameters.WithChoices("file", "database"),
			),
			parameters.NewParameterDefinition(
				"token-store-path",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to store the token (for file token store)"),
				parameters.WithDefault("~/.gcal/token.json"),
			),
			parameters.NewParameterDefinition(
				"token-store-perms",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("File permissions for token store (in octal)"),
				parameters.WithDefault(0600),
			),
			parameters.NewParameterDefinition(
				"server-port",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Port for OAuth2 callback server"),
				parameters.WithDefault(8080),
			),
			parameters.NewParameterDefinition(
				"callback-path",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path for OAuth2 callback endpoint"),
				parameters.WithDefault("/callback"),
			),
			parameters.NewParameterDefinition(
				"timeout",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Timeout in minutes for OAuth2 flow"),
				parameters.WithDefault(5),
			),
		),
	)
}
