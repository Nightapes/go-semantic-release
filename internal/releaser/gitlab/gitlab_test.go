package gitlab

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/Nightapes/go-semantic-release/internal/assets"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
)

func TestGetCommitURL(t *testing.T) {
	os.Setenv("GITLAB_ACCESS_TOKEN", "XXX")
	defer os.Unsetenv("GITLAB_ACCESS_TOKEN")
	client, err := New(&config.GitLabProvider{
		CustomURL: "https://127.0.0.1/",
		Repo:      "test/test",
	}, true)
	assert.NoError(t, err)
	assert.Equal(t, "https://127.0.0.1/test/test/commit/{{hash}}", client.GetCommitURL())
}

func TestGetCompareURL(t *testing.T) {
	os.Setenv("GITLAB_ACCESS_TOKEN", "XXX")
	defer os.Unsetenv("GITLAB_ACCESS_TOKEN")
	client, err := New(&config.GitLabProvider{
		CustomURL: "https://127.0.0.1/",
		Repo:      "test/test",
	}, true)
	assert.NoError(t, err)
	assert.Equal(t, "https://127.0.0.1/test/test/compare/1.0.0...1.0.1", client.GetCompareURL("1.0.0", "1.0.1"))
}

func TestValidateConfig_EmptyRepro(t *testing.T) {
	os.Setenv("GITLAB_ACCESS_TOKEN", "XXX")
	defer os.Unsetenv("GITLAB_ACCESS_TOKEN")
	_, err := New(&config.GitLabProvider{
		CustomURL: "https://127.0.0.1/",
	}, true)
	assert.Error(t, err)
}

func TestValidateConfig_DefaultURL(t *testing.T) {
	os.Setenv("GITLAB_ACCESS_TOKEN", "XXX")
	defer os.Unsetenv("GITLAB_ACCESS_TOKEN")
	config := &config.GitLabProvider{
		Repo: "127.0.0.1/test",
	}
	_, err := New(config, true)
	assert.NoError(t, err)
	assert.Equal(t, "https://gitlab.com", config.CustomURL)
}

func TestValidateConfig_CustomURL(t *testing.T) {
	os.Setenv("GITLAB_ACCESS_TOKEN", "XXX")
	defer os.Unsetenv("GITLAB_ACCESS_TOKEN")
	config := &config.GitLabProvider{
		Repo:      "/127.0.0.1/test/",
		CustomURL: "https://127.0.0.1/",
	}
	_, err := New(config, true)
	assert.NoError(t, err)
	assert.Equal(t, "https://127.0.0.1", config.CustomURL)
	assert.Equal(t, "127.0.0.1/test", config.Repo)
}

func TestCreateRelease(t *testing.T) {

	lastVersion, _ := semver.NewVersion("1.0.0")
	newVersion, _ := semver.NewVersion("2.0.0")

	testReleases := []struct {
		config             config.GitLabProvider
		releaseVersion     *shared.ReleaseVersion
		generatedChangelog *shared.GeneratedChangelog
		responseBody       string
		responseCode       int
		requestBody        string
		valid              bool
	}{
		{
			config: config.GitLabProvider{
				Repo: "foo/bar",
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
			},
			generatedChangelog: &shared.GeneratedChangelog{
				Title:   "title",
				Content: "content",
			},
			responseBody: "{}",
			responseCode: 200,
			requestBody:  `{"tag_name":"v2.0.0","name":"title","ref":"master","description":"content"}`,
			valid:        true,
		},
		{
			config: config.GitLabProvider{
				Repo: "foo/bar",
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
			},
			generatedChangelog: &shared.GeneratedChangelog{
				Title:   "title",
				Content: "content",
			},
			responseBody: "{}",
			responseCode: 500,
			requestBody:  `{"tag_name":"v2.0.0","name":"title","ref":"master","description":"content"}`,
			valid:        false,
		},
		{
			config: config.GitLabProvider{
				Repo:      "foo/bar",
				CustomURL: "broken",
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
			},
			generatedChangelog: &shared.GeneratedChangelog{
				Title:   "title",
				Content: "content",
			},
			responseCode: 400,
			responseBody: "{}",
			requestBody:  `{"tag_name":"v2.0.0","name":"title","ref":"master","description":"content"}`,
			valid:        false,
		},
	}

	for _, testObject := range testReleases {
		testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			log.Infof("Got call from %s %s", req.Method, req.URL.String())

			assert.Equal(t, req.Header.Get("PRIVATE-TOKEN"), "aToken")
			assert.Equal(t, req.Header.Get("Accept"), "application/json")

			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
			}
			assert.Equal(t, testObject.requestBody, string(bodyBytes))

			rw.WriteHeader(testObject.responseCode)
			rw.Header().Set("Content-Type", "application/json")

			if _, err := rw.Write([]byte(testObject.responseBody)); err != nil {
				log.Info(err)
			}

		}))

		if testObject.config.CustomURL == "" {
			testObject.config.CustomURL = testServer.URL
		}
		os.Setenv("GITLAB_ACCESS_TOKEN", "aToken")
		defer os.Unsetenv("GITLAB_ACCESS_TOKEN")
		client, err := New(&testObject.config, false)
		assert.NoError(t, err)

		err = client.makeRelease(testObject.releaseVersion, testObject.generatedChangelog)
		if err != nil {
			t.Log(err)
		}
		assert.Equal(t, testObject.valid, err == nil)

		testServer.Close()

	}
}

func TestUploadAssets(t *testing.T) {

	file, err := ioutil.TempFile("", "prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString("testFile")
	assert.NoError(t, err)

	testReleases := []struct {
		config       config.GitLabProvider
		responseBody []string
		responseCode []int
		assets       []config.Asset
		requestBody  []string
		testDir      string
		url          []string
		method       []string
		valid        bool
	}{
		{
			config: config.GitLabProvider{
				Repo: "foo/bar",
			},
			responseBody: []string{`{"alt" : "dk", "url": "/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png", "markdown" :"![dk](/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png)"}`, ""},
			responseCode: []int{200, 200},
			requestBody: []string{
				filepath.Base(file.Name()), ""},
			url:     []string{`/api/v4/projects/foo%2Fbar/uploads`, "/api/v4/projects/foo%2Fbar/releases/1.0.0/assets/links?name=" + filepath.Base(file.Name()) + "&url=<SERVER>/foo/bar/uploads/"},
			method:  []string{"POST", "POST"},
			valid:   true,
			testDir: os.TempDir(),
			assets: []config.Asset{
				{
					Name:     filepath.Base(file.Name()),
					Compress: false,
				},
			},
		},
		{
			config: config.GitLabProvider{
				Repo: "foo/bar",
			},
			responseBody: []string{`{"alt" : "dk", "url": "/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png", "markdown" :"![dk](/uploads/66dbcd21ec5d24ed6ea225176098d52b/dk.png)"}`, ""},
			responseCode: []int{400, 200},
			requestBody: []string{
				filepath.Base(file.Name()), ""},
			url:     []string{`/api/v4/projects/foo%2Fbar/uploads`, "/api/v4/projects/foo%2Fbar/releases/1.0.0/assets/links?name=" + filepath.Base(file.Name()) + "&url=<SERVER>/foo/bar/uploads/"},
			method:  []string{"POST", "POST"},
			valid:   false,
			testDir: os.TempDir(),
			assets: []config.Asset{
				{
					Name:     filepath.Base(file.Name()),
					Compress: false,
				},
			},
		},
		{
			config: config.GitLabProvider{
				Repo: "foo/bar",
			},
			responseBody: []string{`broken`, ""},
			responseCode: []int{200, 200},
			requestBody: []string{
				filepath.Base(file.Name()), ""},
			url:     []string{`/api/v4/projects/foo%2Fbar/uploads`, "/api/v4/projects/foo%2Fbar/releases/1.0.0/assets/links?name=" + filepath.Base(file.Name()) + "&url=<SERVER>/foo/bar/uploads/"},
			method:  []string{"POST", "POST"},
			valid:   false,
			testDir: os.TempDir(),
			assets: []config.Asset{
				{
					Name:     filepath.Base(file.Name()),
					Compress: false,
				},
			},
		},
	}

	for _, testObject := range testReleases {
		calls := 0
		testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			log.Infof("Got call from %s %s", req.Method, req.URL.String())

			assert.Contains(t, req.URL.String(), strings.ReplaceAll(testObject.url[calls], "<SERVER>", testObject.config.CustomURL))
			assert.Equal(t, req.Method, testObject.method[calls])

			assert.Equal(t, req.Header.Get("PRIVATE-TOKEN"), "aToken")
			assert.Equal(t, req.Header.Get("Accept"), "application/json")

			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Fatal(err)
			}
			assert.Contains(t, string(bodyBytes), testObject.requestBody[calls])

			rw.WriteHeader(testObject.responseCode[calls])
			rw.Header().Set("Content-Type", "application/json")

			if _, err := rw.Write([]byte(testObject.responseBody[calls])); err != nil {
				log.Info(err)
			}
			calls++
		}))

		if testObject.config.CustomURL == "" {
			testObject.config.CustomURL = testServer.URL
		}
		os.Setenv("GITLAB_ACCESS_TOKEN", "aToken")
		defer os.Unsetenv("GITLAB_ACCESS_TOKEN")
		client, err := New(&testObject.config, false)
		assert.NoError(t, err)
		client.Release = "1.0.0"

		assets := assets.New(testObject.testDir, "")
		err = assets.Add(testObject.assets...)
		if err != nil {
			t.Log(err)
		}
		err = client.uploadAssets(assets)
		if err != nil {
			t.Log(err)
		}
		assert.Equal(t, testObject.valid, err == nil)

		testServer.Close()

	}
}
