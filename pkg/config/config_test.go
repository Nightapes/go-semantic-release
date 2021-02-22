package config_test

import (
	"testing"

	"io/ioutil"
	"os"
	"path"

	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestReadCacheNotFound(t *testing.T) {

	_, err := config.Read("notfound/dir")
	assert.Errorf(t, err, "Read non exsiting file")

}

func TestReadCacheInvalidContent(t *testing.T) {

	dir, err := ioutil.TempDir("", "prefix")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	completePath := path.Join(path.Dir(dir), ".release.yml")
	brokenContent := []byte("hello broken\ngo: lang\n")
	err = ioutil.WriteFile(completePath, brokenContent, 0644)
	assert.NoError(t, err)

	_, readError := config.Read(completePath)
	assert.Errorf(t, readError, "Should give error, when broken content")

}

func TestWriteAndReadCache(t *testing.T) {

	dir, err := ioutil.TempDir("", "prefix")

	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	os.Setenv("TEST_CONFIG", "value")
	defer os.Unsetenv("TEST_CONFIG")

	completePath := path.Join(path.Dir(dir), ".release.yml")
	content := []byte(`
commitFormat: angular
title: "go-semantic-release release"
branch:
  master: release
  rc: rc
  beta: beta
  alpha: alpha
  add_git_releases: alpha
changelog:
  printAll: false
  template: ""
  templatePath: '${TEST_CONFIG}'
release: 'github'
hooks:
  preRelease:
    - "Test hook ${RELEASE_VERSION}"
assets:
  - name: ./build/go-semantic-release
    compress: false
github:
  repo: "go-semantic-release"
  user: "nightapes"
  customUrl: ""
`)
	err = ioutil.WriteFile(completePath, content, 0644)
	assert.NoError(t, err)

	result, readError := config.Read(completePath)
	assert.NoErrorf(t, readError, "Should read file")

	assert.EqualValues(t, &config.ReleaseConfig{
		CommitFormat: "angular",
		Branch:       map[string]string{"add_git_releases": "alpha", "alpha": "alpha", "beta": "beta", "master": "release", "rc": "rc"},
		Changelog: config.ChangelogConfig{
			PrintAll:      false,
			TemplateTitle: "",
			TemplatePath:  "value"},
		Release: "github",
		GitHubProvider: config.GitHubProvider{
			Repo:        "go-semantic-release",
			User:        "nightapes",
			CustomURL:   "",
			AccessToken: ""},
		Hooks: config.Hooks{
			PreRelease: []string{
				"Test hook ${RELEASE_VERSION}",
			},
		},
		Assets: []config.Asset{
			{
				Name:     "./build/go-semantic-release",
				Compress: false}},
		ReleaseTitle: "go-semantic-release release",
		IsPreRelease: false,
		Analyzer: config.AnalyzerConfig{TokenSeparators: []string{}},
	}, result)

}
