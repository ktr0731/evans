package update_checker

import (
	"context"

	"github.com/google/go-github/github"
)

const (
	owner = "ktr0731"
	repo  = "evans"
)

type client interface {
	FetchLatestTag(context.Context) (string, error)
}

type gitHubClient struct {
	client *github.Client
}

func newGitHubClient() client {
	return &gitHubClient{
		client: github.NewClient(nil),
	}
}

func (c *gitHubClient) FetchLatestTag(ctx context.Context) (string, error) {
	r, _, err := c.client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return r.GetTagName(), nil
}
