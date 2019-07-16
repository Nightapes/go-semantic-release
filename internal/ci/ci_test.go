package ci_test

import (
	"testing"

	"github.com/Nightapes/go-semantic-release/internal/ci"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func TestSum(t *testing.T) {
	testConfigs := []struct {
		service  string
		envs     map[string]string
		result   *ci.ProviderConfig
		hasError bool
	}{
		{
			service:  "none",
			envs:     map[string]string{},
			result:   nil,
			hasError: true,
		},
		// {
		// 	service: "Git",
		// 	envs: map[string]string{
		// 		"CI": "true",
		// 	},
		// 	result:   &ci.ProviderConfig{IsPR: true, PR: "10", PRBranch: "pr", Branch: "master", Tag: "TAG", Commit: "190bfd6aa60022afd0ef830342cfb07e33c45f37", BuildURL: "https://travis-ci.com/owner/repo/builds/1234", Service: "travis", Name: "Travis CI"},
		// 	hasError: false,
		// },
		{
			service: "Travis PR",
			envs: map[string]string{
				"TRAVIS":                     "true",
				"TRAVIS_PULL_REQUEST":        "10",
				"TRAVIS_COMMIT":              "190bfd6aa60022afd0ef830342cfb07e33c45f37",
				"TRAVIS_TAG":                 "TAG",
				"TRAVIS_BUILD_WEB_URL":       "https://travis-ci.com/owner/repo/builds/1234",
				"TRAVIS_BRANCH":              "master",
				"TRAVIS_PULL_REQUEST_BRANCH": "pr",
			},
			result:   &ci.ProviderConfig{IsPR: true, PR: "10", PRBranch: "pr", Branch: "master", Tag: "TAG", Commit: "190bfd6aa60022afd0ef830342cfb07e33c45f37", BuildURL: "https://travis-ci.com/owner/repo/builds/1234", Service: "travis", Name: "Travis CI"},
			hasError: false,
		},
		{
			service: "Travis Push",
			envs: map[string]string{
				"TRAVIS":               "true",
				"TRAVIS_PULL_REQUEST":  "false",
				"TRAVIS_COMMIT":        "190bfd6aa60022afd0ef830342cfb07e33c45f37",
				"TRAVIS_TAG":           "TAG",
				"TRAVIS_BUILD_WEB_URL": "https://travis-ci.com/owner/repo/builds/1234",
				"TRAVIS_BRANCH":        "master",
			},
			result:   &ci.ProviderConfig{IsPR: false, PR: "", PRBranch: "", Branch: "master", Tag: "TAG", Commit: "190bfd6aa60022afd0ef830342cfb07e33c45f37", BuildURL: "https://travis-ci.com/owner/repo/builds/1234", Service: "travis", Name: "Travis CI"},
			hasError: false,
		},
	}

	repository, err := git.Init(memory.NewStorage(), nil)
	assert.NoError(t, err, "should open git repository")

	gitUtilInMemory := &gitutil.GitUtil{
		Repository: repository,
	}

	for _, config := range testConfigs {
		provider, err := ci.GetCIProvider(gitUtilInMemory, config.envs)
		assert.Equalf(t, config.hasError, err != nil, "Service %s should have error: %t -> %s", config.service, config.hasError, err)
		assert.Equalf(t, config.result, provider, "Service %s should have provider", config.service)
	}

}
