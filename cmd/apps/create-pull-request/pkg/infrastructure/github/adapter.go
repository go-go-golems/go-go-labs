package github

import (
	"context"
)

// Adapter defines the interface for GitHub CLI operations
type Adapter interface {
	// CreatePullRequest runs `gh pr create`
	CreatePullRequest(ctx context.Context, title string, body string) (string, error)
	// GetIssueDetails fetches basic details for a given issue ID
	GetIssueDetails(ctx context.Context, issueID string) (title string, body string, err error)
}

// MockAdapter provides a mock implementation for testing
type MockAdapter struct {
	MockPRURL     string
	MockIssueTitle string
	MockIssueBody string
	MockError     error
}

// CreatePullRequest implements the Adapter interface with mock data
func (m *MockAdapter) CreatePullRequest(ctx context.Context, title string, body string) (string, error) {
	if m.MockError != nil {
		return "", m.MockError
	}
	return m.MockPRURL, nil
}

// GetIssueDetails implements the Adapter interface with mock data
func (m *MockAdapter) GetIssueDetails(ctx context.Context, issueID string) (string, string, error) {
	if m.MockError != nil {
		return "", "", m.MockError
	}
	return m.MockIssueTitle, m.MockIssueBody, nil
}

// NewMockAdapter creates a new mock adapter with sample data
func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		MockPRURL:     "https://github.com/example/repo/pull/123",
		MockIssueTitle: "Example Issue Title",
		MockIssueBody:  "This is an example issue description with details about the problem.",
	}
}