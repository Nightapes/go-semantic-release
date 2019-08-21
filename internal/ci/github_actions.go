package ci

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

//GithubActions struct
type GithubActions struct{}

//Detect if on GithubActions
func (t GithubActions) detect(envs map[string]string) (*ProviderConfig, error) {

	if _, exists := envs["GITHUB_ACTION"]; !exists {
		return nil, fmt.Errorf("not running on Github Actions")
	}

	isPR := false

	value := envs["GITHUB_EVENT_NAME"]

	if value == "pull_request" {
		isPR = true
	} else {
		log.Debugf("GITHUB_EVENT_NAME=%s, not running on pr", value)
	}

	return &ProviderConfig{
		Service: "GithubActions",
		Name:    "GithubActions CI",
		Commit:  envs["GITHUB_SHA"],
		Branch:  envs["GITHUB_REF"],
		IsPR:    isPR,
	}, nil
}
