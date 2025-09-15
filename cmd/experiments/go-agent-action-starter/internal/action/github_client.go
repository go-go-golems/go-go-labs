package action

import (
	"context"
	"net/http"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

// GitHubClient bundles the go-github REST client and exposes the methods the
// action needs. It satisfies the PullRequestClient and ReviewPublisher
// interfaces.
type GitHubClient struct {
	client *github.Client
}

// NewGitHubClient constructs a GitHub client authenticated with the provided
// token. When token is empty the client still works for public repos but may be
// heavily rate limited.
func NewGitHubClient(ctx context.Context, token string) *GitHubClient {
	var httpClient *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient = oauth2.NewClient(ctx, ts)
	}
	return &GitHubClient{client: github.NewClient(httpClient)}
}

func (c *GitHubClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	return pr, err
}

func (c *GitHubClient) ListPullRequestFiles(ctx context.Context, owner, repo string, number int, opts *github.ListOptions) ([]*github.CommitFile, *github.Response, error) {
	return c.client.PullRequests.ListFiles(ctx, owner, repo, number, opts)
}

func (c *GitHubClient) CreateReview(ctx context.Context, owner, repo string, number int, req *github.PullRequestReviewRequest) (*github.PullRequestReview, *github.Response, error) {
	return c.client.PullRequests.CreateReview(ctx, owner, repo, number, req)
}

func (c *GitHubClient) CreateIssueComment(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	return c.client.Issues.CreateComment(ctx, owner, repo, number, comment)
}

// ReviewPublisher groups the GitHub methods used when publishing output.
type ReviewPublisher interface {
	CreateReview(ctx context.Context, owner, repo string, number int, req *github.PullRequestReviewRequest) (*github.PullRequestReview, *github.Response, error)
	CreateIssueComment(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
}

// Ensure GitHubClient satisfies the required interfaces.
var _ PullRequestClient = (*GitHubClient)(nil)
var _ ReviewPublisher = (*GitHubClient)(nil)
