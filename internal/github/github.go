package github

import (
	"context"

	"github.com/google/go-github/v62/github"
)

type Client struct {
	client *github.Client
}

func NewClient(token string) *Client {
	client := github.NewClient(nil).WithAuthToken(token)
	return &Client{client: client}
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, body, head, base string) (*github.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(head),
		Base:                github.String(base),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := c.client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, err
	}
	return pr, nil
}
