package playbook

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
)

// RegisterCommand registers a playbook from file or URL
type RegisterCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type RegisterSettings struct {
	Title       string   `glazed.parameter:"title"`
	Description string   `glazed.parameter:"description"`
	Summary     string   `glazed.parameter:"summary"`
	Metadata    []string `glazed.parameter:"metadata"`
	Tags        []string `glazed.parameter:"tags"`
	Filename    string   `glazed.parameter:"filename"`
	Source      string   `glazed.parameter:"source"`
	IsCommand   bool     `glazed.parameter:"command"`
}

var _ cmds.WriterCommand = &RegisterCommand{}

func NewRegisterCommand(storage *Storage) (*RegisterCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"register",
		cmds.WithShort("Register a playbook from file, URL, or shell command"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"source",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path, URL to the playbook, or shell command (when using --command)"),
				parameters.WithRequired(true),
			),
		),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"title",
				parameters.ParameterTypeString,
				parameters.WithHelp("Title of the playbook"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"description",
				parameters.ParameterTypeString,
				parameters.WithHelp("Description of the playbook"),
			),
			parameters.NewParameterDefinition(
				"summary",
				parameters.ParameterTypeString,
				parameters.WithHelp("Summary of the playbook"),
			),
			parameters.NewParameterDefinition(
				"metadata",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Metadata in key=value format"),
			),
			parameters.NewParameterDefinition(
				"tags",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Tags for the playbook"),
			),
			parameters.NewParameterDefinition(
				"filename",
				parameters.ParameterTypeString,
				parameters.WithHelp("Override filename"),
			),
			parameters.NewParameterDefinition(
				"command",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Register source as a shell command instead of file/URL"),
				parameters.WithDefault(false),
			),
		),
	)

	return &RegisterCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *RegisterCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &RegisterSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Create entity
	entity := &Entity{
		Type:        TypePlaybook,
		Title:       s.Title,
		Description: s.Description,
		Summary:     s.Summary,
		Tags:        s.Tags,
		LastFetched: timePtr(time.Now()),
	}

	if s.IsCommand {
		// Register as shell command
		entity.Command = &s.Source
		if s.Filename == "" {
			s.Filename = s.Title + "-output.txt"
		}
	} else {
		// Read content from file or URL
		var content string
		var canonicalURL *string
		var err error

		if strings.HasPrefix(s.Source, "http://") || strings.HasPrefix(s.Source, "https://") {
			content, err = fetchURL(s.Source)
			canonicalURL = &s.Source
		} else {
			content, err = readFile(s.Source)
			if s.Filename == "" {
				s.Filename = filepath.Base(s.Source)
			}
		}
		if err != nil {
			return fmt.Errorf("failed to read content: %w", err)
		}

		entity.CanonicalURL = canonicalURL
		entity.Content = &content
	}

	if s.Filename != "" {
		entity.Filename = &s.Filename
	}

	// Parse metadata
	if len(s.Metadata) > 0 {
		entity.Metadata = make(map[string]string)
		for _, meta := range s.Metadata {
			parts := strings.SplitN(meta, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid metadata format: %s (expected key=value)", meta)
			}
			entity.Metadata[parts[0]] = parts[1]
		}
	}

	if err := c.storage.CreateEntity(entity); err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	fmt.Fprintf(w, "Registered playbook: %s (slug: %s)\n", entity.Title, entity.Slug)
	return nil
}

// CreateCollectionCommand creates a new collection
type CreateCollectionCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type CreateCollectionSettings struct {
	Name        string   `glazed.parameter:"name"`
	Title       string   `glazed.parameter:"title"`
	Description string   `glazed.parameter:"description"`
	Summary     string   `glazed.parameter:"summary"`
	Metadata    []string `glazed.parameter:"metadata"`
	Tags        []string `glazed.parameter:"tags"`
}

var _ cmds.WriterCommand = &CreateCollectionCommand{}

func NewCreateCollectionCommand(storage *Storage) (*CreateCollectionCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"create-collection",
		cmds.WithShort("Create a new collection"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the collection"),
				parameters.WithRequired(true),
			),
		),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"title",
				parameters.ParameterTypeString,
				parameters.WithHelp("Display title (defaults to name)"),
			),
			parameters.NewParameterDefinition(
				"description",
				parameters.ParameterTypeString,
				parameters.WithHelp("Description of the collection"),
			),
			parameters.NewParameterDefinition(
				"summary",
				parameters.ParameterTypeString,
				parameters.WithHelp("Summary of the collection"),
			),
			parameters.NewParameterDefinition(
				"metadata",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Metadata in key=value format"),
			),
			parameters.NewParameterDefinition(
				"tags",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Tags for the collection"),
			),
		),
	)

	return &CreateCollectionCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *CreateCollectionCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &CreateCollectionSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	title := s.Title
	if title == "" {
		title = s.Name
	}

	entity := &Entity{
		Type:        TypeCollection,
		Title:       title,
		Description: s.Description,
		Summary:     s.Summary,
		Tags:        s.Tags,
	}

	// Parse metadata
	if len(s.Metadata) > 0 {
		entity.Metadata = make(map[string]string)
		for _, meta := range s.Metadata {
			parts := strings.SplitN(meta, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid metadata format: %s (expected key=value)", meta)
			}
			entity.Metadata[parts[0]] = parts[1]
		}
	}

	if err := c.storage.CreateEntity(entity); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	fmt.Fprintf(w, "Created collection: %s (slug: %s)\n", entity.Title, entity.Slug)
	return nil
}

// ListCommand lists entities with optional filters
type ListCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type ListSettings struct {
	EntityType string   `glazed.parameter:"type"`
	Tags       []string `glazed.parameter:"tags"`
}

var _ cmds.GlazeCommand = &ListCommand{}

func NewListCommand(storage *Storage) (*ListCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List entities with optional filters"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"type",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Filter by type"),
				parameters.WithChoices("playbook", "collection"),
			),
			parameters.NewParameterDefinition(
				"tags",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Filter by tags"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &ListCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *ListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &ListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	var typeFilter *EntityType
	if s.EntityType != "" {
		switch s.EntityType {
		case "playbook":
			t := TypePlaybook
			typeFilter = &t
		case "collection":
			t := TypeCollection
			typeFilter = &t
		default:
			return fmt.Errorf("invalid type: %s (must be 'playbook' or 'collection')", s.EntityType)
		}
	}

	entities, err := c.storage.ListEntities(typeFilter, s.Tags)
	if err != nil {
		return fmt.Errorf("failed to list entities: %w", err)
	}

	for _, entity := range entities {
		row := types.NewRow(
			types.MRP("slug", entity.Slug),
			types.MRP("type", string(entity.Type)),
			types.MRP("title", entity.Title),
			types.MRP("summary", entity.Summary),
			types.MRP("description", entity.Description),
			types.MRP("tags", strings.Join(entity.Tags, ", ")),
			types.MRP("created_at", entity.CreatedAt.Format(time.RFC3339)),
		)

		if entity.CanonicalURL != nil {
			row.Set("canonical_url", *entity.CanonicalURL)
		}
		if entity.LastFetched != nil {
			row.Set("last_fetched", entity.LastFetched.Format(time.RFC3339))
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// SearchCommand searches entities
type SearchCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type SearchSettings struct {
	Query      string `glazed.parameter:"query"`
	EntityType string `glazed.parameter:"type"`
}

var _ cmds.GlazeCommand = &SearchCommand{}

func NewSearchCommand(storage *Storage) (*SearchCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"search",
		cmds.WithShort("Search entities by query string"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("Search query"),
				parameters.WithRequired(true),
			),
		),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"type",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Filter by type"),
				parameters.WithChoices("playbook", "collection"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &SearchCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *SearchCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &SearchSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	var typeFilter *EntityType
	if s.EntityType != "" {
		switch s.EntityType {
		case "playbook":
			t := TypePlaybook
			typeFilter = &t
		case "collection":
			t := TypeCollection
			typeFilter = &t
		default:
			return fmt.Errorf("invalid type: %s (must be 'playbook' or 'collection')", s.EntityType)
		}
	}

	entities, err := c.storage.SearchEntities(s.Query, typeFilter)
	if err != nil {
		return fmt.Errorf("failed to search entities: %w", err)
	}

	for _, entity := range entities {
		row := types.NewRow(
			types.MRP("slug", entity.Slug),
			types.MRP("type", string(entity.Type)),
			types.MRP("title", entity.Title),
			types.MRP("summary", entity.Summary),
			types.MRP("description", entity.Description),
			types.MRP("tags", strings.Join(entity.Tags, ", ")),
			types.MRP("created_at", entity.CreatedAt.Format(time.RFC3339)),
		)

		if entity.CanonicalURL != nil {
			row.Set("canonical_url", *entity.CanonicalURL)
		}
		if entity.LastFetched != nil {
			row.Set("last_fetched", entity.LastFetched.Format(time.RFC3339))
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// ShowCommand shows detailed information about an entity
type ShowCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type ShowSettings struct {
	Slug string `glazed.parameter:"slug"`
}

var _ cmds.WriterCommand = &ShowCommand{}

func NewShowCommand(storage *Storage) (*ShowCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"show",
		cmds.WithShort("Show detailed information about an entity"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Entity slug"),
				parameters.WithRequired(true),
			),
		),
	)

	return &ShowCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *ShowCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &ShowSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	entity, err := c.storage.GetEntityBySlug(s.Slug)
	if err != nil {
		return fmt.Errorf("failed to get entity: %w", err)
	}

	fmt.Fprintf(w, "Slug: %s\n", entity.Slug)
	fmt.Fprintf(w, "Type: %s\n", entity.Type)
	fmt.Fprintf(w, "Title: %s\n", entity.Title)
	if entity.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", entity.Description)
	}
	if entity.Summary != "" {
		fmt.Fprintf(w, "Summary: %s\n", entity.Summary)
	}
	if entity.CanonicalURL != nil {
		fmt.Fprintf(w, "URL: %s\n", *entity.CanonicalURL)
	}
	if entity.Filename != nil {
		fmt.Fprintf(w, "Filename: %s\n", *entity.Filename)
	}
	if len(entity.Tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n", strings.Join(entity.Tags, ", "))
	}
	if entity.LastFetched != nil {
		fmt.Fprintf(w, "Last Fetched: %s\n", entity.LastFetched.Format(time.RFC3339))
	}
	fmt.Fprintf(w, "Created: %s\n", entity.CreatedAt.Format(time.RFC3339))

	// Show metadata
	if len(entity.Metadata) > 0 {
		fmt.Fprintf(w, "\nMetadata:\n")
		for key, value := range entity.Metadata {
			fmt.Fprintf(w, "  %s: %s\n", key, value)
		}
	}

	// Show collection members if it's a collection
	if entity.Type == TypeCollection {
		members, err := c.storage.GetCollectionMembers(entity.ID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get collection members")
		} else if len(members) > 0 {
			fmt.Fprintf(w, "\nMembers:\n")
			for _, member := range members {
				memberEntity, err := c.storage.GetEntityByID(member.MemberID)
				if err != nil {
					fmt.Fprintf(w, "  ID %d (error loading details)\n", member.MemberID)
				} else {
					fmt.Fprintf(w, "  %s (%s) - %s\n", memberEntity.Slug, memberEntity.Type, memberEntity.Title)
					if member.RelativePath != nil {
						fmt.Fprintf(w, "    Path: %s\n", *member.RelativePath)
					}
				}
			}
		}
	}

	// Show content or command if it's a playbook
	if entity.Type == TypePlaybook {
		if entity.IsCommand() {
			fmt.Fprintf(w, "\nCommand:\n")
			fmt.Fprintf(w, "%s\n", *entity.Command)
		} else if entity.Content != nil {
			fmt.Fprintf(w, "\nContent:\n")
			fmt.Fprintf(w, "%s\n", *entity.Content)
		}
	}

	return nil
}

// DeployCommand deploys entities to directories
type DeployCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type DeploySettings struct {
	Slug             string `glazed.parameter:"slug"`
	TargetDirectory  string `glazed.parameter:"target_directory"`
	FilenameOverride string `glazed.parameter:"filename"`
}

var _ cmds.WriterCommand = &DeployCommand{}

func NewDeployCommand(storage *Storage) (*DeployCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"deploy",
		cmds.WithShort("Deploy an entity to a target directory"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Entity slug"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"target_directory",
				parameters.ParameterTypeString,
				parameters.WithHelp("Target directory"),
				parameters.WithRequired(true),
			),
		),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"filename",
				parameters.ParameterTypeString,
				parameters.WithHelp("Override filename for playbook deployment"),
			),
		),
	)

	return &DeployCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *DeployCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &DeploySettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	entity, err := c.storage.GetEntityBySlug(s.Slug)
	if err != nil {
		return fmt.Errorf("failed to get entity: %w", err)
	}

	if err := os.MkdirAll(s.TargetDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	if entity.Type == TypePlaybook {
		// Deploy single playbook
		filename := s.FilenameOverride
		if filename == "" && entity.Filename != nil {
			filename = *entity.Filename
		}
		if filename == "" {
			if entity.IsCommand() {
				filename = entity.Slug + "-output.txt"
			} else {
				filename = entity.Slug + ".md"
			}
		}

		filePath := filepath.Join(s.TargetDirectory, filename)

		if entity.IsCommand() {
			// Execute shell command and capture output
			if err := executeCommandToFile(ctx, *entity.Command, filePath); err != nil {
				return fmt.Errorf("failed to execute command: %w", err)
			}
			fmt.Fprintf(w, "Executed command and deployed output %s to %s\n", entity.Title, filePath)
		} else {
			// Deploy regular content
			if err := os.WriteFile(filePath, []byte(*entity.Content), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			fmt.Fprintf(w, "Deployed playbook %s to %s\n", entity.Title, filePath)
		}
	} else {
		// Deploy collection - create subdirectory with collection slug
		collectionDir := filepath.Join(s.TargetDirectory, entity.Slug)
		if err := os.MkdirAll(collectionDir, 0755); err != nil {
			return fmt.Errorf("failed to create collection directory: %w", err)
		}

		members, err := c.storage.GetCollectionMembers(entity.ID)
		if err != nil {
			return fmt.Errorf("failed to get collection members: %w", err)
		}

		for _, member := range members {
			memberEntity, err := c.storage.GetEntityByID(member.MemberID)
			if err != nil {
				log.Warn().Err(err).Int64("member_id", member.MemberID).Msg("Failed to load member")
				continue
			}

			if memberEntity.Type == TypePlaybook {
				filename := ""
				if member.RelativePath != nil {
					filename = *member.RelativePath
				} else if memberEntity.Filename != nil {
					filename = *memberEntity.Filename
				} else {
					if memberEntity.IsCommand() {
						filename = memberEntity.Slug + "-output.txt"
					} else {
						filename = memberEntity.Slug + ".md"
					}
				}

				filePath := filepath.Join(collectionDir, filename)
				
				// Create parent directory if needed
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					log.Warn().Err(err).Str("path", filepath.Dir(filePath)).Msg("Failed to create directory")
					continue
				}

				if memberEntity.IsCommand() {
					// Execute shell command and capture output
					if err := executeCommandToFile(ctx, *memberEntity.Command, filePath); err != nil {
						log.Warn().Err(err).Str("command", *memberEntity.Command).Msg("Failed to execute command")
						continue
					}
					fmt.Fprintf(w, "Executed command and deployed %s to %s\n", memberEntity.Title, filePath)
				} else if memberEntity.Content != nil {
					// Deploy regular content
					if err := os.WriteFile(filePath, []byte(*memberEntity.Content), 0644); err != nil {
						log.Warn().Err(err).Str("path", filePath).Msg("Failed to write file")
						continue
					}
					fmt.Fprintf(w, "Deployed %s to %s\n", memberEntity.Title, filePath)
				}
			}
		}

		fmt.Fprintf(w, "Deployed collection %s to %s\n", entity.Title, collectionDir)
	}

	// Record deployment
	if err := c.storage.RecordDeployment(entity.ID, s.TargetDirectory); err != nil {
		log.Warn().Err(err).Msg("Failed to record deployment")
	}

	return nil
}

// Helper functions

func fetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// executeCommandToFile executes a shell command and writes stdout and stderr to a file
func executeCommandToFile(ctx context.Context, command, filePath string) error {
	// Create the file
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Write command header
	fmt.Fprintf(outFile, "# Command: %s\n", command)
	fmt.Fprintf(outFile, "# Executed at: %s\n\n", time.Now().Format(time.RFC3339))

	// Execute the command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = outFile
	cmd.Stderr = outFile

	if err := cmd.Run(); err != nil {
		// Still write the error to the file but also return it
		fmt.Fprintf(outFile, "\n# Command failed with error: %v\n", err)
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// AddToCollectionCommand adds entities to collections
type AddToCollectionCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type AddToCollectionSettings struct {
	CollectionSlug string `glazed.parameter:"collection_slug"`
	MemberSlug     string `glazed.parameter:"member_slug"`
	RelativePath   string `glazed.parameter:"path"`
}

var _ cmds.WriterCommand = &AddToCollectionCommand{}

func NewAddToCollectionCommand(storage *Storage) (*AddToCollectionCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"add",
		cmds.WithShort("Add a member to a collection"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"collection_slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Collection slug"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"member_slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Member slug"),
				parameters.WithRequired(true),
			),
		),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"path",
				parameters.ParameterTypeString,
				parameters.WithHelp("Relative path within collection"),
			),
		),
	)

	return &AddToCollectionCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *AddToCollectionCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &AddToCollectionSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	collection, err := c.storage.GetEntityBySlug(s.CollectionSlug)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}
	if collection.Type != TypeCollection {
		return fmt.Errorf("%s is not a collection", s.CollectionSlug)
	}

	member, err := c.storage.GetEntityBySlug(s.MemberSlug)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	var relPath *string
	if s.RelativePath != "" {
		relPath = &s.RelativePath
	}

	if err := c.storage.AddToCollection(collection.ID, member.ID, relPath); err != nil {
		return fmt.Errorf("failed to add to collection: %w", err)
	}

	fmt.Fprintf(w, "Added %s to collection %s\n", member.Title, collection.Title)
	return nil
}

// RemoveCommand removes entities from collections or deletes entities
type RemoveCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type RemoveSettings struct {
	CollectionSlug string `glazed.parameter:"collection_slug"`
	MemberSlug     string `glazed.parameter:"member_slug"`
	EntitySlug     string `glazed.parameter:"entity_slug"`
}

var _ cmds.WriterCommand = &RemoveCommand{}

func NewRemoveCommand(storage *Storage) (*RemoveCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"remove",
		cmds.WithShort("Remove a member from a collection or delete an entity"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"collection_slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Collection slug (for removing from collection)"),
			),
			parameters.NewParameterDefinition(
				"member_slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Member slug (for removing from collection) or entity slug (for deletion)"),
			),
		),
	)

	return &RemoveCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *RemoveCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &RemoveSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	if s.MemberSlug == "" {
		return fmt.Errorf("missing required argument: member_slug or entity_slug")
	}

	if s.CollectionSlug == "" {
		// Delete entity
		entity, err := c.storage.GetEntityBySlug(s.MemberSlug)
		if err != nil {
			return fmt.Errorf("failed to get entity: %w", err)
		}

		if err := c.storage.DeleteEntity(entity.ID); err != nil {
			return fmt.Errorf("failed to delete entity: %w", err)
		}

		fmt.Fprintf(w, "Deleted %s (%s)\n", entity.Title, entity.Type)
		return nil
	}

	// Remove from collection
	collection, err := c.storage.GetEntityBySlug(s.CollectionSlug)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}
	if collection.Type != TypeCollection {
		return fmt.Errorf("%s is not a collection", s.CollectionSlug)
	}

	member, err := c.storage.GetEntityBySlug(s.MemberSlug)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	if err := c.storage.RemoveFromCollection(collection.ID, member.ID); err != nil {
		return fmt.Errorf("failed to remove from collection: %w", err)
	}

	fmt.Fprintf(w, "Removed %s from collection %s\n", member.Title, collection.Title)
	return nil
}

// SetMetadataCommand sets metadata for an entity
type SetMetadataCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type SetMetadataSettings struct {
	Slug  string `glazed.parameter:"slug"`
	Key   string `glazed.parameter:"key"`
	Value string `glazed.parameter:"value"`
}

var _ cmds.WriterCommand = &SetMetadataCommand{}

func NewSetMetadataCommand(storage *Storage) (*SetMetadataCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"set-meta",
		cmds.WithShort("Set metadata for an entity"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Entity slug"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"key",
				parameters.ParameterTypeString,
				parameters.WithHelp("Metadata key"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"value",
				parameters.ParameterTypeString,
				parameters.WithHelp("Metadata value"),
				parameters.WithRequired(true),
			),
		),
	)

	return &SetMetadataCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *SetMetadataCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &SetMetadataSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	entity, err := c.storage.GetEntityBySlug(s.Slug)
	if err != nil {
		return fmt.Errorf("failed to get entity: %w", err)
	}

	if err := c.storage.SetMetadata(entity.ID, s.Key, s.Value); err != nil {
		return fmt.Errorf("failed to set metadata: %w", err)
	}

	fmt.Fprintf(w, "Set metadata %s=%s for %s\n", s.Key, s.Value, entity.Title)
	return nil
}

// GetMetadataCommand gets metadata for an entity
type GetMetadataCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type GetMetadataSettings struct {
	Slug string `glazed.parameter:"slug"`
	Key  string `glazed.parameter:"key"`
}

var _ cmds.GlazeCommand = &GetMetadataCommand{}

func NewGetMetadataCommand(storage *Storage) (*GetMetadataCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"get-meta",
		cmds.WithShort("Get metadata for an entity"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Entity slug"),
				parameters.WithRequired(true),
			),
		),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"key",
				parameters.ParameterTypeString,
				parameters.WithHelp("Specific metadata key (optional)"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &GetMetadataCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *GetMetadataCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &GetMetadataSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	entity, err := c.storage.GetEntityBySlug(s.Slug)
	if err != nil {
		return fmt.Errorf("failed to get entity: %w", err)
	}

	metadata, err := c.storage.GetMetadata(entity.ID)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	if s.Key != "" {
		// Get specific key
		if value, exists := metadata[s.Key]; exists {
			row := types.NewRow(
				types.MRP("key", s.Key),
				types.MRP("value", value),
			)
			return gp.AddRow(ctx, row)
		} else {
			return fmt.Errorf("metadata key %s not found", s.Key)
		}
	} else {
		// Get all metadata
		for key, value := range metadata {
			row := types.NewRow(
				types.MRP("key", key),
				types.MRP("value", value),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

// RefreshCommand refreshes deployments by re-executing commands
type RefreshCommand struct {
	*cmds.CommandDescription
	storage *Storage
}

type RefreshSettings struct {
	Slug            string `glazed.parameter:"slug"`
	TargetDirectory string `glazed.parameter:"target_directory"`
}

var _ cmds.WriterCommand = &RefreshCommand{}

func NewRefreshCommand(storage *Storage) (*RefreshCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"refresh",
		cmds.WithShort("Refresh a deployment by re-executing shell commands"),
		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"slug",
				parameters.ParameterTypeString,
				parameters.WithHelp("Entity slug to refresh"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"target_directory",
				parameters.ParameterTypeString,
				parameters.WithHelp("Target directory to refresh"),
				parameters.WithRequired(true),
			),
		),
	)

	return &RefreshCommand{
		CommandDescription: cmdDesc,
		storage:           storage,
	}, nil
}

func (c *RefreshCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	s := &RefreshSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	entity, err := c.storage.GetEntityBySlug(s.Slug)
	if err != nil {
		return fmt.Errorf("failed to get entity: %w", err)
	}

	if entity.Type == TypePlaybook && entity.IsCommand() {
		// Refresh single command playbook
		filename := ""
		if entity.Filename != nil {
			filename = *entity.Filename
		} else {
			filename = entity.Slug + "-output.txt"
		}

		filePath := filepath.Join(s.TargetDirectory, filename)
		if err := executeCommandToFile(ctx, *entity.Command, filePath); err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		fmt.Fprintf(w, "Refreshed command output %s to %s\n", entity.Title, filePath)
	} else if entity.Type == TypeCollection {
		// Refresh collection - re-execute all command playbooks
		collectionDir := filepath.Join(s.TargetDirectory, entity.Slug)
		members, err := c.storage.GetCollectionMembers(entity.ID)
		if err != nil {
			return fmt.Errorf("failed to get collection members: %w", err)
		}

		refreshedCount := 0
		for _, member := range members {
			memberEntity, err := c.storage.GetEntityByID(member.MemberID)
			if err != nil {
				log.Warn().Err(err).Int64("member_id", member.MemberID).Msg("Failed to load member")
				continue
			}

			if memberEntity.Type == TypePlaybook && memberEntity.IsCommand() {
				filename := ""
				if member.RelativePath != nil {
					filename = *member.RelativePath
				} else if memberEntity.Filename != nil {
					filename = *memberEntity.Filename
				} else {
					filename = memberEntity.Slug + "-output.txt"
				}

				filePath := filepath.Join(collectionDir, filename)
				
				// Create parent directory if needed
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					log.Warn().Err(err).Str("path", filepath.Dir(filePath)).Msg("Failed to create directory")
					continue
				}

				if err := executeCommandToFile(ctx, *memberEntity.Command, filePath); err != nil {
					log.Warn().Err(err).Str("command", *memberEntity.Command).Msg("Failed to execute command")
					continue
				}

				fmt.Fprintf(w, "Refreshed command %s to %s\n", memberEntity.Title, filePath)
				refreshedCount++
			}
		}

		fmt.Fprintf(w, "Refreshed %d command outputs in collection %s\n", refreshedCount, entity.Title)
	} else {
		return fmt.Errorf("entity %s is not a command playbook or collection", s.Slug)
	}

	return nil
}
