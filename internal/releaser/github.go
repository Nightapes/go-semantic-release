package releaser

import (
	"context"
	"fmt"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/google/go-github/v25/github"
	"net/http"
	"os"
)

// GITHUB identifer for github interface
const GITHUB = "GitHub"

// GitHubReleaser type struct
type GitHubReleaser struct {
	config  *config.ReleaseConfig
	client  *github.Client
	context context.Context
	release *github.RepositoryRelease
}

type gitHubCreateReleaseResponse struct {
	ReleaseURL     string `json:url`
	AssetUploadURL string `json:upload_url`
}

// NewGitHubReleaser initialize a new GitHubRelease
func NewGitHubReleaser(c *config.ReleaseConfig) *GitHubReleaser {
	ctx := context.Background()
	httpClient := createHTTPClient(ctx, c.Github.AccessToken)

	return &GitHubReleaser{
		config:  c,
		client:  github.NewClient(httpClient),
		context: ctx,
	}
}

// CreateRelease creates release on remote
func (g GitHubReleaser) CreateRelease(tag, releaseName, releaseMessage, targetBranch string) error {

	release, resp, err := g.client.Repositories.CreateRelease(g.context, g.config.Github.User, g.config.Github.URL, &github.RepositoryRelease{
		TagName:         &tag,
		TargetCommitish: &targetBranch,
		Name:            &releaseName,
		Body:            &releaseMessage,
		Draft:           &g.config.IsDraft,
		Prerelease:      &g.config.IsPreRelease,
	})

	if err != nil {
		return fmt.Errorf("releaser: github: Could not create release: %v", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("releaser: github: Could not create release: response statuscode: %s", resp.Status)
	}

	g.release = release
	return nil

}

// UploadAssets uploads specified assets
func (g GitHubReleaser) UploadAssets(assets []config.Asset) error {
	for _, asset := range assets {
		file, err := os.Open(asset.Name)
		if err != nil {
			return err
		}

		_, resp, err := g.client.Repositories.UploadReleaseAsset(g.context, g.config.Github.User, g.config.Github.URL, *g.release.ID, &github.UploadOptions{Name: asset.Name}, file)
		if err != nil {
			return err
		}

		if resp.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf("releaser: github: Could not create release: response statuscode: %s", resp.Status)
		}
	}
	return nil
}
