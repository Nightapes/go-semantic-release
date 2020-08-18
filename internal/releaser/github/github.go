package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Nightapes/go-semantic-release/internal/assets"
	"github.com/Nightapes/go-semantic-release/internal/releaser/util"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"

	"github.com/google/go-github/v25/github"
	log "github.com/sirupsen/logrus"
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
	log     *log.Entry
}

// New initialize a new GitHubRelease
func New(c *config.GitHubProvider, checkConfig bool) (*Client, error) {

	token, err := util.GetAccessToken("GITHUB_TOKEN")
	if err != nil && checkConfig {
		return &Client{}, err
	}
	c.AccessToken = token
	ctx := context.Background()
	httpClient := util.CreateBearerHTTPClient(ctx, c.AccessToken)

	var client *github.Client
	baseURL := "https://github.com"

	if c.Repo == "" && checkConfig {
		return nil, fmt.Errorf("github repro is not set")
	}

	if c.User == "" && checkConfig {
		return nil, fmt.Errorf("github user is not set")
	}

	if c.CustomURL == "" {
		client = github.NewClient(httpClient)
	} else {
		if client, err = github.NewEnterpriseClient(c.CustomURL, c.CustomURL+"/api/v3/", httpClient); err != nil {
			return &Client{}, err
		}
		baseURL = c.CustomURL
	}
	return &Client{
		config:  c,
		client:  client,
		context: ctx,
		baseURL: baseURL,
		log:     log.WithField("releaser", GITHUB),
	}, nil
}

//GetCommitURL for github
func (g *Client) GetCommitURL() string {
	return fmt.Sprintf("%s/%s/%s/commit/{{hash}}", g.baseURL, g.config.User, g.config.Repo)
}

//GetCompareURL for github
func (g *Client) GetCompareURL(oldVersion, newVersion string) string {
	return fmt.Sprintf("%s/%s/%s/compare/%s...%s", g.baseURL, g.config.User, g.config.Repo, oldVersion, newVersion)
}

// CreateRelease creates release on remote
func (g *Client) CreateRelease(releaseVersion *shared.ReleaseVersion, generatedChangelog *shared.GeneratedChangelog, assets *assets.Set) error {
	err := g.makeRelease(releaseVersion, generatedChangelog)
	if err != nil {
		return err
	}
	return g.uploadAssets(assets)
}

// CreateRelease creates release on remote
func (g *Client) makeRelease(releaseVersion *shared.ReleaseVersion, generatedChangelog *shared.GeneratedChangelog) error {

	tag := "v" + releaseVersion.Next.Version.String()
	g.log.Debugf("create release with version %s", tag)

	prerelease := releaseVersion.Next.Version.Prerelease() != ""

	release, _, err := g.client.Repositories.CreateRelease(g.context, g.config.User, g.config.Repo, &github.RepositoryRelease{
		TagName:         &tag,
		TargetCommitish: &releaseVersion.Branch,
		Name:            &generatedChangelog.Title,
		Body:            &generatedChangelog.Content,
		Prerelease:      &prerelease,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already_exists") {
			g.log.Infof("A release with tag %s already exits, will not perform a release or update", tag)
			return nil
		}
		return fmt.Errorf("could not create release: %s", err.Error())
	}
	g.release = release
	g.log.Debugf("Release response: %+v", *release)
	g.log.Infof("Created release")
	return nil

}

// UploadAssets uploads specified assets
func (g *Client) uploadAssets(assets *assets.Set) error {
	if g.release != nil {
		for _, asset := range assets.All() {
			path, err := asset.GetPath()
			if err != nil {
				return err
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}

			_, resp, err := g.client.Repositories.UploadReleaseAsset(g.context, g.config.User, g.config.Repo, g.release.GetID(), &github.UploadOptions{Name: asset.GetName()}, file)
			if err != nil {
				return err
			}

			if resp.StatusCode >= http.StatusBadRequest {
				return fmt.Errorf("could not upload asset %s: %s", asset.GetName(), resp.Status)
			}
		}
	}
	return nil
}
