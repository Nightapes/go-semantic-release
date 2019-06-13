// Package config provides defimition of .release.yml and read method
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// List of all supported git providers
var gitPorviders = map[string]string{"GitHub": "", "GitLab": ""}

// ChangelogConfig struct
type ChangelogConfig struct {
	PrintAll     bool   `yaml:"printAll,omitempty"`
	Template     string `yaml:"template,omitempty"`
	TemplatePath string `yaml:"templatePath,omitempty"`
}

// GitProvider struct
type GitProvider struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	User        string `yaml:"user"`
	AccessToken string `yaml:"accessToken"`
}

type Asset struct {
	Name     string `yaml:"name"`
	Compress bool   `yaml:"compress"`
}

// ReleaseConfig struct
type ReleaseConfig struct {
	CommitFormat          string            `yaml:"commitFormat"`
	Branch                map[string]string `yaml:"branch"`
	Changelog             ChangelogConfig   `yaml:"changelog,omitempty"`
	Release               string            `yaml:"release,omitempty"`
	GitProvider           GitProvider       `yaml:"provider,omitempty"`
	Assets                []Asset           `yaml:"assets"`
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

	releaseConfig, err = checkProvider(releaseConfig)
	if err != nil {
		return &ReleaseConfig{}, err
	}
	return &releaseConfig, nil
}

func checkProvider(config ReleaseConfig) (ReleaseConfig, error) {
	if config.GitProvider != (GitProvider{}) {
		if _, ok := gitPorviders[config.GitProvider.Name]; !ok {
			return ReleaseConfig{}, fmt.Errorf("config: provider: configured provider %s is not supported", config.GitProvider.Name)
		}
		envName := fmt.Sprintf("%s_ACCESS_TOKEN", strings.ToUpper(config.GitProvider.Name))

		token, isSet := os.LookupEnv(envName)
		if !isSet {
			return ReleaseConfig{}, fmt.Errorf("config: Can not find environment variable %s", token)
		}
		config.GitProvider.AccessToken = token
	}
	return config, nil
}
