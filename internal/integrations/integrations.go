package integrations

import (
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
)

// Integrations struct
type Integrations struct {
	version *shared.ReleaseVersion
	config  *config.Integrations
}

func New(config *config.Integrations, version *shared.ReleaseVersion) *Integrations {
	return &Integrations{
		config:  config,
		version: version,
	}
}

func (i Integrations) Run() error {
	if i.config.NPM.Enabled {
		return i.updateNPM()
	}
	return nil
}
