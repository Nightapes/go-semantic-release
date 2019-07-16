package ci

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

//Travis struct
type Travis struct{}

//Detect if on travis
func (t Travis) detect(envs map[string]string) (*ProviderConfig, error) {

	if _, exists := envs["TRAVIS"]; !exists {
		return nil, fmt.Errorf("not running on travis")
	}

	isPR := false

	value := envs["TRAVIS_PULL_REQUEST"]
	pr := ""

	if value == "false" {
		log.Debugf("TRAVIS_PULL_REQUEST=%s, not running on pr", value)
	} else {
		isPR = true
		pr = value
	}

	return &ProviderConfig{
		Service:  "travis",
		Name:     "Travis CI",
		Commit:   envs["TRAVIS_COMMIT"],
		Tag:      envs["TRAVIS_TAG"],
		BuildURL: envs["TRAVIS_BUILD_WEB_URL"],
		Branch:   envs["TRAVIS_BRANCH"],
		IsPR:     isPR,
		PR:       pr,
		PRBranch: envs["TRAVIS_PULL_REQUEST_BRANCH"],
	}, nil
}
