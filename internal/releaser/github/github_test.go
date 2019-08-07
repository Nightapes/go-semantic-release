package github_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Masterminds/semver"

	"github.com/Nightapes/go-semantic-release/internal/releaser/github"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

type testDouble struct {
	config config.GitHubProvider
	valid  bool
}

type testFourth struct {
	config              config.GitHubProvider
	releaseVersion      *shared.ReleaseVersion
	generatedChangelog  *shared.GeneratedChangelog
	requestResponseBody string
	requestResponseCode int
	valid               bool
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

var lastVersion, _ = semver.NewVersion("1.0.0")
var newVersion, _ = semver.NewVersion("2.0.0")

var fourthsReleas = []testFourth{
	testFourth{
		config: config.GitHubProvider{
			Repo: "foo",
			User: "bar",
		},
		releaseVersion: &shared.ReleaseVersion{
			Last: shared.ReleaseVersionEntry{
				Version: lastVersion,
				Commit:  "foo",
			},
			Next: shared.ReleaseVersionEntry{
				Version: newVersion,
				Commit:  "bar",
			},
			Branch: "master",
			Draft:  false,
		},
		generatedChangelog: &shared.GeneratedChangelog{
			Title:   "title",
			Content: "content",
		},
		requestResponseBody: "{  \"url\": \"https://api.github.com/repos/octocat/Hello-World/releases/1\",  \"html_url\": \"https://github.com/octocat/Hello-World/releases/v1.0.0\",  \"assets_url\": \"https://api.github.com/repos/octocat/Hello-World/releases/1/assets\",  \"upload_url\": \"https://uploads.github.com/repos/octocat/Hello-World/releases/1/assets{?name,label}\",  \"tarball_url\": \"https://api.github.com/repos/octocat/Hello-World/tarball/v1.0.0\",  \"zipball_url\": \"https://api.github.com/repos/octocat/Hello-World/zipball/v1.0.0\",  \"id\": 1,  \"node_id\": \"MDc6UmVsZWFzZTE=\",  \"tag_name\": \"v1.0.0\",  \"target_commitish\": \"master\",  \"name\": \"v1.0.0\",  \"body\": \"Description of the release\",  \"draft\": false,  \"prerelease\": false,  \"created_at\": \"2013-02-27T19:35:32Z\",  \"published_at\": \"2013-02-27T19:35:32Z\",  \"author\": {    \"login\": \"octocat\",    \"id\": 1,    \"node_id\": \"MDQ6VXNlcjE=\",    \"avatar_url\": \"https://github.com/images/error/octocat_happy.gif\",    \"gravatar_id\": \"\",    \"url\": \"https://api.github.com/users/octocat\",    \"html_url\": \"https://github.com/octocat\",    \"followers_url\": \"https://api.github.com/users/octocat/followers\",    \"following_url\": \"https://api.github.com/users/octocat/following{/other_user}\",    \"gists_url\": \"https://api.github.com/users/octocat/gists{/gist_id}\",    \"starred_url\": \"https://api.github.com/users/octocat/starred{/owner}{/repo}\",    \"subscriptions_url\": \"https://api.github.com/users/octocat/subscriptions\",    \"organizations_url\": \"https://api.github.com/users/octocat/orgs\",    \"repos_url\": \"https://api.github.com/users/octocat/repos\",    \"events_url\": \"https://api.github.com/users/octocat/events{/privacy}\",    \"received_events_url\": \"https://api.github.com/users/octocat/received_events\",    \"type\": \"User\",    \"site_admin\": false  },  \"assets\": [  ]}",
		requestResponseCode: 500,
		valid:               true,
	},
	testFourth{
		config: config.GitHubProvider{
			Repo: "foo",
			User: "bar",
		},
		releaseVersion: &shared.ReleaseVersion{
			Last: shared.ReleaseVersionEntry{
				Version: lastVersion,
				Commit:  "foo",
			},
			Next: shared.ReleaseVersionEntry{
				Version: newVersion,
				Commit:  "bar",
			},
			Branch: "master",
			Draft:  false,
		},
		generatedChangelog: &shared.GeneratedChangelog{
			Title:   "title",
			Content: "content",
		},
		requestResponseBody: "",
		requestResponseCode: 422,
		valid:               false,
	},
}

func initHTTPServer(respCode int, body string) *httptest.Server {

	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		rw.Write([]byte(body))
		rw.Header().Set("Content-Type", "application/json")

		rw.WriteHeader(respCode)
	}))
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

func TestCreateRelease(t *testing.T) {
	os.Setenv("GITHUB_ACCESS_TOKEN", "XX")

	for _, testObejct := range fourthsReleas {
		if testObejct.valid {
			server := initHTTPServer(testObejct.requestResponseCode, "")
			testObejct.config.CustomURL = server.URL
			client, _ := github.New(&testObejct.config)

			err := client.CreateRelease(testObejct.releaseVersion, testObejct.generatedChangelog)
			assert.Equal(t, testObejct.valid, err == nil)

		} else {
			testObejct.config.CustomURL = "foo"
			client, _ := github.New(&testObejct.config)

			err := client.CreateRelease(testObejct.releaseVersion, testObejct.generatedChangelog)
			if err != nil {
				t.Log(err)
			}
			assert.Equal(t, testObejct.valid, err == nil)
		}
	}
	os.Unsetenv("GITHUB_ACCESS_TOKEN")

}
