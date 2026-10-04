package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	gh "github.com/google/go-github/v68/github"
)

type Client struct {
	AppID          int64
	InstallationID int64
	HTTPClient     *http.Client
	API            *gh.Client
}

func NewClient() (*Client, error) {
	ctx := context.Background()
	appID, err := strconv.ParseInt(
		os.Getenv("GITHUB_APP_ID"), 10, 64,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub App ID: %w", err)
	}

	installationID, err := strconv.ParseInt(
		os.Getenv("GITHUB_INSTALLATION_ID"), 10, 64,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid installation ID: %w", err)
	}

	privateKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")

	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	transport, err := ghinstallation.New(
		http.DefaultTransport,
		appID,
		installationID,
		privateKey,
	)

	if err != nil {
		return nil, fmt.Errorf("creating GitHub transport: %w", err)
	}

	tokenJwt, err := transport.Token(ctx)

	if err != nil {
		return nil, fmt.Errorf("getting token JWT: %w", err)
	}
	fmt.Println(tokenJwt)

	httpClient := &http.Client{
		Transport: transport,
	}

	return &Client{
		AppID:          appID,
		InstallationID: installationID,
		HTTPClient:     httpClient,
		API:            gh.NewClient(httpClient),
	}, nil
}

func (c *Client) GetRepository(
	ctx context.Context,
	owner string,
	repo string,
) (*gh.Repository, error) {
	repository, _, err := c.API.Repositories.Get(
		ctx,
		owner,
		repo,
	)

	if err != nil {
		return nil, fmt.Errorf("getting repository: %w", err)
	}

	return repository, nil
}
