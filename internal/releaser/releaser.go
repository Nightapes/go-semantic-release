package releaser

import (
	"fmt"
	"strings"

	"github.com/Nightapes/go-semantic-release/internal/releaser/github"
	"github.com/Nightapes/go-semantic-release/internal/releaser/gitlab"
	"github.com/Nightapes/go-semantic-release/internal/releaser/util"
	"github.com/Nightapes/go-semantic-release/internal/shared"

	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

// Releasers struct type
type Releasers struct {
	config *config.ReleaseConfig
}

// Releaser interface for providers
type Releaser interface {
	ValidateConfig() error
	CreateRelease(*shared.ReleaseVersion, *shared.GeneratedChangelog) error
	UploadAssets(repoDir string, assets []config.Asset) error
	GetCommitURL() string
	GetCompareURL(oldVersion, newVersion string) string
}

// New initialize a Relerser
func New(c *config.ReleaseConfig) *Releasers {
	return &Releasers{
		config: c,
	}
}

//GetReleaser returns an initialized releaser
func (r *Releasers) GetReleaser() (Releaser, error) {
	switch r.config.Release {
	case github.GITHUB:
		log.Debugf("initialize new %s-provider", github.GITHUB)
		return github.New(&r.config.GitHubProvider)
	case gitlab.GITLAB:
		log.Debugf("initialize new %s-provider", gitlab.GITLAB)
		accessToken, err := util.GetAccessToken(fmt.Sprintf("%s_ACCESS_TOKEN", strings.ToUpper(gitlab.GITLAB)))
		if err != nil {
			return nil, err
		}
		return gitlab.New(&r.config.GitLabProvider, accessToken)
	}
	return nil, fmt.Errorf("could not initialize a releaser from this type: %s", r.config.Release)
}
