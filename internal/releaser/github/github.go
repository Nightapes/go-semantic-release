package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Nightapes/go-semantic-release/internal/releaser/util"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"

	"github.com/google/go-github/v25/github"
)

// GITHUB identifer for github interface
const GITHUB = "github"

// Client type struct
type Client struct {
	config  *config.GitHubProvider
	client  *github.Client
	context context.Context
	release *github.RepositoryRelease
	baseURL string
	token   string
}

// New initialize a new GitHubRelease
func New(c *config.GitHubProvider) (*Client, error) {
	ctx := context.Background()
	httpClient := util.CreateBearerHTTPClient(ctx, c.AccessToken)

	var client *github.Client
	var err error
	baseURL := "https://github.com"
	if c.CustomURL == "" {
		client = github.NewClient(httpClient)
	} else {
		client, err = github.NewEnterpriseClient(c.CustomURL, c.CustomURL+"/api/v3/", httpClient)
		baseURL = c.CustomURL
	}
	return &Client{
		config:  c,
		client:  client,
		context: ctx,
		baseURL: baseURL,
	}, err
}

//GetCommitURL for github
func (g Client) GetCommitURL() string {
	return fmt.Sprintf("%s/%s/%s/commit/{{hash}}", g.baseURL, g.config.User, g.config.Repo)
}

//GetCompareURL for github
func (g Client) GetCompareURL(oldVersion, newVersion string) string {
	return fmt.Sprintf("%s/%s/%s/compare/%s...%s", g.baseURL, g.config.User, g.config.Repo, oldVersion, newVersion)
}

//ValidateConfig for github
func (g Client) ValidateConfig() error {

	if g.config.Repo == "" {
		return fmt.Errorf("github Repro is not set")
	}

	if g.config.User == "" {
		return fmt.Errorf("github User is not set")
	}

	envName := fmt.Sprintf("%s_ACCESS_TOKEN", strings.ToUpper(GITHUB))
	token, isSet := os.LookupEnv(envName)
	if !isSet {
		return fmt.Errorf("can not find environment variable %s", envName)
	}
	g.token = token
	return nil

}

// CreateRelease creates release on remote
func (g Client) CreateRelease(releaseVersion *shared.ReleaseVersion, generatedChangelog *shared.GeneratedChangelog) error {

	tag := releaseVersion.Next.Version.String()
	prerelease := releaseVersion.Next.Version.Prerelease() != ""
	release, resp, err := g.client.Repositories.CreateRelease(g.context, g.config.User, g.config.Repo, &github.RepositoryRelease{
		TagName:         &tag,
		TargetCommitish: &releaseVersion.Branch,
		Name:            &generatedChangelog.Title,
		Body:            &generatedChangelog.Content,
		Draft:           &prerelease,
		Prerelease:      &prerelease,
	})

	if err != nil {
		return fmt.Errorf("could not create release: %v", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("could not create release: response statuscode: %s", resp.Status)
	}

	g.release = release
	return nil

}

// UploadAssets uploads specified assets
func (g Client) UploadAssets(assets []config.Asset) error {
	for _, asset := range assets {
		file, err := os.Open(asset.Name)
		if err != nil {
			return err
		}

		_, resp, err := g.client.Repositories.UploadReleaseAsset(g.context, g.config.User, g.config.Repo, *g.release.ID, &github.UploadOptions{Name: asset.Name}, file)
		if err != nil {
			return err
		}

		if resp.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf("releaser: github: Could not create release: response statuscode: %s", resp.Status)
		}
	}
	return nil
}
