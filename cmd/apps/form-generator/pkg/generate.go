package pkg

import (
    "context"
    "fmt"

    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    gauth "github.com/go-go-golems/go-go-labs/pkg/google/auth"
    uhwizard "github.com/go-go-golems/uhoh/pkg/wizard"
    "github.com/rs/zerolog"
    // "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
    "google.golang.org/api/forms/v1"
    "google.golang.org/api/option"
)

// GenerateSettings holds command parameters.
type GenerateSettings struct {
    WizardFile   string `glazed.parameter:"wizard"`
    FormID       string `glazed.parameter:"form-id"`
    Title        string `glazed.parameter:"title"`
    Description  string `glazed.parameter:"description"`
    Debug        bool   `glazed.parameter:"debug"`
}

type GenerateCommand struct {
    *cmds.CommandDescription
}

// NewGenerateCommand creates the Cobra command for generating a Google Form from a Uhoh wizard.
func NewGenerateCommand() (*cobra.Command, error) {
    // OAuth layers for credentials and token storage
    oauthLayers, err := gauth.GetOAuthTokenStoreLayers()
    if err != nil {
        return nil, fmt.Errorf("could not create OAuth token store layers: %w", err)
    }

    desc := cmds.NewCommandDescription(
        "generate",
        cmds.WithShort("Generate or update a Google Form from a Uhoh Wizard DSL file"),
        cmds.WithLong(`
Generate Google Forms items from a Uhoh Wizard DSL file.

If --form-id is provided, the command reuses the existing form and appends the generated items.
You can optionally override the title and description.
`),
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "wizard",
                parameters.ParameterTypeFile,
                parameters.WithHelp("Path to Uhoh wizard YAML file"),
                parameters.WithRequired(true),
            ),
            parameters.NewParameterDefinition(
                "form-id",
                parameters.ParameterTypeString,
                parameters.WithHelp("Existing Google Form ID to reuse (append items)"),
            ),
            parameters.NewParameterDefinition(
                "title",
                parameters.ParameterTypeString,
                parameters.WithHelp("Override form title (also applied to existing form if provided)"),
            ),
            parameters.NewParameterDefinition(
                "description",
                parameters.ParameterTypeString,
                parameters.WithHelp("Override form description"),
            ),
            parameters.NewParameterDefinition(
                "debug",
                parameters.ParameterTypeBool,
                parameters.WithDefault(false),
                parameters.WithHelp("Enable debug logging"),
            ),
        ),
        cmds.WithLayers(oauthLayers),
    )

    c := &GenerateCommand{CommandDescription: desc}
    return cli.BuildCobraCommandFromBareCommand(c)
}

// buildFormsAuthenticator constructs an authenticator for Google Forms with proper scopes from layers.
func buildFormsAuthenticator(parsedLayers *layers.ParsedLayers) (*gauth.Authenticator, error) {
    // Parse auth settings
    s := &gauth.AuthSettings{}
    if err := parsedLayers.InitializeStruct(gauth.AuthSlug, s); err != nil {
        return nil, fmt.Errorf("failed to initialize auth settings: %w", err)
    }

    // Build options (like CreateOptionsFromSettings but with Forms scope)
    opts := []gauth.Option{}
    // credentials file + server mode + timeout come from settings helpers
    o, err := gauth.CreateOptionsFromSettings(s)
    if err != nil {
        return nil, err
    }
    // Replace default scopes by appending Forms scope and ignoring defaults.
    // CreateOptionsFromSettings already appended default scopes; to ensure Forms scope,
    // add it again (OAuth allows multiple). Forms requires forms.FormsBodyScope.
    opts = append(opts, o...)
    opts = append(opts, gauth.WithScopes(forms.FormsBodyScope))

    tokenStore, err := gauth.CreateTokenStoreFromLayers(parsedLayers)
    if err != nil {
        return nil, fmt.Errorf("failed to create token store: %w", err)
    }

    opts = append(opts, gauth.WithTokenStore(tokenStore))
    return gauth.NewAuthenticator(opts...)
}

// Run implements cmds.BareCommand.
func (c *GenerateCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    settings := &GenerateSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return err
    }

    if settings.Debug {
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    } else {
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    }

    // Load wizard DSL
    wz, err := uhwizard.LoadWizard(settings.WizardFile)
    if err != nil {
        return fmt.Errorf("failed to load wizard DSL: %w", err)
    }

    // Determine form title/description defaults
    formTitle := wz.Name
    if settings.Title != "" {
        formTitle = settings.Title
    }
    if formTitle == "" {
        formTitle = "Generated Form"
    }
    formDescription := wz.Description
    if settings.Description != "" {
        formDescription = settings.Description
    }

    // Authenticate using Google OAuth with Forms scope
    authenticator, err := buildFormsAuthenticator(parsedLayers)
    if err != nil {
        return fmt.Errorf("failed to create authenticator: %w", err)
    }
    result, err := authenticator.Authenticate(ctx)
    if err != nil {
        return fmt.Errorf("authentication failed: %w", err)
    }

    // Create Forms service
    ts := result.Client.TokenSource(ctx, result.Token)
    svc, err := forms.NewService(ctx, option.WithTokenSource(ts))
    if err != nil {
        return fmt.Errorf("failed to create forms service: %w", err)
    }

    // Build requests from wizard
    requests, err := BuildRequestsFromWizard(wz)
    if err != nil {
        return fmt.Errorf("failed to build form items: %w", err)
    }

    // Create or update the form
    form, err := CreateOrUpdateForm(ctx, svc, settings.FormID, formTitle, formDescription, requests)
    if err != nil {
        return err
    }

    // Output result in a simple human-readable way
    if settings.FormID == "" {
        fmt.Printf("Created form: %s\n", form.FormId)
    } else {
        fmt.Printf("Updated form: %s\n", form.FormId)
    }
    if form.ResponderUri != "" {
        fmt.Printf("Fill-in link: %s\n", form.ResponderUri)
    }

    return nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &GenerateCommand{}

