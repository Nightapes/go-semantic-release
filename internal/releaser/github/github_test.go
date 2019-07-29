package github_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Nightapes/go-semantic-release/internal/releaser/github"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

type testDouble struct {
	config config.GitHubProvider
	valid  bool
}

var doublesNew = []testDouble{
	testDouble{config: config.GitHubProvider{
		Repo: "foo",
		User: "bar",
	},
		valid: true,
	},

	testDouble{config: config.GitHubProvider{
		Repo:      "foo",
		User:      "bar",
		CustomURL: "https://test.com",
	},
		valid: false,
	},
}

var doublesValidateConfig = []testDouble{
	testDouble{config: config.GitHubProvider{
		Repo: "foo",
		User: "bar",
	},
		valid: true,
	},

	testDouble{config: config.GitHubProvider{
		Repo: "",
		User: "bar",
	},
		valid: false,
	},

	testDouble{config: config.GitHubProvider{
		Repo: "foo",
		User: "",
	},
		valid: false,
	},
}

func TestNew(t *testing.T) {
	for _, testOject := range doublesNew {
		if testOject.valid {
			os.Setenv("GITHUB_ACCESS_TOKEN", "XXX")
		}

		_, err := github.New(&testOject.config)
		assert.Equal(t, testOject.valid, err == nil)

		os.Unsetenv("GITHUB_ACCESS_TOKEN")

	}
}

func TestGetCommitURL(t *testing.T) {
	os.Setenv("GITHUB_ACCESS_TOKEN", "XX")
	for _, testOject := range doublesNew {
		client, _ := github.New(&testOject.config)
		actualUrl := client.GetCommitURL()
		if testOject.config.CustomURL != "" {
			expectedUrl := fmt.Sprintf("%s/%s/%s/commit/{{hash}}", testOject.config.CustomURL, testOject.config.User, testOject.config.Repo)
			assert.EqualValues(t, expectedUrl, actualUrl)

		} else {
			expectedUrl := fmt.Sprintf("%s/%s/%s/commit/{{hash}}", "https://github.com", testOject.config.User, testOject.config.Repo)
			assert.EqualValues(t, expectedUrl, actualUrl)
		}
	}
	os.Unsetenv("GITHUB_ACCESS_TOKEN")

}

func TestGetCompareURL(t *testing.T) {
	os.Setenv("GITHUB_ACCESS_TOKEN", "XX")
	for _, testOject := range doublesNew {
		client, _ := github.New(&testOject.config)
		actualUrl := client.GetCompareURL("1", "2")
		if testOject.config.CustomURL != "" {
			expectedUrl := fmt.Sprintf("%s/%s/%s/compare/%s...%s", testOject.config.CustomURL, testOject.config.User, testOject.config.Repo, "1", "2")
			assert.EqualValues(t, expectedUrl, actualUrl)

		} else {
			expectedUrl := fmt.Sprintf("%s/%s/%s/compare/%s...%s", "https://github.com", testOject.config.User, testOject.config.Repo, "1", "2")
			assert.EqualValues(t, expectedUrl, actualUrl)
		}
	}
	os.Unsetenv("GITHUB_ACCESS_TOKEN")

}

func TestValidateConfig(t *testing.T) {
	os.Setenv("GITHUB_ACCESS_TOKEN", "XX")
	for _, testOject := range doublesValidateConfig {
		client, _ := github.New(&testOject.config)
		err := client.ValidateConfig()

		assert.Equal(t, testOject.valid, err == nil)

	}
	os.Unsetenv("GITHUB_ACCESS_TOKEN")
}
