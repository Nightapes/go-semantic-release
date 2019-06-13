package releaser

import (
	"context"
	"fmt"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

// Releasers struct type
type Releasers struct {
	config *config.ReleaseConfig
}

// Releaser interface for providers
type Releaser interface {
	CreateRelease(tag, releaseName, releaseMessage, targetBranch string) error
	UploadAssets(assets []config.Asset) error
}

// New initialize a Relerser
func New(c *config.ReleaseConfig) *Releasers {
	return &Releasers{
		config: c,
	}
}

//GetReleaser returns an initialized releaser
func (r *Releasers) GetReleaser() (Releaser, error) {
	switch r.config.GitProvider.Name {
	case GITHUB:
		return NewGitHubReleaser(r.config), nil
	}
	return nil, fmt.Errorf("Could not initialize a releaser from this type: %s", r.config.GitProvider.Name)
}

// tbd. http helper function

func createHTTPClient(ctx context.Context, token string) *http.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token},
	)

	client := oauth2.NewClient(ctx, tokenSource)

	return client
}

func checkIfAssetsExists(assets []config.Asset) error {
	var missingAssets []string
	for _, asset := range assets {

		if _, err := os.Stat(asset.Name); err != nil {
			missingAssets = append(missingAssets, asset.Name)
		}
	}

	if len(missingAssets) != 0 {
		return fmt.Errorf("Could not find specified Asset: %+v ", assets)
	}

	return nil

}
