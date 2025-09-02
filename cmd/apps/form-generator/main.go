package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	gauth "github.com/go-go-golems/go-go-labs/pkg/google/auth"
	gatoken "github.com/go-go-golems/go-go-labs/pkg/google/auth/store"
	uform "github.com/go-go-golems/uhoh/pkg"
	"github.com/go-go-golems/uhoh/pkg/wizard"
	"github.com/go-go-golems/uhoh/pkg/wizard/steps"

	"google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
)

func expandHome(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	if path[0] != '~' {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get home directory")
	}
	return filepath.Join(home, path[1:]), nil
}

func newQuestionItem(title string, q *forms.Question) *forms.Item {
	return &forms.Item{
		Title:        title,
		QuestionItem: &forms.QuestionItem{Question: q},
	}
}

func fieldToItem(f *uform.Field) (*forms.Item, error) {
	switch f.Type {
	case "input":
		return newQuestionItem(f.Title, &forms.Question{
			Required:     false,
			TextQuestion: &forms.TextQuestion{Paragraph: false},
		}), nil
	case "text":
		return newQuestionItem(f.Title, &forms.Question{
			Required:     false,
			TextQuestion: &forms.TextQuestion{Paragraph: true},
		}), nil
	case "select":
		opts := make([]*forms.Option, 0, len(f.Options))
		for _, o := range f.Options {
			label := o.Label
			if label == "" {
				// Fallback to value string representation
				label = fmt.Sprintf("%v", o.Value)
			}
			opts = append(opts, &forms.Option{Value: label})
		}
		return newQuestionItem(f.Title, &forms.Question{
			Required: true,
			ChoiceQuestion: &forms.ChoiceQuestion{
				Type:    "RADIO",
				Options: opts,
				Shuffle: false,
			},
		}), nil
	case "multiselect":
		opts := make([]*forms.Option, 0, len(f.Options))
		for _, o := range f.Options {
			label := o.Label
			if label == "" {
				label = fmt.Sprintf("%v", o.Value)
			}
			opts = append(opts, &forms.Option{Value: label})
		}
		return newQuestionItem(f.Title, &forms.Question{
			Required: false,
			ChoiceQuestion: &forms.ChoiceQuestion{
				Type:    "CHECKBOX",
				Options: opts,
				Shuffle: false,
			},
		}), nil
	case "confirm":
		opts := []*forms.Option{
			{Value: "Yes"},
			{Value: "No"},
		}
		return newQuestionItem(f.Title, &forms.Question{
			Required: true,
			ChoiceQuestion: &forms.ChoiceQuestion{
				Type:    "RADIO",
				Options: opts,
				Shuffle: false,
			},
		}), nil
	default:
		return nil, errors.Errorf("unsupported field type: %s", f.Type)
	}
}

func buildRequestsFromWizard(wz *wizard.Wizard) ([]*forms.Request, error) {
	requests := []*forms.Request{}
	index := 0

	for _, s := range wz.Steps {
		switch st := s.(type) {
		case *steps.FormStep:
			for _, g := range st.FormData.Groups {
				for _, f := range g.Fields {
					item, err := fieldToItem(f)
					if err != nil {
						// Skip unsupported field types rather than failing the whole form
						continue
					}
					requests = append(requests, &forms.Request{
						CreateItem: &forms.CreateItemRequest{
							Item:     item,
							Location: &forms.Location{Index: int64(index)},
						},
					})
					index++
				}
			}
		case *steps.DecisionStep:
			// Map a decision step to a multiple-choice question
			opts := make([]*forms.Option, 0, len(st.Choices))
			for _, c := range st.Choices {
				opts = append(opts, &forms.Option{Value: c})
			}
			item := newQuestionItem(st.Title(), &forms.Question{
				Required: true,
				ChoiceQuestion: &forms.ChoiceQuestion{
					Type:    "RADIO",
					Options: opts,
					Shuffle: false,
				},
			})
			requests = append(requests, &forms.Request{
				CreateItem: &forms.CreateItemRequest{
					Item:     item,
					Location: &forms.Location{Index: int64(index)},
				},
			})
			index++
		default:
			// Ignore info, action, summary for Google Forms structure
			continue
		}
	}

	return requests, nil
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if len(args) < 1 {
		return errors.New("missing DSL file path argument")
	}
	dslPath := args[0]

	credentialsFile, _ := cmd.Flags().GetString("credentials")
	tokenFile, _ := cmd.Flags().GetString("token")
	titleOverride, _ := cmd.Flags().GetString("title")
	descriptionOverride, _ := cmd.Flags().GetString("description")

	// Expand ~ in paths
	var err error
	credentialsFile, err = expandHome(credentialsFile)
	if err != nil {
		return err
	}
	tokenFile, err = expandHome(tokenFile)
	if err != nil {
		return err
	}

	// Load the wizard DSL
	wz, err := wizard.LoadWizard(dslPath)
	if err != nil {
		return errors.Wrap(err, "failed to load wizard DSL")
	}

	formTitle := wz.Name
	if titleOverride != "" {
		formTitle = titleOverride
	}
	if formTitle == "" {
		formTitle = "Generated Form"
	}
	formDescription := wz.Description
	if descriptionOverride != "" {
		formDescription = descriptionOverride
	}

	// Authenticate with Google using our auth helper (Forms scope)
	authenticator, err := gauth.NewAuthenticator(
		gauth.WithScopes(forms.FormsBodyScope),
		gauth.WithCredentialsFile(credentialsFile),
		gauth.WithTokenStore(gatoken.NewFileTokenStore(tokenFile, 0600)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create authenticator")
	}

	result, err := authenticator.Authenticate(ctx)
	if err != nil {
		return errors.Wrap(err, "authentication failed")
	}

	// Create Forms service
	svc, err := forms.NewService(ctx, option.WithTokenSource(result.Client.TokenSource(ctx, result.Token)))
	if err != nil {
		return errors.Wrap(err, "failed to create forms service")
	}

	// Create the form
	form, err := svc.Forms.Create(&forms.Form{
		Info: &forms.Info{Title: formTitle, Description: formDescription},
	}).Do()
	if err != nil {
		return errors.Wrap(err, "failed to create form")
	}
	fmt.Printf("Created form: %s\n", form.FormId)
	if form.ResponderUri != "" {
		fmt.Printf("Responder URL: %s\n", form.ResponderUri)
	}

	// Build and apply batch update requests
	requests, err := buildRequestsFromWizard(wz)
	if err != nil {
		return errors.Wrap(err, "failed to build form items")
	}
	if len(requests) > 0 {
		_, err = svc.Forms.BatchUpdate(form.FormId, &forms.BatchUpdateFormRequest{Requests: requests}).Do()
		if err != nil {
			return errors.Wrap(err, "batch update failed")
		}
	}

	updated, err := svc.Forms.Get(form.FormId).Do()
	if err != nil {
		return errors.Wrap(err, "failed to fetch updated form")
	}
	fmt.Printf("Form title: %s\n", updated.Info.Title)
	if updated.ResponderUri != "" {
		fmt.Printf("Fill-in link: %s\n", updated.ResponderUri)
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "form-generator [wizard.yaml]",
		Short: "Generate a Google Form from a Uhoh Wizard DSL file",
		Args:  cobra.MinimumNArgs(1),
		RunE:  run,
	}

	rootCmd.Flags().String("credentials", "~/.google-form/client_secret.json", "Path to OAuth client secret JSON")
	rootCmd.Flags().String("token", "~/.google-form/token.json", "Path to store OAuth token JSON")
	rootCmd.Flags().String("title", "", "Override form title")
	rootCmd.Flags().String("description", "", "Override form description")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}


