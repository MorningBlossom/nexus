package github

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	gh "github.com/google/go-github/v68/github"
)

type ReleaseArtifacts struct {
	Repo     string    `json:"repo"`
	Releases []Release `json:"releases"`
}

type Release struct {
	Tag           string        `json:"tag"`
	CommitMessage string        `json:"commitMessage"`
	CommitSHA     string        `json:"commitSha"`
	Packages      []RepoPackage `json:"packages"`
	CreatedAt     time.Time     `json:"-"`
	CommitTime    time.Time     `json:"-"`
}

type RepoPackage struct {
	PackageName string `json:"packageName"`
}

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

const packageTypeContainer = "container"

func (r *Repository) GetServicesByRepo(
	ctx context.Context,
	owner string,
	repo string,
) ([]string, error) {
	packages, res, err := r.api.Organizations.ListPackages(
		ctx,
		owner,
		&gh.PackageListOptions{
			PackageType: new(packageTypeContainer),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list packages for organization %q: %w", owner, err)
	}

	if res == nil || res.StatusCode != http.StatusOK {
		if res == nil {
			return nil, fmt.Errorf("list packages for organization %q: empty response", owner)
		}

		return nil, fmt.Errorf(
			"list packages for organization %q: unexpected status code %d",
			owner,
			res.StatusCode,
		)
	}

	services := make([]string, 0, len(packages))

	for _, pkg := range packages {
		if pkg == nil {
			continue
		}

		repository := pkg.GetRepository()
		if repository == nil || repository.GetName() != repo {
			continue
		}

		services = append(services, pkg.GetName())
	}

	// GitHub can return an empty result from the organization package-list
	// endpoint even when a package is directly accessible. Fall back to the
	// repository name as a package name so GHCR repositories such as
	// ghcr.io/MorningBlossom/auth-service are still resolved.
	if len(services) == 0 {
		pkg, pkgRes, pkgErr := r.api.Organizations.GetPackage(
			ctx,
			owner,
			packageTypeContainer,
			repo,
		)
		if pkgErr != nil {
			if pkgRes != nil {
				return nil, fmt.Errorf(
					"get package %q for organization %q: %s: %w",
					repo,
					owner,
					pkgRes.Status,
					pkgErr,
				)
			}
			return nil, fmt.Errorf(
				"get package %q for organization %q: %w",
				repo,
				owner,
				pkgErr,
			)
		}

		if pkg != nil {
			services = append(services, pkg.GetName())
		}
	}

	return services, nil
}

func (r *Repository) GetPackagesByName(
	ctx context.Context,
	owner string,
	repoName string,
) (ReleaseArtifacts, error) {
	services, err := r.GetServicesByRepo(ctx, owner, repoName)
	if err != nil {
		return ReleaseArtifacts{}, err
	}

	result := ReleaseArtifacts{
		Repo: repoName,
	}

	for _, service := range services {
		versions, res, err := r.api.Organizations.PackageGetAllVersions(
			ctx,
			owner,
			packageTypeContainer,
			service,
			nil,
		)
		if err != nil {
			return ReleaseArtifacts{}, fmt.Errorf(
				"get versions for package %q: %w",
				service,
				err,
			)
		}

		if res == nil || res.StatusCode != http.StatusOK {
			if res == nil {
				return ReleaseArtifacts{}, fmt.Errorf(
					"get versions for package %q: empty response",
					service,
				)
			}

			return ReleaseArtifacts{}, fmt.Errorf(
				"get versions for package %q: unexpected status code %d",
				service,
				res.StatusCode,
			)
		}

		// Tags are not required to match between services. Pick the newest
		// tagged package version independently for each service.
		var latestVersion *gh.PackageVersion
		for _, version := range versions {
			if version == nil || version.Metadata == nil || version.Metadata.Container == nil {
				continue
			}
			if len(version.Metadata.Container.Tags) == 0 {
				continue
			}
			if latestVersion == nil || version.CreatedAt.Time.After(latestVersion.CreatedAt.Time) {
				latestVersion = version
			}
		}

		if latestVersion == nil {
			continue
		}

		for _, tag := range latestVersion.Metadata.Container.Tags {
			// The mutable "latest" tag is not a release identity. Keep the
			// immutable version/commit tags as the source for artifact ordering.
			if strings.EqualFold(tag, "latest") {
				continue
			}
			result.addPackageToRelease(tag, service, latestVersion.CreatedAt.Time)
		}
	}

	// Resolve each image tag to its Git commit so the UI can display the
	// commit subject and SHA rather than the container tag.
	for i := range result.Releases {
		release := &result.Releases[i]
		commit, _, commitErr := r.api.Repositories.GetCommit(ctx, owner, repoName, release.Tag, nil)
		if commitErr != nil || commit == nil || commit.GetCommit() == nil {
			// Keep the image tag visible, but don't misrepresent it as a commit SHA.
			continue
		}

		release.CommitMessage = strings.SplitN(commit.GetCommit().GetMessage(), "\n", 2)[0]
		// go-github returns a github.Timestamp here; use its embedded time.Time.
		// The committer metadata can be absent, so guard it before reading the date.
		if committer := commit.GetCommit().GetCommitter(); committer != nil {
			release.CommitTime = committer.GetDate().Time
		}
		commitSHA := commit.GetSHA()
		if len(commitSHA) > 7 {
			commitSHA = commitSHA[:7]
		}
		release.CommitSHA = commitSHA
	}

	// Prefer source commit time when available. If a tag cannot be resolved
	// to a commit, fall back to the package's creation time so it is still
	// ordered newest-first rather than being pushed behind every resolved tag.
	sort.SliceStable(result.Releases, func(i, j int) bool {
		left, right := result.Releases[i], result.Releases[j]
		leftTime, rightTime := left.CommitTime, right.CommitTime
		if leftTime.IsZero() {
			leftTime = left.CreatedAt
		}
		if rightTime.IsZero() {
			rightTime = right.CreatedAt
		}
		return leftTime.After(rightTime)
	})

	return result, nil
}

func (r *ReleaseArtifacts) addPackageToRelease(tag, packageName string, createdAt time.Time) {
	for i := range r.Releases {
		if r.Releases[i].Tag != tag {
			continue
		}

		r.Releases[i].Packages = append(
			r.Releases[i].Packages,
			RepoPackage{
				PackageName: packageName,
			},
		)
		if createdAt.After(r.Releases[i].CreatedAt) {
			r.Releases[i].CreatedAt = createdAt
		}
		return
	}

	r.Releases = append(r.Releases, Release{
		Tag:       tag,
		Packages: []RepoPackage{
			{
				PackageName: packageName,
			},
		},
		CreatedAt: createdAt,
	})
}

func (r *Repository) GetPackagesByPackageName(ctx context.Context, owner string, packageName string) ([]*gh.PackageVersion, error) {
	packages, res, err := r.api.Organizations.PackageGetAllVersions(
		ctx,
		owner,
		"container",
		packageName,
		nil,
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

	return packages, nil
}

func (r *Repository) GetPackageByVersion(
	ctx context.Context,
	owner string,
	packageName string,
) (*gh.PackageVersion, error) {
	packages, res, err := r.api.Organizations.PackageGetVersion(
		ctx,
		owner,
		"container",
		packageName,
		1330274475,
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

	return packages, nil
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
