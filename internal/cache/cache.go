// Package cache helper for cache version
package cache

import (
	"io/ioutil"
	"path"

	"gopkg.in/yaml.v2"
)

// ReleaseVersion struct
type ReleaseVersion struct {
	Last   ReleaseVersionEntry `yaml:"last"`
	Next   ReleaseVersionEntry `yaml:"next"`
	Branch string              `yaml:"branch"`
	Draft  bool                `yaml:"draft"`
}

//ReleaseVersionEntry struct
type ReleaseVersionEntry struct {
	Commit  string `yaml:"commit"`
	Version string `yaml:"version"`
}

// Write version into .version
func Write(repository string, versionFileContent ReleaseVersion) error {
	completePath := path.Join(path.Dir(repository), ".version")

	data, err := yaml.Marshal(&versionFileContent)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(completePath, data, 0644)
}

// Read version into .version
func Read(repository string) (*ReleaseVersion, error) {
	completePath := path.Join(path.Dir(repository), ".version")

	content, err := ioutil.ReadFile(completePath)
	if err != nil {
		return &ReleaseVersion{}, err
	}

	var versionFileContent ReleaseVersion
	err = yaml.Unmarshal(content, &versionFileContent)
	if err != nil {
		return &ReleaseVersion{}, err
	}

	return &versionFileContent, nil
}
