package releaser

import (
	"fmt"
	"github.com/Nightapes/go-semantic-release/internal/cache"
	"github.com/Nightapes/go-semantic-release/internal/releaser/releasers"
	"github.com/Nightapes/go-semantic-release/pkg/config"
)

// Releasers struct type
type Releasers struct {
	config *config.ReleaseConfig
}

// Releaser interface for providers
type Releaser interface {
	CreateRelease(releaseName, releaseMessage string) error
}

// New initialize a Relerser
func New(c *config.ReleaseConfig) *Releasers {
	return &Releasers{
		config: c,
	}
}

//GetReleaser returns an initialized releaser
func (r *Releasers) GetReleaser(releaserType string) (Releaser, error) {
	switch releaserType {
	case releasers.GITHUB:
		return releasers.NewGitHubReleaser(r.config), nil
	}
	return nil, fmt.Errorf("Could not initialize a releaser from this type: %s", releaserType)
}

// tbd. http helper function
