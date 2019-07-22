// Package cache helper for cache version
package cache

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/shared"
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
func Write(repository string, releaseVersion shared.ReleaseVersion) error {
	completePath := path.Join(path.Dir(repository), ".version")

	toCache := &ReleaseVersion{
		Next: ReleaseVersionEntry{
			Commit:  releaseVersion.Next.Commit,
			Version: releaseVersion.Next.Version.String(),
		},
		Last: ReleaseVersionEntry{
			Commit:  releaseVersion.Last.Commit,
			Version: releaseVersion.Last.Version.String(),
		},
		Branch: releaseVersion.Branch,
		Draft:  releaseVersion.Draft,
	}

	data, err := yaml.Marshal(toCache)
	if err != nil {
		return err
	}

	log.Debugf("Save %s with hash %s to cache", releaseVersion.Next.Version.String(), releaseVersion.Next.Commit)
	return ioutil.WriteFile(completePath, data, 0644)
}

// Read version into .version
func Read(repository string) (*shared.ReleaseVersion, error) {
	completePath := path.Join(path.Dir(repository), ".version")

	content, err := ioutil.ReadFile(completePath)
	if err != nil {
		return &shared.ReleaseVersion{}, err
	}

	var parsedContent ReleaseVersion
	err = yaml.Unmarshal(content, &parsedContent)
	if err != nil {
		return &shared.ReleaseVersion{}, err
	}

	nextVersion, err := semver.NewVersion(parsedContent.Next.Version)
	if err != nil {
		return nil, err
	}

	lastVersion, err := semver.NewVersion(parsedContent.Last.Version)
	if err != nil {
		return nil, err
	}

	releaseVersion := &shared.ReleaseVersion{
		Next: shared.ReleaseVersionEntry{
			Commit:  parsedContent.Next.Commit,
			Version: nextVersion,
		},
		Last: shared.ReleaseVersionEntry{
			Commit:  parsedContent.Last.Commit,
			Version: lastVersion,
		},
		Branch: parsedContent.Branch,
		Draft:  parsedContent.Draft,
	}
	log.Infof("Found cache, will return cached version %s", parsedContent.Next.Version)
	return releaseVersion, nil
}
