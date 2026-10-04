package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v68/github"
)

type Repository struct {
	api *gh.Client
}

func NewRepository(api *gh.Client) *Repository {
	return &Repository{api: api}
}

func (r *Repository) GetRepository(ctx context.Context, owner string, repo string) (*gh.Repository, error) {
	repos, err, _ := r.api.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("getting repository: %w", err)
	}
	return repos, nil
}

func (r *Repository) GetPackages(
	ctx context.Context,
	owner string,
	packageName string,
) ([]*gh.Package, error) {
	packages, res, err := r.api.Organizations.GetPackage(
		ctx,
		owner,
		"container",
		packageName,
	)
	fmt.Println(packages)
	if err != nil {
		if res != nil {
			return nil, fmt.Errorf(
				"listing package %q for organization: %s: %w",
				packageName,
				res.Status,
				err,
			)
		}

		return nil, fmt.Errorf("listing package %q: %w", packageName, err)
	}

	return []*gh.Package{packages}, nil
}

func (r *Repository) GetInstallationRepositories(
	ctx context.Context,
) error {
	repos, _, err := r.api.Apps.ListRepos(
		ctx,
		&gh.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	)
	if err != nil {
		return fmt.Errorf("listing installation repositories: %w", err)
	}

	for _, repo := range repos.Repositories {
		fmt.Printf(
			"installation repo: %s/%s\n",
			repo.GetOwner().GetLogin(),
			repo.GetName(),
		)
	}

	return nil
}
