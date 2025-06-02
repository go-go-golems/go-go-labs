package playbook

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Storage handles database operations for playbooks
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new storage instance and initializes the database
func NewStorage(dbPath string) (*Storage, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.initDB(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// initDB creates the database schema
func (s *Storage) initDB() error {
	schema := `
	-- Unified entities table (both playbooks and collections)
	CREATE TABLE IF NOT EXISTS entities (
		id INTEGER PRIMARY KEY,
		slug TEXT UNIQUE NOT NULL,
		type TEXT NOT NULL CHECK (type IN ('playbook', 'collection')),
		title TEXT NOT NULL,
		description TEXT,
		summary TEXT,
		canonical_url TEXT,
		content TEXT,
		command TEXT,
		content_hash TEXT,
		filename TEXT,
		tags TEXT DEFAULT '[]',
		last_fetched DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS entity_metadata (
		entity_id INTEGER REFERENCES entities(id) ON DELETE CASCADE,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		PRIMARY KEY (entity_id, key)
	);

	-- Collection membership (what playbooks/collections are in a collection)
	CREATE TABLE IF NOT EXISTS collection_members (
		collection_id INTEGER REFERENCES entities(id) ON DELETE CASCADE,
		member_id INTEGER REFERENCES entities(id) ON DELETE CASCADE,
		relative_path TEXT,
		PRIMARY KEY (collection_id, member_id)
	);

	CREATE TABLE IF NOT EXISTS deployments (
		id INTEGER PRIMARY KEY,
		entity_id INTEGER REFERENCES entities(id),
		target_directory TEXT,
		deployed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_entities_slug ON entities(slug);
	CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(type);
	CREATE INDEX IF NOT EXISTS idx_entities_created_at ON entities(created_at);
	CREATE INDEX IF NOT EXISTS idx_collection_members_collection_id ON collection_members(collection_id);
	CREATE INDEX IF NOT EXISTS idx_collection_members_member_id ON collection_members(member_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

// CreateEntity creates a new entity (playbook or collection)
func (s *Storage) CreateEntity(entity *Entity) error {
	// Generate slug from title
	entity.Slug = GenerateSlug(entity.Title)
	
	// Handle slug conflicts by appending numbers
	originalSlug := entity.Slug
	counter := 1
	for {
		if !s.slugExists(entity.Slug) {
			break
		}
		entity.Slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	// Calculate content hash for playbooks
	if entity.Type == TypePlaybook {
		var hashContent string
		if entity.Content != nil {
			hashContent = *entity.Content
		} else if entity.Command != nil {
			hashContent = *entity.Command
		}
		if hashContent != "" {
			hash := sha256.Sum256([]byte(hashContent))
			hashStr := hex.EncodeToString(hash[:])
			entity.ContentHash = &hashStr
		}
	}

	// Marshal tags
	tagsJSON, err := entity.MarshalTags()
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	// Insert entity
	result, err := s.db.Exec(`
		INSERT INTO entities (slug, type, title, description, summary, canonical_url, content, command, content_hash, filename, tags, last_fetched)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entity.Slug, entity.Type, entity.Title, entity.Description, entity.Summary,
		entity.CanonicalURL, entity.Content, entity.Command, entity.ContentHash, entity.Filename, tagsJSON, entity.LastFetched)

	if err != nil {
		return fmt.Errorf("failed to insert entity: %w", err)
	}

	entityID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get entity ID: %w", err)
	}
	entity.ID = entityID

	// Insert metadata if provided
	if entity.Metadata != nil {
		for key, value := range entity.Metadata {
			if err := s.SetMetadata(entity.ID, key, value); err != nil {
				return fmt.Errorf("failed to set metadata %s: %w", key, err)
			}
		}
	}

	return nil
}

// GetEntityBySlug retrieves an entity by its slug
func (s *Storage) GetEntityBySlug(slug string) (*Entity, error) {
	entity := &Entity{}
	var tagsJSON string

	err := s.db.QueryRow(`
		SELECT id, slug, type, title, description, summary, canonical_url, content, command, content_hash, filename, tags, last_fetched, created_at
		FROM entities WHERE slug = ?
	`, slug).Scan(&entity.ID, &entity.Slug, &entity.Type, &entity.Title, &entity.Description, &entity.Summary,
		&entity.CanonicalURL, &entity.Content, &entity.Command, &entity.ContentHash, &entity.Filename, &tagsJSON, &entity.LastFetched, &entity.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entity with slug %s not found", slug)
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	// Unmarshal tags
	if err := entity.UnmarshalTags(tagsJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	// Load metadata
	metadata, err := s.GetMetadata(entity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}
	entity.Metadata = metadata

	return entity, nil
}

// GetEntityByID retrieves an entity by its ID
func (s *Storage) GetEntityByID(id int64) (*Entity, error) {
	entity := &Entity{}
	var tagsJSON string

	err := s.db.QueryRow(`
		SELECT id, slug, type, title, description, summary, canonical_url, content, command, content_hash, filename, tags, last_fetched, created_at
		FROM entities WHERE id = ?
	`, id).Scan(&entity.ID, &entity.Slug, &entity.Type, &entity.Title, &entity.Description, &entity.Summary,
		&entity.CanonicalURL, &entity.Content, &entity.Command, &entity.ContentHash, &entity.Filename, &tagsJSON, &entity.LastFetched, &entity.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entity with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	// Unmarshal tags
	if err := entity.UnmarshalTags(tagsJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	// Load metadata
	metadata, err := s.GetMetadata(entity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}
	entity.Metadata = metadata

	return entity, nil
}

// ListEntities lists entities with optional filters
func (s *Storage) ListEntities(entityType *EntityType, tags []string) ([]*Entity, error) {
	query := `
		SELECT id, slug, type, title, description, summary, canonical_url, content, command, content_hash, filename, tags, last_fetched, created_at
		FROM entities WHERE 1=1
	`
	args := []interface{}{}

	if entityType != nil {
		query += " AND type = ?"
		args = append(args, *entityType)
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			query += " AND tags LIKE ?"
			args = append(args, "%\""+tag+"\"%")
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query entities: %w", err)
	}
	defer rows.Close()

	var entities []*Entity
	for rows.Next() {
		entity := &Entity{}
		var tagsJSON string

		err := rows.Scan(&entity.ID, &entity.Slug, &entity.Type, &entity.Title, &entity.Description, &entity.Summary,
			&entity.CanonicalURL, &entity.Content, &entity.Command, &entity.ContentHash, &entity.Filename, &tagsJSON, &entity.LastFetched, &entity.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entity: %w", err)
		}

		// Unmarshal tags
		if err := entity.UnmarshalTags(tagsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// SearchEntities searches entities by query string
func (s *Storage) SearchEntities(query string, entityType *EntityType) ([]*Entity, error) {
	searchQuery := `
		SELECT id, slug, type, title, description, summary, canonical_url, content, command, content_hash, filename, tags, last_fetched, created_at
		FROM entities 
		WHERE (title LIKE ? OR description LIKE ? OR summary LIKE ? OR content LIKE ? OR command LIKE ?)
	`
	args := []interface{}{
		"%" + query + "%",
		"%" + query + "%",
		"%" + query + "%",
		"%" + query + "%",
		"%" + query + "%",
	}

	if entityType != nil {
		searchQuery += " AND type = ?"
		args = append(args, *entityType)
	}

	searchQuery += " ORDER BY created_at DESC"

	rows, err := s.db.Query(searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}
	defer rows.Close()

	var entities []*Entity
	for rows.Next() {
		entity := &Entity{}
		var tagsJSON string

		err := rows.Scan(&entity.ID, &entity.Slug, &entity.Type, &entity.Title, &entity.Description, &entity.Summary,
			&entity.CanonicalURL, &entity.Content, &entity.Command, &entity.ContentHash, &entity.Filename, &tagsJSON, &entity.LastFetched, &entity.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entity: %w", err)
		}

		// Unmarshal tags
		if err := entity.UnmarshalTags(tagsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// SetMetadata sets metadata for an entity
func (s *Storage) SetMetadata(entityID int64, key, value string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO entity_metadata (entity_id, key, value)
		VALUES (?, ?, ?)
	`, entityID, key, value)
	return err
}

// GetMetadata gets all metadata for an entity
func (s *Storage) GetMetadata(entityID int64) (map[string]string, error) {
	rows, err := s.db.Query(`
		SELECT key, value FROM entity_metadata WHERE entity_id = ?
	`, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metadata := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		metadata[key] = value
	}

	return metadata, nil
}

// DeleteMetadata removes metadata for an entity
func (s *Storage) DeleteMetadata(entityID int64, key string) error {
	_, err := s.db.Exec(`
		DELETE FROM entity_metadata WHERE entity_id = ? AND key = ?
	`, entityID, key)
	return err
}

// AddToCollection adds an entity to a collection
func (s *Storage) AddToCollection(collectionID, memberID int64, relativePath *string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO collection_members (collection_id, member_id, relative_path)
		VALUES (?, ?, ?)
	`, collectionID, memberID, relativePath)
	return err
}

// RemoveFromCollection removes an entity from a collection
func (s *Storage) RemoveFromCollection(collectionID, memberID int64) error {
	_, err := s.db.Exec(`
		DELETE FROM collection_members WHERE collection_id = ? AND member_id = ?
	`, collectionID, memberID)
	return err
}

// GetCollectionMembers gets all members of a collection
func (s *Storage) GetCollectionMembers(collectionID int64) ([]*CollectionMember, error) {
	rows, err := s.db.Query(`
		SELECT collection_id, member_id, relative_path
		FROM collection_members WHERE collection_id = ?
	`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*CollectionMember
	for rows.Next() {
		member := &CollectionMember{}
		if err := rows.Scan(&member.CollectionID, &member.MemberID, &member.RelativePath); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

// DeleteEntity deletes an entity and all related data
func (s *Storage) DeleteEntity(id int64) error {
	_, err := s.db.Exec(`DELETE FROM entities WHERE id = ?`, id)
	return err
}

// RecordDeployment records a deployment
func (s *Storage) RecordDeployment(entityID int64, targetDirectory string) error {
	_, err := s.db.Exec(`
		INSERT INTO deployments (entity_id, target_directory)
		VALUES (?, ?)
	`, entityID, targetDirectory)
	return err
}

// GetDeployments gets deployments for an entity
func (s *Storage) GetDeployments(entityID int64) ([]*Deployment, error) {
	rows, err := s.db.Query(`
		SELECT id, entity_id, target_directory, deployed_at
		FROM deployments WHERE entity_id = ? ORDER BY deployed_at DESC
	`, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*Deployment
	for rows.Next() {
		deployment := &Deployment{}
		if err := rows.Scan(&deployment.ID, &deployment.EntityID, &deployment.TargetDirectory, &deployment.DeployedAt); err != nil {
			return nil, err
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// slugExists checks if a slug already exists
func (s *Storage) slugExists(slug string) bool {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM entities WHERE slug = ?", slug).Scan(&count)
	return count > 0
}
