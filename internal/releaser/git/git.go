package git

import (
	"fmt"
	"time"

	"github.com/Nightapes/go-semantic-release/internal/assets"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	log "github.com/sirupsen/logrus"
)

// GITONLY identifer for git interface
const GITONLY = "git"

// Client type struct
type Client struct {
	config *config.GitProvider
	log    *log.Entry
	git    *gitutil.GitUtil
}

// New initialize a new gitRelease
func New(config *config.GitProvider, git *gitutil.GitUtil, checkConfig bool) (*Client, error) {

	logger := log.WithField("releaser", GITONLY)

	if config.Email == "" && checkConfig {
		return nil, fmt.Errorf("git email not set")
	}

	if config.Username == "" && checkConfig {
		return nil, fmt.Errorf("git username not set")
	}

	if !config.SSH && config.Auth == "" && checkConfig {
		return nil, fmt.Errorf("git auth not set")
	}

	if config.SSH {
		return nil, fmt.Errorf("git ssh not supported yet")
	}

	return &Client{
		config: config,
		log:    logger,
		git:    git,
	}, nil
}

//GetCommitURL for git
func (g *Client) GetCommitURL() string {
	return ""
}

//GetCompareURL for git
func (g *Client) GetCompareURL(oldVersion, newVersion string) string {
	return ""
}

// CreateRelease creates release on remote
func (g *Client) CreateRelease(releaseVersion *shared.ReleaseVersion, generatedChangelog *shared.GeneratedChangelog, _ *assets.Set) error {

	tagPrefix := "v"
	if g.config.SkipTagPrefix {
		tagPrefix = ""
	}
	tag := tagPrefix + releaseVersion.Next.Version.String()
	g.log.Infof("create release with version %s", tag)

	head, err := g.git.Repository.Head()
	if err != nil {
		return err
	}

	_, err = g.git.Repository.CreateTag(tag, head.Hash(), &git.CreateTagOptions{Message: "Release " + tag, Tagger: &object.Signature{
		Name:  g.config.Username,
		Email: g.config.Email,
		When:  time.Now(),
	}})
	if err != nil {
		return err
	}

	g.log.Infof("Created release")

	return g.git.Repository.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: g.config.Username,
			Password: g.config.Auth,
		},
		RefSpecs: []gitConfig.RefSpec{"refs/tags/*:refs/tags/*"},
	})

}

// UploadAssets uploads specified assets
func (g *Client) UploadAssets(repoDir string, assets []config.Asset) error {
	return nil
}
