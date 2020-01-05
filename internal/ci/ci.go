package ci

import (
	"fmt"

	"os"
	"strings"

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
	detect(envs map[string]string) (*ProviderConfig, error)
}

//ReadAllEnvs as a map
func ReadAllEnvs() map[string]string {
	envs := map[string]string{}
	for _, pair := range os.Environ() {
		splitted := strings.SplitN(pair, "=", 2)
		envs[splitted[0]] = splitted[1]
	}
	return envs
}

//GetCIProvider get provider
func GetCIProvider(gitUtil *gitutil.GitUtil, configCheck bool, envs map[string]string) (*ProviderConfig, error) {

	services := []Service{
		Travis{},
		GithubActions{},
		GitlabCI{},
		Git{gitUtil: gitUtil}, // Git must be the last option to check
	}

	for _, service := range services {
		config, err := service.detect(envs)
		if err == nil {
			log.Infof("Found CI: %s", config.Name)
			log.Tracef("Found CI config: %+v", config)
			return config, nil
		}
		log.Debugf("%s", err.Error())
	}
	if configCheck {
		return nil, fmt.Errorf("could not find any CI, if running locally set env CI=true")

	}
	return Git{gitUtil: gitUtil}.detect(map[string]string{"CI": "true"})
}
