package git

import (
	"context"
)

// Adapter defines the interface for Git operations
type Adapter interface {
	// GetDiff fetches the git diff based on parameters
	GetDiff(ctx context.Context, branch string, exclusions []string, contextSize int, includePaths []string, noTests bool, noPackage bool) (string, error)
	// GetDiffStat fetches the git diff --stat
	GetDiffStat(ctx context.Context, branch string, exclusions []string, includePaths []string, noTests bool, noPackage bool) (string, error)
	// FetchOrigin runs `git fetch origin`
	FetchOrigin(ctx context.Context) error
	// GetDiffFromClipboard reads diff from system clipboard
	GetDiffFromClipboard(ctx context.Context) (string, error)
	// GetCommitHistory fetches commit messages from a specified range
	GetCommitHistory(ctx context.Context, baseBranch string, head string) ([]string, error)
}

// MockAdapter provides a mock implementation for testing
type MockAdapter struct {
	MockDiff           string
	MockDiffStat       string
	MockCommitMessages []string
	MockError          error
}

// GetDiff implements the Adapter interface with mock data
func (m *MockAdapter) GetDiff(ctx context.Context, branch string, exclusions []string, contextSize int, includePaths []string, noTests bool, noPackage bool) (string, error) {
	if m.MockError != nil {
		return "", m.MockError
	}
	return m.MockDiff, nil
}

// GetDiffStat implements the Adapter interface with mock data
func (m *MockAdapter) GetDiffStat(ctx context.Context, branch string, exclusions []string, includePaths []string, noTests bool, noPackage bool) (string, error) {
	if m.MockError != nil {
		return "", m.MockError
	}
	return m.MockDiffStat, nil
}

// FetchOrigin implements the Adapter interface
func (m *MockAdapter) FetchOrigin(ctx context.Context) error {
	return m.MockError
}

// GetDiffFromClipboard implements the Adapter interface with mock data
func (m *MockAdapter) GetDiffFromClipboard(ctx context.Context) (string, error) {
	if m.MockError != nil {
		return "", m.MockError
	}
	return m.MockDiff, nil
}

// GetCommitHistory implements the Adapter interface with mock data
func (m *MockAdapter) GetCommitHistory(ctx context.Context, baseBranch string, head string) ([]string, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockCommitMessages, nil
}

// NewMockAdapter creates a new mock adapter with sample data
func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		MockDiff: `diff --git a/sample.go b/sample.go
index 1234567..abcdef 100644
--- a/sample.go
+++ b/sample.go
@@ -1,5 +1,7 @@
 package main
 
+import "fmt"
+
 func main() {
-    // TODO: implement
+    fmt.Println("Hello, World!")
 }
`,
		MockDiffStat: ` sample.go | 4 ++--
 1 file changed, 2 insertions(+), 2 deletions(-)
`,
		MockCommitMessages: []string{
			"feat: implement hello world",
			"chore: initial commit",
		},
	}
}
