package github_test

import (
	"github.com/Nightapes/go-semantic-release/internal/releaser/github"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type testDoubleNew struct {
	config config.GitHubProvider
	valid  bool
}

var doublesNew = []testDoubleNew{
	testDoubleNew{config: config.GitHubProvider{
		Repo: "foo",
		User: "bar",
	},
		valid: true,
	},

	testDoubleNew{config: config.GitHubProvider{
		Repo: "foo",
		User: "bar",
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
