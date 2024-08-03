package bee

import (
	"sort"
	"time"
)

// Conversation-related structs
type ConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
	CurrentPage   int            `json:"currentPage"`
	TotalPages    int            `json:"totalPages"`
	TotalCount    int            `json:"totalCount"`
}

type SuggestedLink struct {
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
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
	SuggestedLinks  []SuggestedLink `json:"suggested_links"`
}

func (c *Conversation) Transcript(withSpeaker bool) string {
	var ret string
	// sort transcripts by id
	sortedTranscripts := make([]Transcription, len(c.Transcriptions))
	copy(sortedTranscripts, c.Transcriptions)
	sort.Slice(sortedTranscripts, func(i, j int) bool {
		return sortedTranscripts[i].ID < sortedTranscripts[j].ID
	})

	for _, t := range sortedTranscripts {
		ret += t.ToString(withSpeaker)
		ret += "\n"
	}
	return ret

}

type Location struct {
	// Add location properties as needed
}

type Transcription struct {
	ID         int         `json:"id"`
	Realtime   bool        `json:"realtime"`
	Utterances []Utterance `json:"utterances"`
}

func (t *Transcription) ToString(withSpeaker bool) string {
	var ret string
	prevSpeaker := ""
	for _, u := range t.Utterances {
		if withSpeaker && prevSpeaker != u.Speaker {
			ret += "\n" + u.Speaker + ": "
			prevSpeaker = u.Speaker
		}
		ret += u.Text + " "
	}
	return ret
}

type Utterance struct {
	ID        int       `json:"id"`
	Realtime  bool      `json:"realtime"`
	Start     float64   `json:"start"`
	End       float64   `json:"end"`
	SpokenAt  time.Time `json:"spoken_at"`
	Text      string    `json:"text"`
	Speaker   string    `json:"speaker"`
	CreatedAt time.Time `json:"created_at"`
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
