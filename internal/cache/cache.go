// Package cache helper for cache version
package cache

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// VersionFileContent struct
type VersionFileContent struct {
	Version     string `yaml:"version"`
	NextVersion string `yaml:"nextVersion"`
	Commit      string `yaml:"commit"`
	Branch      string `yaml:"branch"`
}

// Write version into .version
func Write(versionFileContent VersionFileContent) error {
	data, err := yaml.Marshal(&versionFileContent)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(".version", data, 0644)
}

// Read version into .version
func Read() (*VersionFileContent, error) {

	content, err := ioutil.ReadFile(".version")
	if err != nil {
		return &VersionFileContent{}, err
	}

	var versionFileContent VersionFileContent
	err = yaml.Unmarshal(content, &versionFileContent)
	if err != nil {
		return &VersionFileContent{}, err
	}

	return &versionFileContent, nil
}
