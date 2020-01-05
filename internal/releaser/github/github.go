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
func (g *Client) CreateRelease(releaseVersion *shared.ReleaseVersion, generatedChangelog *shared.GeneratedChangelog) error {

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
	g.log.Debugf("Release repsone: %+v", *release)
	g.log.Infof("Crated release")
	return nil

}

// UploadAssets uploads specified assets
func (g *Client) UploadAssets(repoDir string, assets []config.Asset) error {
	if g.release != nil {
		filesToUpload, err := util.PrepareAssets(repoDir, assets)
		if err != nil {
			return err
		}
		for _, f := range filesToUpload {

			file, err := os.Open(*f)
			if err != nil {
				return err
			}
			fileInfo, _ := file.Stat()

			_, resp, err := g.client.Repositories.UploadReleaseAsset(g.context, g.config.User, g.config.Repo, g.release.GetID(), &github.UploadOptions{Name: fileInfo.Name()}, file)
			if err != nil {
				return err
			}

			if resp.StatusCode >= http.StatusBadRequest {
				return fmt.Errorf("could not upload asset %s: %s", file.Name(), resp.Status)
			}
		}
	}
	return nil
}
