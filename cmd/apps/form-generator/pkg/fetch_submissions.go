package pkg

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	gauth "github.com/go-go-golems/go-go-labs/pkg/google/auth"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
)

type FetchSubmissionsSettings struct {
	FormID string `glazed.parameter:"form-id"`
	Debug  bool   `glazed.parameter:"debug"`
}

type FetchSubmissionsCommand struct {
	*cmds.CommandDescription
}

type Submission struct {
	ResponseID      string             `yaml:"response_id"`
	SubmittedAt     string             `yaml:"submitted_at,omitempty"`
	LastSubmittedAt string             `yaml:"last_submitted_at,omitempty"`
	RespondentEmail string             `yaml:"respondent_email,omitempty"`
	TotalScore      *float64           `yaml:"total_score,omitempty"`
	Answers         []SubmissionAnswer `yaml:"answers"`
}

type SubmissionAnswer struct {
	QuestionID string           `yaml:"question_id"`
	FieldKey   string           `yaml:"field_key,omitempty"`
	FieldTitle string           `yaml:"field_title,omitempty"`
	StepID     string           `yaml:"step_id,omitempty"`
	StepTitle  string           `yaml:"step_title,omitempty"`
	Value      interface{}      `yaml:"value,omitempty"`
	Values     []string         `yaml:"values,omitempty"`
	Files      []SubmissionFile `yaml:"files,omitempty"`
}

type SubmissionFile struct {
	FileID   string `yaml:"file_id"`
	FileName string `yaml:"file_name"`
	MimeType string `yaml:"mime_type,omitempty"`
}

func NewFetchSubmissionsCommand() (*cobra.Command, error) {
	oauthLayers, err := gauth.GetOAuthTokenStoreLayersWithOptions(
		gauth.WithCredentialsDefault("~/.google-form/client_secret.json"),
		gauth.WithTokenDefault("~/.google-form/token.json"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create OAuth token store layers: %w", err)
	}

	desc := cmds.NewCommandDescription(
		"fetch-submissions",
		cmds.WithShort("Fetch submitted responses for a Google Form"),
		cmds.WithLong(`
Download all responses for a form and map them back to the generated Wizard DSL fields.
`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"form-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Existing Google Form ID"),
				parameters.WithRequired(true),
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

	c := &FetchSubmissionsCommand{CommandDescription: desc}
	return cli.BuildCobraCommand(c)
}

func (c *FetchSubmissionsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &FetchSubmissionsSettings{}
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

	ts := result.Client.TokenSource(ctx, result.Token)
	svc, err := forms.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("failed to create forms service: %w", err)
	}

	form, err := svc.Forms.Get(settings.FormID).Do()
	if err != nil {
		return fmt.Errorf("failed to fetch form %s: %w", settings.FormID, err)
	}

	_, fieldMeta, err := ConvertFormToWizard(form)
	if err != nil {
		return fmt.Errorf("failed to build metadata from form: %w", err)
	}

	responses, err := fetchAllResponses(ctx, svc, settings.FormID)
	if err != nil {
		return fmt.Errorf("failed to fetch form responses: %w", err)
	}

	records := buildSubmissionRecords(responses, fieldMeta)
	return emitSubmissionRows(ctx, gp, records)
}

func fetchAllResponses(ctx context.Context, svc *forms.Service, formID string) ([]*forms.FormResponse, error) {
	var all []*forms.FormResponse
	pageToken := ""
	for {
		call := svc.Forms.Responses.List(formID)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Context(ctx).Do()
		if err != nil {
			return nil, err
		}
		all = append(all, res.Responses...)
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	return all, nil
}

func buildSubmissionRecords(responses []*forms.FormResponse, meta map[string]FieldMetadata) []Submission {
	records := make([]Submission, 0, len(responses))
	for _, resp := range responses {
		if resp == nil {
			continue
		}
		record := Submission{
			ResponseID:      resp.ResponseId,
			SubmittedAt:     resp.CreateTime,
			LastSubmittedAt: resp.LastSubmittedTime,
			RespondentEmail: resp.RespondentEmail,
		}
		if resp.TotalScore != 0 {
			score := resp.TotalScore
			record.TotalScore = &score
		}
		answers := make([]SubmissionAnswer, 0, len(resp.Answers))
		for qid, answer := range resp.Answers {
			ans := SubmissionAnswer{QuestionID: qid}
			if metaInfo, ok := meta[qid]; ok {
				ans.FieldKey = metaInfo.FieldKey
				ans.FieldTitle = metaInfo.FieldTitle
				ans.StepID = metaInfo.StepID
				ans.StepTitle = metaInfo.StepTitle
			}
			if answer.TextAnswers != nil {
				values := make([]string, 0, len(answer.TextAnswers.Answers))
				for _, ta := range answer.TextAnswers.Answers {
					if ta == nil {
						continue
					}
					values = append(values, strings.TrimSpace(ta.Value))
				}
				if len(values) == 1 {
					ans.Value = values[0]
				} else if len(values) > 1 {
					ans.Values = values
				}
			}
			if answer.FileUploadAnswers != nil {
				files := make([]SubmissionFile, 0, len(answer.FileUploadAnswers.Answers))
				for _, fa := range answer.FileUploadAnswers.Answers {
					if fa == nil {
						continue
					}
					files = append(files, SubmissionFile{
						FileID:   fa.FileId,
						FileName: fa.FileName,
						MimeType: fa.MimeType,
					})
				}
				if len(files) > 0 {
					ans.Files = files
				}
			}
			answers = append(answers, ans)
		}
		sort.Slice(answers, func(i, j int) bool {
			if answers[i].StepID == answers[j].StepID {
				return answers[i].FieldKey < answers[j].FieldKey
			}
			return answers[i].StepID < answers[j].StepID
		})
		record.Answers = answers
		records = append(records, record)
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].SubmittedAt == records[j].SubmittedAt {
			return records[i].ResponseID < records[j].ResponseID
		}
		return records[i].SubmittedAt < records[j].SubmittedAt
	})

	return records
}

func emitSubmissionRows(ctx context.Context, gp middlewares.Processor, records []Submission) error {
	for _, record := range records {
		basePairs := []types.MapRowPair{
			types.MRP("response_id", record.ResponseID),
			types.MRP("submitted_at", record.SubmittedAt),
			types.MRP("last_submitted_at", record.LastSubmittedAt),
			types.MRP("respondent_email", record.RespondentEmail),
		}
		scoreValue := interface{}(nil)
		if record.TotalScore != nil {
			scoreValue = *record.TotalScore
		}
		basePairs = append(basePairs, types.MRP("total_score", scoreValue))

		if len(record.Answers) == 0 {
			row := types.NewRow(basePairs...)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		for idx, answer := range record.Answers {
			pairs := make([]types.MapRowPair, 0, len(basePairs)+9)
			pairs = append(pairs, basePairs...)
			pairs = append(pairs,
				types.MRP("answer_index", idx),
				types.MRP("question_id", answer.QuestionID),
				types.MRP("field_key", answer.FieldKey),
				types.MRP("field_title", answer.FieldTitle),
				types.MRP("step_id", answer.StepID),
				types.MRP("step_title", answer.StepTitle),
				types.MRP("value", answer.Value),
				types.MRP("values", answer.Values),
			)

			if len(answer.Files) > 0 {
				files := make([]map[string]string, 0, len(answer.Files))
				for _, file := range answer.Files {
					files = append(files, map[string]string{
						"file_id":   file.FileID,
						"file_name": file.FileName,
						"mime_type": file.MimeType,
					})
				}
				pairs = append(pairs, types.MRP("files", files))
			} else {
				pairs = append(pairs, types.MRP("files", []map[string]string{}))
			}

			row := types.NewRow(pairs...)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &FetchSubmissionsCommand{}
