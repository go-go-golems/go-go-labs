package pkg

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	gauth "github.com/go-go-golems/go-go-labs/pkg/google/auth"
	uhwizard "github.com/go-go-golems/uhoh/pkg/wizard"
	uhsteps "github.com/go-go-golems/uhoh/pkg/wizard/steps"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v3"
)

// GenerateSettings holds command parameters.
type GenerateSettings struct {
	WizardFile  *parameters.FileData `glazed.parameter:"wizard"`
	FormID      string                `glazed.parameter:"form-id"`
	Create      bool                  `glazed.parameter:"create"`
	Title       string                `glazed.parameter:"title"`
	Description string                `glazed.parameter:"description"`
	Debug       bool                  `glazed.parameter:"debug"`
}

type GenerateCommand struct {
	*cmds.CommandDescription
}

// NewGenerateCommand creates the Cobra command for generating a Google Form from a Uhoh wizard.
func NewGenerateCommand() (*cobra.Command, error) {
	// OAuth layers for credentials and token storage
	oauthLayers, err := gauth.GetOAuthTokenStoreLayersWithOptions(
		gauth.WithCredentialsDefault("~/.google-form/client_secret.json"),
		gauth.WithTokenDefault("~/.google-form/token.json"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create OAuth token store layers: %w", err)
	}

	desc := cmds.NewCommandDescription(
		"generate",
		cmds.WithShort("Generate or update a Google Form from a Uhoh Wizard DSL file"),
		cmds.WithLong(`
Generate Google Forms items from a Uhoh Wizard DSL file.

Use --create to create a new form. If --form-id is provided, the command updates the existing form
by replacing its items with the ones described in the DSL. You can optionally override the title and description.
`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"wizard",
				parameters.ParameterTypeFile,
				parameters.WithHelp("Path to Uhoh wizard YAML file"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"create",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Create a new form (mutually exclusive with --form-id)"),
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
func buildFormsAuthenticator(parsedLayers *layers.ParsedLayers, extraScopes ...string) (*gauth.Authenticator, error) {
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
	scopes := []string{forms.FormsBodyScope}
	if len(extraScopes) > 0 {
		scopes = append(scopes, extraScopes...)
	}
	opts = append(opts, gauth.WithScopes(scopes...))

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

	if settings.Create && settings.FormID != "" {
		return fmt.Errorf("--create cannot be used together with --form-id; choose one")
	}

	if settings.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Load wizard DSL from FileData
	var wz *uhwizard.Wizard
	var yamlFile string
	
	if settings.WizardFile == nil {
		return fmt.Errorf("wizard file is required")
	}
	
	// FileData contains both path and content
	yamlFile = settings.WizardFile.AbsolutePath
	if yamlFile == "" {
		yamlFile = settings.WizardFile.Path
	}
	
	// Try to load from file path first
	if yamlFile != "" && yamlFile != "stdin" {
		log.Debug().Str("yamlFile", yamlFile).Msg("Loading wizard from file")
		var err error
		wz, err = uhwizard.LoadWizard(yamlFile)
		if err != nil {
			return fmt.Errorf("failed to load wizard DSL from file: %w", err)
		}
	} else {
		// Fallback to content (for stdin or when path is not available)
		log.Debug().Msg("Loading wizard from content")
		var w uhwizard.Wizard
		content := settings.WizardFile.Content
		if content == "" {
			content = settings.WizardFile.StringContent
		}
		if err := yaml.Unmarshal([]byte(content), &w); err != nil {
			return fmt.Errorf("failed to load wizard DSL from content: %w", err)
		}
		wz = &w
		// Note: DecisionStep choices can't be fixed without file path
		yamlFile = ""
	}

	// Fix DecisionStep choices by parsing options from YAML
	if yamlFile != "" {
		log.Debug().Str("yamlFile", yamlFile).Msg("Calling fixDecisionStepChoices")
		if err := fixDecisionStepChoices(wz, yamlFile); err != nil {
			return fmt.Errorf("failed to fix decision step choices: %w", err)
		}
	} else {
		log.Debug().Msg("yamlFile is empty, skipping fixDecisionStepChoices")
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

	// Create Forms service
	authenticator, err := buildFormsAuthenticator(parsedLayers)
	if err != nil {
		return fmt.Errorf("failed to create authenticator: %w", err)
	}
	result, err := authenticator.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

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
	targetFormID := settings.FormID
	if settings.Create {
		targetFormID = ""
	}
	form, err := CreateOrUpdateForm(ctx, svc, targetFormID, formTitle, formDescription, requests)
	if err != nil {
		return err
	}

	// Output result
	if targetFormID == "" {
		fmt.Printf("Created form: %s\n", form.FormId)
	} else {
		fmt.Printf("Updated form: %s\n", form.FormId)
	}
	if form.ResponderUri != "" {
		fmt.Printf("Fill-in link: %s\n", form.ResponderUri)
	}

	return nil
}

// fixDecisionStepChoices extracts options from YAML and populates DecisionStep.Choices
func fixDecisionStepChoices(wz *uhwizard.Wizard, yamlFile string) error {
	log.Debug().Str("yamlFile", yamlFile).Int("steps", len(wz.Steps)).Msg("Fixing DecisionStep choices")
	
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	var rawWizard map[string]interface{}
	if err := yaml.Unmarshal(data, &rawWizard); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	steps, ok := rawWizard["steps"].([]interface{})
	if !ok {
		log.Debug().Msg("No steps found in YAML")
		return nil // No steps to process
	}
	
	log.Debug().Int("yamlSteps", len(steps)).Msg("Found steps in YAML")

	// Create a map of step IDs to their raw YAML data
	stepMap := make(map[string]map[string]interface{})
	for i, stepRaw := range steps {
		stepData, ok := stepRaw.(map[string]interface{})
		if !ok {
			continue
		}
		stepID, _ := stepData["id"].(string)
		if stepID == "" {
			stepID = fmt.Sprintf("step-%d", i)
		}
		stepMap[stepID] = stepData
		log.Debug().Str("stepID", stepID).Str("type", fmt.Sprintf("%v", stepData["type"])).Msg("Mapped step from YAML")
	}

	// Fix DecisionStep choices
	for i, step := range wz.Steps {
		ds, ok := step.(*uhsteps.DecisionStep)
		if !ok {
			continue
		}
		
		stepID := ds.ID()
		if stepID == "" {
			stepID = fmt.Sprintf("step-%d", i)
		}
		
		log.Debug().Str("stepID", stepID).Msg("Processing DecisionStep")
		
		// Get raw step data
		stepData, exists := stepMap[stepID]
		if !exists {
			log.Debug().Str("stepID", stepID).Msg("Step not found in stepMap")
			// Try to find by index
			if i < len(steps) {
				if stepRaw, ok := steps[i].(map[string]interface{}); ok {
					stepData = stepRaw
					exists = true
					log.Debug().Int("index", i).Msg("Found step by index")
				}
			}
			if !exists {
				continue
			}
		}

		// Extract options from YAML
		optionsRaw, ok := stepData["options"].([]interface{})
		if !ok {
			log.Debug().Str("stepID", stepID).Interface("stepData", stepData).Msg("No options field found in step data")
			continue
		}

		choices := make([]string, 0, len(optionsRaw))
		for _, optRaw := range optionsRaw {
			optMap, ok := optRaw.(map[string]interface{})
			if !ok {
				continue
			}
			// Prefer label, fallback to value
			label, _ := optMap["label"].(string)
			value, _ := optMap["value"].(string)
			if label != "" {
				choices = append(choices, label)
			} else if value != "" {
				choices = append(choices, value)
			}
		}

		// Update the DecisionStep choices
		if len(choices) > 0 {
			log.Debug().Str("stepID", stepID).Int("choices", len(choices)).Strs("choices", choices).Msg("Updating DecisionStep choices")
			ds.Choices = choices
		} else {
			log.Debug().Str("stepID", stepID).Msg("No choices extracted from options")
		}
	}

	return nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &GenerateCommand{}
