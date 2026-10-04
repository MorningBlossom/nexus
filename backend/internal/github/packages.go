package github

//
//import (
//	"context"
//	"fmt"
//
//	"github.com/shurcooL/githubv4"
//)
//
//type Package struct {
//	Name        string
//	PackageType string
//}
//
//func (r *Repository) GetPackages(
//	ctx context.Context,
//	owner string,
//	repo string,
//) ([]*Package, error) {
//
//	if r.httpClient == nil {
//		return nil, fmt.Errorf("GitHub HTTP client is not configured")
//	}
//
//	client := githubv4.NewClient(r.httpClient)
//
//	var allPackages []*Package
//
//	variables := map[string]interface{}{
//		"owner":       githubv4.String(owner),
//		"repo":        githubv4.String(repo),
//		"packageType": githubv4.PackageType("CONTAINER"),
//		"after":       (*githubv4.String)(nil),
//	}
//
//	for {
//		var query struct {
//			Repository struct {
//				Packages struct {
//					Nodes []struct {
//						Name        githubv4.String
//						PackageType githubv4.String `graphql:"packageType"`
//					}
//					PageInfo struct {
//						HasNextPage githubv4.Boolean
//						EndCursor   githubv4.String
//					}
//				} `graphql:"packages(first: 100, after: $after, packageType: $packageType)"`
//			} `graphql:"repository(owner: $owner, name: $repo)"`
//		}
//
//		err := client.Query(ctx, &query, variables)
//		if err != nil {
//			return nil, fmt.Errorf(
//				"getting packages for %s/%s: %w",
//				owner,
//				repo,
//				err,
//			)
//		}
//
//		for _, pkg := range query.Repository.Packages.Nodes {
//			allPackages = append(allPackages, &Package{
//				Name:        string(pkg.Name),
//				PackageType: string(pkg.PackageType),
//			})
//		}
//
//		if !query.Repository.Packages.PageInfo.HasNextPage {
//			break
//		}
//
//		cursor := query.Repository.Packages.PageInfo.EndCursor
//		variables["after"] = &cursor
//	}
//
//	return allPackages, nil
//}
