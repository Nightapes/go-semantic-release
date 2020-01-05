// Package config provides defimition of .release.yml and read method
package config

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ChangelogConfig struct
type ChangelogConfig struct {
	PrintAll      bool            `yaml:"printAll,omitempty"`
	TemplateTitle string          `yaml:"title,omitempty"`
	TemplatePath  string          `yaml:"templatePath,omitempty"`
	Docker        ChangelogDocker `yaml:"docker,omitempty"`
	NPM           ChangelogNPM    `yaml:"npm,omitempty"`
}

//ChangelogDocker type struct
type ChangelogDocker struct {
	Latest     bool   `yaml:"latest"`
	Repository string `yaml:"repository"`
}

//ChangelogNPM type struct
type ChangelogNPM struct {
	YARN       bool   `yaml:"latest"`
	Repository string `yaml:"repository"`
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

// GitLabProvider struct
type GitLabProvider struct {
	Repo        string `yaml:"repo"`
	CustomURL   string `yaml:"customUrl,omitempty"`
	AccessToken string
}

// GitProvider struct
type GitProvider struct {
	Email    string `yaml:"email"`
	Username string `yaml:"user"`
	Auth     string `yaml:"auth"`
	SSH      bool   `yaml:"ssh"`
}

// Hooks struct
type Hooks struct {
	PreRelease  []string `yaml:"preRelease"`
	PostRelease []string `yaml:"postRelease"`
}

// ReleaseConfig struct
type ReleaseConfig struct {
	CommitFormat   string            `yaml:"commitFormat"`
	Branch         map[string]string `yaml:"branch"`
	Changelog      ChangelogConfig   `yaml:"changelog,omitempty"`
	Release        string            `yaml:"release,omitempty"`
	GitHubProvider GitHubProvider    `yaml:"github,omitempty"`
	GitLabProvider GitLabProvider    `yaml:"gitlab,omitempty"`
	GitProvider    GitProvider       `yaml:"git,omitempty"`
	Assets         []Asset           `yaml:"assets"`
	Hooks          Hooks             `yaml:"hooks"`
	ReleaseTitle   string            `yaml:"title"`
	IsPreRelease   bool
}

// Read ReleaseConfig
func Read(configPath string) (*ReleaseConfig, error) {

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return &ReleaseConfig{}, err
	}

	log.Tracef("Found config %s", string(content))
	releaseConfig := &ReleaseConfig{}
	err = yaml.Unmarshal(content, releaseConfig)
	if err != nil {
		return &ReleaseConfig{}, err
	}

	org := *releaseConfig

	releaseConfig.Hooks = Hooks{}

	configWithoutHooks, err := yaml.Marshal(releaseConfig)
	if err != nil {
		return &ReleaseConfig{}, err
	}
	configWithoutHooks = []byte(os.ExpandEnv(string(configWithoutHooks)))
	releaseConfigWithExpanedEnvs := &ReleaseConfig{}
	err = yaml.Unmarshal(configWithoutHooks, releaseConfigWithExpanedEnvs)
	if err != nil {
		return &ReleaseConfig{}, err
	}

	releaseConfigWithExpanedEnvs.Hooks = org.Hooks

	log.Tracef("Found config %+v", releaseConfigWithExpanedEnvs)

	return releaseConfigWithExpanedEnvs, nil
}
