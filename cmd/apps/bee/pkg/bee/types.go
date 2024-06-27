package bee

import "time"

// Conversation-related structs
type ConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
	CurrentPage   int            `json:"currentPage"`
	TotalPages    int            `json:"totalPages"`
	TotalCount    int            `json:"totalCount"`
}

type Conversation struct {
	ID              int             `json:"id"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         *time.Time      `json:"end_time"`
	DeviceType      string          `json:"device_type"`
	Summary         *string         `json:"summary"`
	ShortSummary    *string         `json:"short_summary"`
	State           string          `json:"state"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	PrimaryLocation *Location       `json:"primary_location"`
	Transcriptions  []Transcription `json:"transcriptions"`
	SuggestedLinks  []string        `json:"suggested_links"`
}

type Location struct {
	// Add location properties as needed
}

type Transcription struct {
	ID         int         `json:"id"`
	Realtime   bool        `json:"realtime"`
	Utterances []Utterance `json:"utterances"`
}

type Utterance struct {
	// Add utterance properties as needed
}

// Fact-related structs
type FactsResponse struct {
	Facts       []Fact `json:"facts"`
	CurrentPage int    `json:"currentPage"`
	TotalPages  int    `json:"totalPages"`
	TotalCount  int    `json:"totalCount"`
}

type Fact struct {
	ID             int       `json:"id"`
	Text           string    `json:"text"`
	Tags           []string  `json:"tags"`
	Visibility     string    `json:"visibility"`
	Confirmed      bool      `json:"confirmed"`
	UserID         int       `json:"user_id"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
	Topic          *string   `json:"topic"`
	Source         *string   `json:"source"`
	Score          *float64  `json:"score"`
	Embedding      []float64 `json:"embedding"`
	FTS            string    `json:"fts"`
	ConversationID *int      `json:"conversation_id"`
}

type FactInput struct {
	Text      string `json:"text"`
	Confirmed bool   `json:"confirmed"`
}

// Todo-related structs
type TodosResponse struct {
	Todos       []Todo `json:"todos"`
	CurrentPage int    `json:"currentPage"`
	TotalPages  int    `json:"totalPages"`
	TotalCount  int    `json:"totalCount"`
}

type Todo struct {
	ID        int        `json:"id"`
	Text      string     `json:"text"`
	AlarmAt   *time.Time `json:"alarm_at"`
	Completed bool       `json:"completed"`
	CreatedAt time.Time  `json:"created_at"`
}

type TodoInput struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}
