package github

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v68/github"
)

type Client struct {
	Client *github.Client
}

func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_PAT")

	if token == "" {
		return nil, fmt.Errorf("GITHUB_PAT is not set")
	}

	httpClient := &http.Client{}

	githubClient := github.NewClient(httpClient).WithAuthToken(token)

	return &Client{
		Client: githubClient,
	}, nil
}

func (c *Client) GetRepository(
	ctx context.Context,
	owner string,
	repo string,
) (*github.Repository, error) {

	repository, _, err := c.Client.Repositories.Get(
		ctx,
		owner,
		repo,
	)

	if err != nil {
		return nil, fmt.Errorf("get repository: %w", err)
	}

	return repository, nil
}
