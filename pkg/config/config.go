// Package config provides defimition of .release.yml and read method
package config

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ChangelogConfig struct
type ChangelogConfig struct {
	Print        string `yaml:"print,omitempty"`
	Template     string `yaml:"template,omitempty"`
	TemplatePath string `yaml:"templatePath,omitempty"`
}

// GithubConfig struct
type GithubConfig struct {
	URL string `yaml:"url"`
}

// GitlabConfig struct
type GitlabConfig struct {
	URL string `yaml:"url"`
}

// ReleaseConfig struct
type ReleaseConfig struct {
	CommitFormat string            `yaml:"commitFormat"`
	Branch       map[string]string `yaml:"branch"`
	Changelog    ChangelogConfig   `yaml:"changelog,omitempty"`
	Release      string            `yaml:"release,omitempty"`
	Github       map[string]string `yaml:"github"`
	Gitlab       map[string]string `yaml:"gitlab"`
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
