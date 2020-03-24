package releaser

import (
	"fmt"

	"github.com/Nightapes/go-semantic-release/internal/assets"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/internal/releaser/git"
	"github.com/Nightapes/go-semantic-release/internal/releaser/github"
	"github.com/Nightapes/go-semantic-release/internal/releaser/gitlab"
	"github.com/Nightapes/go-semantic-release/internal/shared"

	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

// Releasers struct type
type Releasers struct {
	config *config.ReleaseConfig
	git    *gitutil.GitUtil
}

// Releaser interface for providers
type Releaser interface {
	CreateRelease(*shared.ReleaseVersion, *shared.GeneratedChangelog, *assets.Set) error
	GetCommitURL() string
	GetCompareURL(oldVersion, newVersion string) string
}

// New initialize a releaser
func New(c *config.ReleaseConfig, git *gitutil.GitUtil) *Releasers {
	return &Releasers{
		config: c,
		git:    git,
	}
}

//GetReleaser returns an initialized releaser
func (r *Releasers) GetReleaser(checkConfig bool) (Releaser, error) {
	switch r.config.Release {
	case github.GITHUB:
		log.Debugf("initialize new %s-provider", github.GITHUB)
		return github.New(&r.config.GitHubProvider, checkConfig)
	case gitlab.GITLAB:
		log.Debugf("initialize new %s-provider", gitlab.GITLAB)
		return gitlab.New(&r.config.GitLabProvider, checkConfig)
	case git.GITONLY:
		log.Debugf("initialize new %s-provider", git.GITONLY)
		return git.New(&r.config.GitProvider, r.git, checkConfig)
	}
	return nil, fmt.Errorf("could not initialize a releaser from this type: %s", r.config.Release)
}
