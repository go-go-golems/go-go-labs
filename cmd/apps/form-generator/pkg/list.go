package pkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	gauth "github.com/go-go-golems/go-go-labs/pkg/google/auth"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const driveFormsListFields = "files(id,name,createdTime,modifiedTime,owners(displayName,emailAddress),webViewLink),nextPageToken"

type ListFormsSettings struct {
	Sort  string `glazed.parameter:"sort"`
	Desc  bool   `glazed.parameter:"desc"`
	Limit int64  `glazed.parameter:"limit"`
	Debug bool   `glazed.parameter:"debug"`
}

type ListFormsCommand struct {
	*cmds.CommandDescription
}

func NewListFormsCommand() (*cobra.Command, error) {
	oauthLayers, err := gauth.GetOAuthTokenStoreLayersWithOptions(
		gauth.WithCredentialsDefault("~/.google-form/client_secret.json"),
		gauth.WithTokenDefault("~/.google-form/token.json"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create OAuth token store layers: %w", err)
	}

	glazedLayers, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create glazed parameter layers: %w", err)
	}

	layersList := oauthLayers.Clone().AsList()
	layersList = append(layersList, glazedLayers.Clone())

	desc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List Google Forms available in Drive"),
		cmds.WithLong(`Retrieve metadata about Google Forms stored in Drive. Results can be sorted by name, creation date, or modified date and formatted using Glazed output flags.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"sort",
				parameters.ParameterTypeString,
				parameters.WithDefault("name"),
				parameters.WithHelp("Sort by 'name', 'created', or 'modified'"),
			),
			parameters.NewParameterDefinition(
				"desc",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Sort in descending order"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Limit the number of returned forms"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"debug",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable debug logging"),
				parameters.WithDefault(false),
			),
		),
		cmds.WithLayersList(layersList...),
	)

	c := &ListFormsCommand{CommandDescription: desc}
	return cli.BuildCobraCommand(c)
}

func (c *ListFormsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListFormsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	if settings.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	scope := drive.DriveMetadataReadonlyScope
	authenticator, err := buildFormsAuthenticator(parsedLayers, scope)
	if err != nil {
		return fmt.Errorf("failed to create authenticator: %w", err)
	}
	result, err := authenticator.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ts := result.Client.TokenSource(ctx, result.Token)

	driveService, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("failed to create Drive service: %w", err)
	}

	orderBy, err := resolveSort(settings.Sort, settings.Desc)
	if err != nil {
		return err
	}

	limit := settings.Limit
	if limit < 0 {
		limit = 0
	}

	const defaultPageSize int64 = 100
	pageSize := defaultPageSize
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}

	query := "mimeType='application/vnd.google-apps.form' and trashed=false"
	request := newDriveFormsListCall(driveService, query, orderBy, "", pageSize)

	fetched := int64(0)

	for {
		resp, err := request.Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to list forms: %w", err)
		}

		for _, file := range resp.Files {
			row, err := buildFormRow(file, fetched)
			if err != nil {
				return err
			}
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			fetched++
			if limit > 0 && fetched >= limit {
				return nil
			}
		}

		if resp.NextPageToken == "" {
			break
		}

		if limit > 0 {
			remaining := limit - fetched
			if remaining <= 0 {
				return nil
			}
			if remaining < pageSize {
				pageSize = remaining
			} else {
				pageSize = defaultPageSize
			}
		}
		request = newDriveFormsListCall(driveService, query, orderBy, resp.NextPageToken, pageSize)
	}

	return nil
}

func resolveSort(sort string, desc bool) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(sort))
	var field string
	switch normalized {
	case "name", "title":
		field = "name"
	case "created", "createdtime", "created_time", "creation":
		field = "createdTime"
	case "modified", "modifiedtime", "modified_time", "updated":
		field = "modifiedTime"
	default:
		return "", fmt.Errorf("unsupported sort field: %s", sort)
	}
	if desc {
		field = field + " desc"
	}
	return field, nil
}

func buildFormRow(file *drive.File, index int64) (types.Row, error) {
	if file == nil {
		return nil, fmt.Errorf("encountered nil file metadata")
	}

	createdValue := parseDriveTime(file.CreatedTime)
	modifiedValue := parseDriveTime(file.ModifiedTime)

	ownerNames := make([]string, 0, len(file.Owners))
	ownerEmails := make([]string, 0, len(file.Owners))
	for _, owner := range file.Owners {
		if owner == nil {
			continue
		}
		if owner.DisplayName != "" {
			ownerNames = append(ownerNames, owner.DisplayName)
		}
		if owner.EmailAddress != "" {
			ownerEmails = append(ownerEmails, owner.EmailAddress)
		}
	}

	return types.NewRow(
		types.MRP("index", index+1),
		types.MRP("id", file.Id),
		types.MRP("name", file.Name),
		types.MRP("created_time", createdValue),
		types.MRP("modified_time", modifiedValue),
		types.MRP("owner_names", ownerNames),
		types.MRP("owner_emails", ownerEmails),
		types.MRP("web_view_link", file.WebViewLink),
	), nil
}

func parseDriveTime(value string) interface{} {
	if value == "" {
		return value
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return value
}

var _ cmds.GlazeCommand = &ListFormsCommand{}

func newDriveFormsListCall(service *drive.Service, query, orderBy, pageToken string, pageSize int64) *drive.FilesListCall {
	if pageSize <= 0 {
		pageSize = 1
	}
	call := service.Files.List().
		Q(query).
		Spaces("drive").
		Fields(driveFormsListFields).
		OrderBy(orderBy).
		PageSize(pageSize).
		SupportsAllDrives(false).
		IncludeItemsFromAllDrives(false)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	return call
}
