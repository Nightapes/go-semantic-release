package ci

import (
	"fmt"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	log "github.com/sirupsen/logrus"
)

//ProviderConfig struct
type ProviderConfig struct {
	IsPR     bool
	PR       string
	PRBranch string
	Branch   string
	Tag      string
	Commit   string
	BuildURL string
	Service  string
	Name     string
}

//Service interface
type Service interface {
	detect() (*ProviderConfig, error)
}

//GetCIProvider get provider
func GetCIProvider(gitUtil *gitutil.GitUtil) (*ProviderConfig, error) {

	services := []Service{
		Travis{},
		Git{gitUtil: gitUtil}, // GIt must be the last option to check
	}

	for _, service := range services {
		config, err := service.detect()
		if err == nil {
			log.Infof("Found CI: %s", config.Name)
			return config, nil
		}
		log.Infof("%s", err.Error())
	}
	return nil, fmt.Errorf("could not find any CI, if running locally set env CI=true")
}
