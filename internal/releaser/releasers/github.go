package releasers

import (
	"github.com/Nightapes/go-semantic-release/pkg/config"
)

// GITHUB identifer for github interface
const GITHUB = "github"

// GitHubReleaser type struct
type GitHubReleaser struct {
	RepositoryURL   string
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	ReleaseName     string `json:"name"`
	ReleaseMessage  string `json:"body"`
	Draft           bool   `json:"draft,omitempty"`
	Prerelease      bool   `json:"prerelease,omitempty"`
	// Assets
}

// CreateRelease creates release on remote
func (g *GitHubReleaser) CreateRelease(releaseName, releaseMessage string) error {
	g.ReleaseName = releaseName
	g.ReleaseMessage = releaseMessage
	return nil
}

// NewGitHubReleaser initialize a new GitHubRelease
func NewGitHubReleaser(c *config.ReleaseConfig, branch, latestVersion string) *GitHubReleaser {
	return &GitHubReleaser{
		RepositoryURL:   c.Github["url"],
		TargetCommitish: branch,
		TagName:         latestVersion,
	}
}
