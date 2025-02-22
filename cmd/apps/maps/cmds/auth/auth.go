package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/go-go-labs/pkg/google/auth"
	"github.com/go-go-golems/go-go-labs/pkg/google/auth/store"
	"github.com/spf13/cobra"
)

type LoginCommand struct {
	*cmds.CommandDescription
}

func NewLoginCommand() (*cobra.Command, error) {
	layers_, err := auth.GetOAuthTokenStoreLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create OAuth token store layers: %w", err)
	}

	cmd := &LoginCommand{
		CommandDescription: cmds.NewCommandDescription(
			"login",
			cmds.WithShort("Login to Google Maps"),
			cmds.WithLong("Authenticate with Google Maps using OAuth2"),
			cmds.WithLayers(layers_),
		),
	}

	return cli.BuildCobraCommandFromBareCommand(cmd)
}

func (c *LoginCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	authenticator, err := auth.CreateAuthenticatorFromLayers(parsedLayers)
	if err != nil {
		return fmt.Errorf("failed to create authenticator: %w", err)
	}

	result, err := authenticator.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	fmt.Printf("Successfully logged in (token expires at %s)\n", result.Token.Expiry.Format("2006-01-02 15:04:05"))
	return nil
}

type LogoutCommand struct {
	*cmds.CommandDescription
}

func NewLogoutCommand() (*cobra.Command, error) {
	layers_, err := auth.GetOAuthTokenStoreLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create OAuth token store layers: %w", err)
	}

	cmd := &LogoutCommand{
		CommandDescription: cmds.NewCommandDescription(
			"logout",
			cmds.WithShort("Logout from Google Maps"),
			cmds.WithLong("Remove stored Google Maps authentication credentials"),
			cmds.WithLayers(layers_),
		),
	}

	return cli.BuildCobraCommandFromBareCommand(cmd)
}

func (c *LogoutCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	tokenStore, err := auth.CreateTokenStoreFromLayers(parsedLayers)
	if err != nil {
		return fmt.Errorf("failed to create token store: %w", err)
	}

	err = tokenStore.Clear(ctx)
	if err != nil && !errors.Is(err, store.ErrTokenNotFound) {
		return fmt.Errorf("failed to clear token: %w", err)
	}

	fmt.Println("Successfully logged out")
	return nil
}

type StatusCommand struct {
	*cmds.CommandDescription
}

func NewStatusCommand() (*cobra.Command, error) {
	layers_, err := auth.GetOAuthTokenStoreLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create OAuth token store layers: %w", err)
	}

	cmd := &StatusCommand{
		CommandDescription: cmds.NewCommandDescription(
			"status",
			cmds.WithShort("Check Google Maps authentication status"),
			cmds.WithLong("Check if you are currently authenticated with Google Maps"),
			cmds.WithLayers(layers_),
		),
	}

	return cli.BuildCobraCommandFromBareCommand(cmd)
}

func (c *StatusCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	tokenStore, err := auth.CreateTokenStoreFromLayers(parsedLayers)
	if err != nil {
		return fmt.Errorf("failed to create token store: %w", err)
	}

	token, err := tokenStore.Load(ctx)
	if err != nil {
		if errors.Is(err, store.ErrTokenNotFound) {
			fmt.Println("Not logged in")
			return nil
		}
		return fmt.Errorf("failed to load token: %w", err)
	}

	fmt.Printf("Logged in (token expires at %s)\n", token.Expiry.Format("2006-01-02 15:04:05"))
	return nil
}

func NewAuthCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Google Maps authentication",
		Long:  "Login, logout, and check authentication status for Google Maps",
	}

	loginCmd, err := NewLoginCommand()
	if err != nil {
		return nil, err
	}

	logoutCmd, err := NewLogoutCommand()
	if err != nil {
		return nil, err
	}

	statusCmd, err := NewStatusCommand()
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(loginCmd, logoutCmd, statusCmd)
	return cmd, nil
}
