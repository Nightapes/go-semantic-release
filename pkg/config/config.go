// Package config provides defimition of .release.yml and read method
package config

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ChangelogConfig struct
type ChangelogConfig struct {
	PrintAll     bool   `yaml:"printAll,omitempty"`
	Template     string `yaml:"template,omitempty"`
	TemplatePath string `yaml:"templatePath,omitempty"`
}

//Asset type struct
type Asset struct {
	Name     string `yaml:"name"`
	Compress bool   `yaml:"compress"`
}

// GitHubProvider struct
type GitHubProvider struct {
	Repo        string `yaml:"repo"`
	User        string `yaml:"user"`
	CustomURL   string `yaml:"customUrl,omitempty"`
	AccessToken string
}

// ReleaseConfig struct
type ReleaseConfig struct {
	CommitFormat          string            `yaml:"commitFormat"`
	Branch                map[string]string `yaml:"branch"`
	Changelog             ChangelogConfig   `yaml:"changelog,omitempty"`
	Release               string            `yaml:"release,omitempty"`
	GitHubProvider        GitHubProvider    `yaml:"github,omitempty"`
	Assets                []Asset           `yaml:"assets"`
	ReleaseTitle          string            `yaml:"title"`
	IsPreRelease, IsDraft bool
}

// Read ReleaseConfig
func Read(configPath string) (*ReleaseConfig, error) {

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return &ReleaseConfig{}, err
	}

	var releaseConfig ReleaseConfig
	err = yaml.Unmarshal(content, &releaseConfig)
	if err != nil {
		return &ReleaseConfig{}, err
	}

	log.Debugf("Found config %+v", releaseConfig)

	return &releaseConfig, nil
}
