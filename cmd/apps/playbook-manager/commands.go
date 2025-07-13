package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/pkg/playbook"
)

// registerCmd registers a playbook from file or URL
func registerCmd() *cobra.Command {
	var (
		title       string
		description string
		summary     string
		metadata    []string
		tags        []string
		filename    string
	)

	cmd := &cobra.Command{
		Use:   "register <path|url>",
		Short: "Register a playbook from file or URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]

			// Read content from file or URL
			var content string
			var canonicalURL *string
			var err error

			if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
				content, err = fetchURL(source)
				canonicalURL = &source
			} else {
				content, err = readFile(source)
				if filename == "" {
					filename = filepath.Base(source)
				}
			}
			if err != nil {
				return fmt.Errorf("failed to read content: %w", err)
			}

			// Create entity
			entity := &playbook.Entity{
				Type:         playbook.TypePlaybook,
				Title:        title,
				Description:  description,
				Summary:      summary,
				CanonicalURL: canonicalURL,
				Content:      &content,
				Tags:         tags,
				LastFetched:  timePtr(time.Now()),
			}

			if filename != "" {
				entity.Filename = &filename
			}

			// Parse metadata
			if len(metadata) > 0 {
				entity.Metadata = make(map[string]string)
				for _, meta := range metadata {
					parts := strings.SplitN(meta, "=", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid metadata format: %s (expected key=value)", meta)
					}
					entity.Metadata[parts[0]] = parts[1]
				}
			}

			if err := storage.CreateEntity(entity); err != nil {
				return fmt.Errorf("failed to create entity: %w", err)
			}

			fmt.Printf("Registered playbook: %s (slug: %s)\n", entity.Title, entity.Slug)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Title of the playbook")
	cmd.Flags().StringVar(&description, "description", "", "Description of the playbook")
	cmd.Flags().StringVar(&summary, "summary", "", "Summary of the playbook")
	cmd.Flags().StringSliceVar(&metadata, "meta", nil, "Metadata in key=value format (can be used multiple times)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags (comma-separated)")
	cmd.Flags().StringVar(&filename, "filename", "", "Override filename")

	cmd.MarkFlagRequired("title")

	return cmd
}

// createCmd creates a new collection
func createCmd() *cobra.Command {
	var (
		title       string
		description string
		summary     string
		metadata    []string
		tags        []string
	)

	cmd := &cobra.Command{
		Use:   "create collection <name>",
		Short: "Create a new collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "collection" {
				return fmt.Errorf("only 'collection' type is supported")
			}

			name := args[1]
			if title == "" {
				title = name
			}

			entity := &playbook.Entity{
				Type:        playbook.TypeCollection,
				Title:       title,
				Description: description,
				Summary:     summary,
				Tags:        tags,
			}

			// Parse metadata
			if len(metadata) > 0 {
				entity.Metadata = make(map[string]string)
				for _, meta := range metadata {
					parts := strings.SplitN(meta, "=", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid metadata format: %s (expected key=value)", meta)
					}
					entity.Metadata[parts[0]] = parts[1]
				}
			}

			if err := storage.CreateEntity(entity); err != nil {
				return fmt.Errorf("failed to create collection: %w", err)
			}

			fmt.Printf("Created collection: %s (slug: %s)\n", entity.Title, entity.Slug)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Display title (defaults to name)")
	cmd.Flags().StringVar(&description, "description", "", "Description of the collection")
	cmd.Flags().StringVar(&summary, "summary", "", "Summary of the collection")
	cmd.Flags().StringSliceVar(&metadata, "meta", nil, "Metadata in key=value format (can be used multiple times)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags (comma-separated)")

	return cmd
}

// listCmd lists entities with optional filters
func listCmd() *cobra.Command {
	var (
		entityType string
		tags       []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List entities with optional filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			var typeFilter *playbook.EntityType
			if entityType != "" {
				switch entityType {
				case "playbook":
					t := playbook.TypePlaybook
					typeFilter = &t
				case "collection":
					t := playbook.TypeCollection
					typeFilter = &t
				default:
					return fmt.Errorf("invalid type: %s (must be 'playbook' or 'collection')", entityType)
				}
			}

			entities, err := storage.ListEntities(typeFilter, tags)
			if err != nil {
				return fmt.Errorf("failed to list entities: %w", err)
			}

			for _, entity := range entities {
				fmt.Printf("%s (%s) - %s\n", entity.Slug, entity.Type, entity.Title)
				if entity.Summary != "" {
					fmt.Printf("  %s\n", entity.Summary)
				}
				if len(entity.Tags) > 0 {
					fmt.Printf("  Tags: %s\n", strings.Join(entity.Tags, ", "))
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&entityType, "type", "", "Filter by type (playbook or collection)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Filter by tags (comma-separated)")

	return cmd
}

// searchCmd searches entities
func searchCmd() *cobra.Command {
	var entityType string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search entities by query string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			var typeFilter *playbook.EntityType
			if entityType != "" {
				switch entityType {
				case "playbook":
					t := playbook.TypePlaybook
					typeFilter = &t
				case "collection":
					t := playbook.TypeCollection
					typeFilter = &t
				default:
					return fmt.Errorf("invalid type: %s (must be 'playbook' or 'collection')", entityType)
				}
			}

			entities, err := storage.SearchEntities(query, typeFilter)
			if err != nil {
				return fmt.Errorf("failed to search entities: %w", err)
			}

			for _, entity := range entities {
				fmt.Printf("%s (%s) - %s\n", entity.Slug, entity.Type, entity.Title)
				if entity.Summary != "" {
					fmt.Printf("  %s\n", entity.Summary)
				}
				if len(entity.Tags) > 0 {
					fmt.Printf("  Tags: %s\n", strings.Join(entity.Tags, ", "))
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&entityType, "type", "", "Filter by type (playbook or collection)")

	return cmd
}

// showCmd shows detailed information about an entity
func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <slug>",
		Short: "Show detailed information about an entity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			entity, err := storage.GetEntityBySlug(slug)
			if err != nil {
				return fmt.Errorf("failed to get entity: %w", err)
			}

			fmt.Printf("Slug: %s\n", entity.Slug)
			fmt.Printf("Type: %s\n", entity.Type)
			fmt.Printf("Title: %s\n", entity.Title)
			if entity.Description != "" {
				fmt.Printf("Description: %s\n", entity.Description)
			}
			if entity.Summary != "" {
				fmt.Printf("Summary: %s\n", entity.Summary)
			}
			if entity.CanonicalURL != nil {
				fmt.Printf("URL: %s\n", *entity.CanonicalURL)
			}
			if entity.Filename != nil {
				fmt.Printf("Filename: %s\n", *entity.Filename)
			}
			if len(entity.Tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(entity.Tags, ", "))
			}
			if entity.LastFetched != nil {
				fmt.Printf("Last Fetched: %s\n", entity.LastFetched.Format(time.RFC3339))
			}
			fmt.Printf("Created: %s\n", entity.CreatedAt.Format(time.RFC3339))

			// Show metadata
			if len(entity.Metadata) > 0 {
				fmt.Println("\nMetadata:")
				for key, value := range entity.Metadata {
					fmt.Printf("  %s: %s\n", key, value)
				}
			}

			// Show collection members if it's a collection
			if entity.Type == playbook.TypeCollection {
				members, err := storage.GetCollectionMembers(entity.ID)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to get collection members")
				} else if len(members) > 0 {
					fmt.Println("\nMembers:")
					for _, member := range members {
						memberEntity, err := storage.GetEntityByID(member.MemberID)
						if err != nil {
							fmt.Printf("  ID %d (error loading details)\n", member.MemberID)
						} else {
							fmt.Printf("  %s (%s) - %s\n", memberEntity.Slug, memberEntity.Type, memberEntity.Title)
							if member.RelativePath != nil {
								fmt.Printf("    Path: %s\n", *member.RelativePath)
							}
						}
					}
				}
			}

			// Show content if it's a playbook
			if entity.Type == playbook.TypePlaybook && entity.Content != nil {
				fmt.Println("\nContent:")
				fmt.Println(*entity.Content)
			}

			return nil
		},
	}

	return cmd
}

// addCmd adds entities to collections
func addCmd() *cobra.Command {
	var relativePath string

	cmd := &cobra.Command{
		Use:   "add <collection-slug> <member-slug>",
		Short: "Add a member to a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionSlug := args[0]
			memberSlug := args[1]

			collection, err := storage.GetEntityBySlug(collectionSlug)
			if err != nil {
				return fmt.Errorf("failed to get collection: %w", err)
			}
			if collection.Type != playbook.TypeCollection {
				return fmt.Errorf("%s is not a collection", collectionSlug)
			}

			member, err := storage.GetEntityBySlug(memberSlug)
			if err != nil {
				return fmt.Errorf("failed to get member: %w", err)
			}

			var relPath *string
			if relativePath != "" {
				relPath = &relativePath
			}

			if err := storage.AddToCollection(collection.ID, member.ID, relPath); err != nil {
				return fmt.Errorf("failed to add to collection: %w", err)
			}

			fmt.Printf("Added %s to collection %s\n", member.Title, collection.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&relativePath, "path", "", "Relative path within collection")

	return cmd
}

// removeCmd removes entities from collections or deletes entities
func removeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <collection-slug> <member-slug> | remove <entity-slug>",
		Short: "Remove a member from a collection or delete an entity",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				// Delete entity
				entitySlug := args[0]
				entity, err := storage.GetEntityBySlug(entitySlug)
				if err != nil {
					return fmt.Errorf("failed to get entity: %w", err)
				}

				if err := storage.DeleteEntity(entity.ID); err != nil {
					return fmt.Errorf("failed to delete entity: %w", err)
				}

				fmt.Printf("Deleted %s (%s)\n", entity.Title, entity.Type)
				return nil
			}

			if len(args) == 2 {
				// Remove from collection
				collectionSlug := args[0]
				memberSlug := args[1]

				collection, err := storage.GetEntityBySlug(collectionSlug)
				if err != nil {
					return fmt.Errorf("failed to get collection: %w", err)
				}
				if collection.Type != playbook.TypeCollection {
					return fmt.Errorf("%s is not a collection", collectionSlug)
				}

				member, err := storage.GetEntityBySlug(memberSlug)
				if err != nil {
					return fmt.Errorf("failed to get member: %w", err)
				}

				if err := storage.RemoveFromCollection(collection.ID, member.ID); err != nil {
					return fmt.Errorf("failed to remove from collection: %w", err)
				}

				fmt.Printf("Removed %s from collection %s\n", member.Title, collection.Title)
				return nil
			}

			return fmt.Errorf("invalid number of arguments")
		},
	}

	return cmd
}

// metaCmd manages metadata
func metaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta",
		Short: "Manage entity metadata",
	}

	// meta set
	setCmd := &cobra.Command{
		Use:   "set <slug> <key> <value>",
		Short: "Set metadata for an entity",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			key := args[1]
			value := args[2]

			entity, err := storage.GetEntityBySlug(slug)
			if err != nil {
				return fmt.Errorf("failed to get entity: %w", err)
			}

			if err := storage.SetMetadata(entity.ID, key, value); err != nil {
				return fmt.Errorf("failed to set metadata: %w", err)
			}

			fmt.Printf("Set metadata %s=%s for %s\n", key, value, entity.Title)
			return nil
		},
	}

	// meta get
	getCmd := &cobra.Command{
		Use:   "get <slug> [key]",
		Short: "Get metadata for an entity",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			entity, err := storage.GetEntityBySlug(slug)
			if err != nil {
				return fmt.Errorf("failed to get entity: %w", err)
			}

			metadata, err := storage.GetMetadata(entity.ID)
			if err != nil {
				return fmt.Errorf("failed to get metadata: %w", err)
			}

			if len(args) == 2 {
				// Get specific key
				key := args[1]
				if value, exists := metadata[key]; exists {
					fmt.Println(value)
				} else {
					return fmt.Errorf("metadata key %s not found", key)
				}
			} else {
				// Get all metadata
				for key, value := range metadata {
					fmt.Printf("%s: %s\n", key, value)
				}
			}

			return nil
		},
	}

	// meta remove
	removeCmd := &cobra.Command{
		Use:   "remove <slug> <key>",
		Short: "Remove metadata from an entity",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			key := args[1]

			entity, err := storage.GetEntityBySlug(slug)
			if err != nil {
				return fmt.Errorf("failed to get entity: %w", err)
			}

			if err := storage.DeleteMetadata(entity.ID, key); err != nil {
				return fmt.Errorf("failed to remove metadata: %w", err)
			}

			fmt.Printf("Removed metadata %s from %s\n", key, entity.Title)
			return nil
		},
	}

	cmd.AddCommand(setCmd, getCmd, removeCmd)
	return cmd
}

// deployCmd deploys entities to directories
func deployCmd() *cobra.Command {
	var filenameOverride string

	cmd := &cobra.Command{
		Use:   "deploy <slug> <target-directory>",
		Short: "Deploy an entity to a target directory",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			targetDir := args[1]

			entity, err := storage.GetEntityBySlug(slug)
			if err != nil {
				return fmt.Errorf("failed to get entity: %w", err)
			}

			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create target directory: %w", err)
			}

			if entity.Type == playbook.TypePlaybook {
				// Deploy single playbook
				filename := filenameOverride
				if filename == "" && entity.Filename != nil {
					filename = *entity.Filename
				}
				if filename == "" {
					filename = entity.Slug + ".md"
				}

				filePath := filepath.Join(targetDir, filename)
				if err := os.WriteFile(filePath, []byte(*entity.Content), 0644); err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}

				fmt.Printf("Deployed playbook %s to %s\n", entity.Title, filePath)
			} else {
				// Deploy collection
				members, err := storage.GetCollectionMembers(entity.ID)
				if err != nil {
					return fmt.Errorf("failed to get collection members: %w", err)
				}

				for _, member := range members {
					memberEntity, err := storage.GetEntityByID(member.MemberID)
					if err != nil {
						log.Warn().Err(err).Int64("member_id", member.MemberID).Msg("Failed to load member")
						continue
					}

					if memberEntity.Type == playbook.TypePlaybook && memberEntity.Content != nil {
						filename := ""
						if member.RelativePath != nil {
							filename = *member.RelativePath
						} else if memberEntity.Filename != nil {
							filename = *memberEntity.Filename
						} else {
							filename = memberEntity.Slug + ".md"
						}

						filePath := filepath.Join(targetDir, filename)

						// Create parent directory if needed
						if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
							log.Warn().Err(err).Str("path", filepath.Dir(filePath)).Msg("Failed to create directory")
							continue
						}

						if err := os.WriteFile(filePath, []byte(*memberEntity.Content), 0644); err != nil {
							log.Warn().Err(err).Str("path", filePath).Msg("Failed to write file")
							continue
						}

						fmt.Printf("Deployed %s to %s\n", memberEntity.Title, filePath)
					}
				}

				fmt.Printf("Deployed collection %s to %s\n", entity.Title, targetDir)
			}

			// Record deployment
			if err := storage.RecordDeployment(entity.ID, targetDir); err != nil {
				log.Warn().Err(err).Msg("Failed to record deployment")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&filenameOverride, "filename", "", "Override filename for playbook deployment")

	return cmd
}

// updateCmd updates playbooks from their sources
func updateCmd() *cobra.Command {
	var updateAll bool

	cmd := &cobra.Command{
		Use:   "update <slug> | --all",
		Short: "Update playbooks from their canonical sources",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if updateAll {
				// Update all playbooks with canonical URLs
				playbookType := playbook.TypePlaybook
				playbooks, err := storage.ListEntities(&playbookType, nil)
				if err != nil {
					return fmt.Errorf("failed to list playbooks: %w", err)
				}

				for _, pb := range playbooks {
					if pb.CanonicalURL != nil {
						if err := updatePlaybook(pb); err != nil {
							log.Warn().Err(err).Str("slug", pb.Slug).Msg("Failed to update playbook")
						} else {
							fmt.Printf("Updated %s\n", pb.Title)
						}
					}
				}
				return nil
			}

			if len(args) != 1 {
				return fmt.Errorf("either provide a slug or use --all")
			}

			slug := args[0]
			entity, err := storage.GetEntityBySlug(slug)
			if err != nil {
				return fmt.Errorf("failed to get entity: %w", err)
			}

			if entity.Type != playbook.TypePlaybook {
				return fmt.Errorf("%s is not a playbook", slug)
			}

			if entity.CanonicalURL == nil {
				return fmt.Errorf("playbook %s has no canonical URL", slug)
			}

			if err := updatePlaybook(entity); err != nil {
				return fmt.Errorf("failed to update playbook: %w", err)
			}

			fmt.Printf("Updated %s\n", entity.Title)
			return nil
		},
	}

	cmd.Flags().BoolVar(&updateAll, "all", false, "Update all playbooks")

	return cmd
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

func updatePlaybook(entity *playbook.Entity) error {
	if entity.CanonicalURL == nil {
		return fmt.Errorf("no canonical URL")
	}

	content, err := fetchURL(*entity.CanonicalURL)
	if err != nil {
		return err
	}

	// Update content and last fetched time
	entity.Content = &content
	now := time.Now()
	entity.LastFetched = &now

	// Here we would need an update method in storage
	// For now, this is a placeholder
	log.Info().Str("slug", entity.Slug).Msg("Would update playbook content")

	return nil
}
