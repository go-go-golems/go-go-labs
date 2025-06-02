package playbook

import (
	"encoding/json"
	"time"

	"github.com/gosimple/slug"
)

type EntityType string

const (
	TypePlaybook   EntityType = "playbook"
	TypeCollection EntityType = "collection"
)

// Entity represents both playbooks and collections
type Entity struct {
	ID           int64     `json:"id" db:"id"`
	Slug         string    `json:"slug" db:"slug"`
	Type         EntityType `json:"type" db:"type"`
	Title        string    `json:"title" db:"title"`
	Description  string    `json:"description" db:"description"`
	Summary      string    `json:"summary" db:"summary"`
	CanonicalURL *string   `json:"canonical_url,omitempty" db:"canonical_url"`
	Content      *string   `json:"content,omitempty" db:"content"`
	Command      *string   `json:"command,omitempty" db:"command"`
	ContentHash  *string   `json:"content_hash,omitempty" db:"content_hash"`
	Filename     *string   `json:"filename,omitempty" db:"filename"`
	Tags         []string  `json:"tags" db:"tags"`
	LastFetched  *time.Time `json:"last_fetched,omitempty" db:"last_fetched"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// EntityMetadata represents key-value metadata for entities
type EntityMetadata struct {
	EntityID int64  `json:"entity_id" db:"entity_id"`
	Key      string `json:"key" db:"key"`
	Value    string `json:"value" db:"value"`
}

// CollectionMember represents membership in a collection
type CollectionMember struct {
	CollectionID int64   `json:"collection_id" db:"collection_id"`
	MemberID     int64   `json:"member_id" db:"member_id"`
	RelativePath *string `json:"relative_path,omitempty" db:"relative_path"`
}

// Deployment represents a deployment of an entity to a directory
type Deployment struct {
	ID              int64     `json:"id" db:"id"`
	EntityID        int64     `json:"entity_id" db:"entity_id"`
	TargetDirectory string    `json:"target_directory" db:"target_directory"`
	DeployedAt      time.Time `json:"deployed_at" db:"deployed_at"`
}

// GenerateSlug creates a URL-friendly slug from the title
func GenerateSlug(title string) string {
	return slug.Make(title)
}

// MarshalTags converts tags slice to JSON string for database storage
func (e *Entity) MarshalTags() (string, error) {
	if len(e.Tags) == 0 {
		return "[]", nil
	}
	bytes, err := json.Marshal(e.Tags)
	return string(bytes), err
}

// UnmarshalTags converts JSON string from database to tags slice
func (e *Entity) UnmarshalTags(tagsJSON string) error {
	if tagsJSON == "" {
		e.Tags = []string{}
		return nil
	}
	return json.Unmarshal([]byte(tagsJSON), &e.Tags)
}

// HasTag checks if entity has a specific tag
func (e *Entity) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag if it doesn't already exist
func (e *Entity) AddTag(tag string) {
	if !e.HasTag(tag) {
		e.Tags = append(e.Tags, tag)
	}
}

// RemoveTag removes a tag if it exists
func (e *Entity) RemoveTag(tag string) {
	for i, t := range e.Tags {
		if t == tag {
			e.Tags = append(e.Tags[:i], e.Tags[i+1:]...)
			break
		}
	}
}

// IsCommand returns true if this entity represents a shell command
func (e *Entity) IsCommand() bool {
	return e.Command != nil && *e.Command != ""
}

// GetContentOrCommand returns the content for file-based playbooks or command for shell playbooks
func (e *Entity) GetContentOrCommand() string {
	if e.IsCommand() {
		return *e.Command
	}
	if e.Content != nil {
		return *e.Content
	}
	return ""
}
