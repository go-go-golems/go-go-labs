package pkg

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	gauth "github.com/go-go-golems/go-go-labs/pkg/google/auth"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v3"
)

type FetchSettings struct {
	FormID string `glazed.parameter:"form-id"`
	Output string `glazed.parameter:"output"`
	Debug  bool   `glazed.parameter:"debug"`
}

type FetchCommand struct {
	*cmds.CommandDescription
}

func NewFetchCommand() (*cobra.Command, error) {
	oauthLayers, err := gauth.GetOAuthTokenStoreLayersWithOptions(
		gauth.WithCredentialsDefault("~/.google-form/client_secret.json"),
		gauth.WithTokenDefault("~/.google-form/token.json"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create OAuth token store layers: %w", err)
	}

	desc := cmds.NewCommandDescription(
		"fetch",
		cmds.WithShort("Fetch a Google Form and convert it back to a Wizard DSL"),
		cmds.WithLong(`
Retrieve an existing Google Form by ID and turn it into a Uhoh Wizard DSL file.
`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"form-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Existing Google Form ID to fetch"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"output",
				parameters.ParameterTypeString,
				parameters.WithHelp("Optional path to write the wizard YAML"),
			),
			parameters.NewParameterDefinition(
				"debug",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable debug logging"),
				parameters.WithDefault(false),
			),
		),
		cmds.WithLayers(oauthLayers),
	)

	c := &FetchCommand{CommandDescription: desc}
	return cli.BuildCobraCommandFromBareCommand(c)
}

func (c *FetchCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &FetchSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	authenticator, err := buildFormsAuthenticator(parsedLayers)
	if err != nil {
		return fmt.Errorf("failed to create authenticator: %w", err)
	}
	result, err := authenticator.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	svc, err := forms.NewService(ctx, option.WithTokenSource(result.Client.TokenSource(ctx, result.Token)))
	if err != nil {
		return fmt.Errorf("failed to create forms service: %w", err)
	}

	form, err := svc.Forms.Get(settings.FormID).Do()
	if err != nil {
		return fmt.Errorf("failed to fetch form %s: %w", settings.FormID, err)
	}

	wizard, _, err := ConvertFormToWizard(form)
	if err != nil {
		return fmt.Errorf("failed to convert form to wizard DSL: %w", err)
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(wizard); err != nil {
		return fmt.Errorf("failed to marshal wizard DSL: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("failed to finalize YAML encoding: %w", err)
	}

	if settings.Output != "" {
		if err := os.WriteFile(settings.Output, buf.Bytes(), 0o644); err != nil {
			return fmt.Errorf("failed to write wizard file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Wrote wizard YAML to %s\n", settings.Output)
		return nil
	}

	fmt.Printf("%s", buf.String())
	return nil
}

var _ cmds.BareCommand = &FetchCommand{}
