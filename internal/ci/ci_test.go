package ci_test

import (
	"testing"
	"time"

	"github.com/Nightapes/go-semantic-release/internal/ci"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func TestCi(t *testing.T) {

	fs := memfs.New()

	repository, err := git.Init(memory.NewStorage(), fs)
	assert.NoError(t, err, "should open git repository")

	file, err := fs.Create("README.md")
	assert.NoError(t, err, "should create file")

	w, err := repository.Worktree()
	assert.NoError(t, err, "should get worktree")

	_, err = w.Add(file.Name())
	assert.NoError(t, err, "should add file")

	gitUtilInMemory := &gitutil.GitUtil{
		Repository: repository,
	}

	newCommit, err := w.Commit("fix(test): add a commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err, "should commit")

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
		{
			service: "Git",
			envs: map[string]string{
				"CI": "true",
			},
			result:   &ci.ProviderConfig{IsPR: false, PR: "", PRBranch: "", Branch: "master", Tag: "", Commit: newCommit.String(), BuildURL: "", Service: "git", Name: "Git only"},
			hasError: false,
		},
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
		{
			service: "Github Actions PR",
			envs: map[string]string{
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_SHA":        "190bfd6aa60022afd0ef830342cfb07e33c45f37",
				"GITHUB_REF":        "master",
				"GITHUB_ACTION":     "action",
			},
			result:   &ci.ProviderConfig{IsPR: true, PR: "", PRBranch: "", Branch: "master", Tag: "", Commit: "190bfd6aa60022afd0ef830342cfb07e33c45f37", BuildURL: "", Service: "GithubActions", Name: "GithubActions CI"},
			hasError: false,
		},
		{
			service: "Github Actions Push",
			envs: map[string]string{
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_SHA":        "190bfd6aa60022afd0ef830342cfb07e33c45f37",
				"GITHUB_REF":        "refs/heads/feature-branch-1",
				"GITHUB_ACTION":     "action",
			},
			result:   &ci.ProviderConfig{IsPR: false, PR: "", PRBranch: "", Branch: "feature-branch-1", Tag: "", Commit: "190bfd6aa60022afd0ef830342cfb07e33c45f37", BuildURL: "", Service: "GithubActions", Name: "GithubActions CI"},
			hasError: false,
		},
		{
			service: "GitLab CI/CD PR",
			envs: map[string]string{
				"GITLAB_CI":          "true",
				"CI_COMMIT_SHA":      "190bfd6aa60022afd0ef830342cfb07e33c45f37",
				"CI_COMMIT_REF_NAME": "master",
				"CI_COMMIT_TAG":      "tag",
				"CI_PROJECT_URL":     "https://my.gitlab.com",
				"CI_PIPELINE_ID":     "1",
			},
			result:   &ci.ProviderConfig{IsPR: false, PR: "", PRBranch: "", Branch: "master", Tag: "tag", Commit: "190bfd6aa60022afd0ef830342cfb07e33c45f37", BuildURL: "https://my.gitlab.com/pipelines/1", Service: "gitlab", Name: "GitLab CI/CD"},
			hasError: false,
		},
	}

	for _, config := range testConfigs {
		provider, err := ci.GetCIProvider(gitUtilInMemory, true, config.envs)
		assert.Equalf(t, config.hasError, err != nil, "Service %s should have error: %t -> %s", config.service, config.hasError, err)
		assert.Equalf(t, config.result, provider, "Service %s should have provider", config.service)
	}

}
