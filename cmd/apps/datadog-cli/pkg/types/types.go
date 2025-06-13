package types

import (
	"time"
)

// DatadogQuery represents a Datadog Logs Search API query
type DatadogQuery struct {
	Query   string    `json:"query"`
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
	Limit   int       `json:"limit,omitempty"`
	Sort    string    `json:"sort,omitempty"`
	GroupBy []string  `json:"group_by,omitempty"`
	Aggs    []string  `json:"aggs,omitempty"`
}

// QueryMetadata contains additional query configuration
type QueryMetadata struct {
	GroupBy []string `yaml:"group_by,omitempty"`
	Sort    string   `yaml:"sort,omitempty"`
	Aggs    []string `yaml:"aggs,omitempty"`
}
