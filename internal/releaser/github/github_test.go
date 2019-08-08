package github_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/Masterminds/semver"

	"github.com/Nightapes/go-semantic-release/internal/releaser/github"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

type testHelperMethodStruct struct {
	config config.GitHubProvider
	valid  bool
}

type testReleaseStruct struct {
	config              config.GitHubProvider
	releaseVersion      *shared.ReleaseVersion
	generatedChangelog  *shared.GeneratedChangelog
	requestResponseBody string
	requestResponseCode int
	valid               bool
}

var testNewClient = []testHelperMethodStruct{
	testHelperMethodStruct{config: config.GitHubProvider{
		Repo: "foo",
		User: "bar",
	},
		valid: true,
	},

	testHelperMethodStruct{config: config.GitHubProvider{
		Repo:      "foo",
		User:      "bar",
		CustomURL: "https://test.com",
	},
		valid: false,
	},
}

var testHelperMethod = []testHelperMethodStruct{
	testHelperMethodStruct{config: config.GitHubProvider{
		Repo: "foo",
		User: "bar",
	},
		valid: true,
	},

	testHelperMethodStruct{config: config.GitHubProvider{
		Repo: "",
		User: "bar",
	},
		valid: false,
	},

	testHelperMethodStruct{config: config.GitHubProvider{
		Repo: "foo",
		User: "",
	},
		valid: false,
	},
}

var lastVersion, _ = semver.NewVersion("1.0.0")
var newVersion, _ = semver.NewVersion("2.0.0")

var testReleases = []testReleaseStruct{
	testReleaseStruct{
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
		requestResponseCode: 200,
		valid:               true,
	},
	testReleaseStruct{
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
		requestResponseCode: 400,
		valid:               false,
	},
}

func initHTTPServer(respCode int, body string) *httptest.Server {

	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		log.Infof("Got call from %s %s", req.Method, req.URL.String())

		rw.WriteHeader(respCode)
		rw.Header().Set("Content-Type", "application/json")

		if _, err := rw.Write([]byte(body)); err != nil {
			log.Info(err)
		}

	}))
}

func TestNew(t *testing.T) {
	for _, testOject := range testNewClient {
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
	for _, testOject := range testNewClient {
		client, _ := github.New(&testOject.config)
		actualURL := client.GetCommitURL()
		if testOject.config.CustomURL != "" {
			expectedURL := fmt.Sprintf("%s/%s/%s/commit/{{hash}}", testOject.config.CustomURL, testOject.config.User, testOject.config.Repo)
			assert.EqualValues(t, expectedURL, actualURL)

		} else {
			expectedURL := fmt.Sprintf("%s/%s/%s/commit/{{hash}}", "https://github.com", testOject.config.User, testOject.config.Repo)
			assert.EqualValues(t, expectedURL, actualURL)
		}
	}
	os.Unsetenv("GITHUB_ACCESS_TOKEN")

}

func TestGetCompareURL(t *testing.T) {
	os.Setenv("GITHUB_ACCESS_TOKEN", "XX")
	for _, testOject := range testNewClient {
		client, _ := github.New(&testOject.config)
		actualURL := client.GetCompareURL("1", "2")
		if testOject.config.CustomURL != "" {
			expectedURL := fmt.Sprintf("%s/%s/%s/compare/%s...%s", testOject.config.CustomURL, testOject.config.User, testOject.config.Repo, "1", "2")
			assert.EqualValues(t, expectedURL, actualURL)

		} else {
			expectedURL := fmt.Sprintf("%s/%s/%s/compare/%s...%s", "https://github.com", testOject.config.User, testOject.config.Repo, "1", "2")
			assert.EqualValues(t, expectedURL, actualURL)
		}
	}
	os.Unsetenv("GITHUB_ACCESS_TOKEN")

}

func TestValidateConfig(t *testing.T) {
	os.Setenv("GITHUB_ACCESS_TOKEN", "XX")
	for _, testOject := range testHelperMethod {
		client, _ := github.New(&testOject.config)
		err := client.ValidateConfig()

		assert.Equal(t, testOject.valid, err == nil)

	}
	os.Unsetenv("GITHUB_ACCESS_TOKEN")
}

func TestCreateRelease(t *testing.T) {
	os.Setenv("GITHUB_ACCESS_TOKEN", "XX")

	for _, testObejct := range testReleases {
		if testObejct.valid {
			server := initHTTPServer(testObejct.requestResponseCode, testObejct.requestResponseBody)
			testObejct.config.CustomURL = server.URL
			client, _ := github.New(&testObejct.config)

			err := client.CreateRelease(testObejct.releaseVersion, testObejct.generatedChangelog)
			if err != nil {
				t.Log(err)
			}
			assert.Equal(t, testObejct.valid, err == nil)

			server.Close()

		} else {
			testObejct.config.CustomURL = "http://foo"
			client, _ := github.New(&testObejct.config)

			err := client.CreateRelease(testObejct.releaseVersion, testObejct.generatedChangelog)
			if err != nil {
				t.Log(err)
			}
			assert.Error(t, err)
		}
	}
	os.Unsetenv("GITHUB_ACCESS_TOKEN")

}
