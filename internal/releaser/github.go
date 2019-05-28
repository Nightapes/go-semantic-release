package releaser

import (
	"encoding/json"
	"fmt"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"io/ioutil"
	"strings"
)

// GITHUB identifer for github interface
const GITHUB = "github"
const githubCreateRleasURL = "https://api.github.com/repos"
const githubUploadAssetURL = "https://uploads.github.com"

// GitHubReleaser type struct
type GitHubReleaser struct {
	repositoryURL  string
	authToken      string
	assets         []config.Asset
	Version        string `json:"tag_name"`
	Branch         string `json:"target_commitish"`
	ReleaseName    string `json:"name"`
	ReleaseMessage string `json:"body"`
	Draft          bool   `json:"draft,omitempty"`
	Prerelease     bool   `json:"prerelease,omitempty"`
}

type gitHubCreateReleaseResponse struct {
	ReleaseURL     string `json:url`
	AssetUploadURL string `json:upload_url`
}

// NewGitHubReleaser initialize a new GitHubRelease
func NewGitHubReleaser(c *config.ReleaseConfig) *GitHubReleaser {
	return &GitHubReleaser{
		repositoryURL: c.Github["url"],
		authToken:     c.Github["authToken"],
		assets:        c.Assets,
	}
}

// CreateRelease creates release on remote
func (g *GitHubReleaser) CreateRelease(releaseName, releaseMessage, branch, version string) error {
	g.ReleaseName = releaseName
	g.ReleaseMessage = releaseMessage
	g.Branch = branch
	g.Version = version

	repositoryURI := strings.TrimLeft(g.repositoryURL, "/")
	jsonRelease, err := json.Marshal(g)

	if err != nil {
		return fmt.Errorf("releaser: github: could not marshal GitHubReleaser struct. Error: %s", err.Error())
	}

	tempDir, err := ioutil.TempDir(".", "tempZipAssets")
	if err != nil {
		return fmt.Errorf("releaser: github: Could not create a temp directory. Error: %s", err.Error())
	}
	assetList, err := prepareAssets(tempDir, g.assets)
	if err != nil {
		return err
	}

	response, err := makeReleaseRequest(githubCreateRleasURL+repositoryURI, g.authToken, jsonRelease)
	if err != nil {
		return err
	}

	releaseInfo := gitHubCreateReleaseResponse{}
	if err := json.Unmarshal(response, &releaseInfo); err != nil {
		return err
	}

	// tbd build new upload url
	if err := uploadReleaseAssets(releaseInfo.AssetUploadURL, g.authToken, assetList); err != nil {
		return err
	}

	return nil
}
