package ci

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

//Travis struct
type Travis struct{}

//Detect if on travis
func (t Travis) detect() (*ProviderConfig, error) {

	if _, exists := os.LookupEnv("TRAVIS"); !exists {
		return nil, fmt.Errorf("not running on travis")
	}

	isPR := false

	value := os.Getenv("TRAVIS_PULL_REQUEST")
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
		Commit:   os.Getenv("TRAVIS_COMMIT"),
		Tag:      os.Getenv("TRAVIS_TAG"),
		BuildURL: os.Getenv("TRAVIS_BUILD_WEB_URL"),
		Branch:   os.Getenv("TRAVIS_BRANCH"),
		IsPR:     isPR,
		PR:       pr,
		PRBranch: os.Getenv("TRAVIS_PULL_REQUEST_BRANCH"),
	}, nil
}
