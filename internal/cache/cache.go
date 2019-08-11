// Package cache helper for cache version
package cache

import (
	"io/ioutil"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"gopkg.in/yaml.v2"
)

// Write version into .version
func Write(repository string, releaseVersion shared.ReleaseVersion) error {
	completePath := path.Join(path.Dir(repository), ".version")

	if releaseVersion.Last.Version != nil {
		releaseVersion.Last.VersionString = releaseVersion.Last.Version.String()
	}

	if releaseVersion.Next.Version != nil {
		releaseVersion.Next.VersionString = releaseVersion.Next.Version.String()
	}

	//toCache := &ReleaseVersion(releaseVersion)
	data, err := yaml.Marshal(releaseVersion)
	if err != nil {
		return err
	}

	log.Infof("Save %s with hash %s to cache %s", releaseVersion.Next.Version.String(), releaseVersion.Next.Commit, completePath)
	return ioutil.WriteFile(completePath, data, 0644)
}

// Read version into .version
func Read(repository string) (*shared.ReleaseVersion, error) {
	completePath := path.Join(path.Dir(repository), ".version")

	content, err := ioutil.ReadFile(completePath)
	if err != nil {
		log.Warnf("Could not read cache %s, will ignore cache", completePath)
		return &shared.ReleaseVersion{}, nil
	}

	var parsedContent shared.ReleaseVersion
	err = yaml.Unmarshal(content, &parsedContent)
	if err != nil {
		return &shared.ReleaseVersion{}, err
	}

	parsedContent.Next.Version, err = semver.NewVersion(parsedContent.Next.VersionString)
	if err != nil {
		return nil, err
	}

	parsedContent.Last.Version, err = semver.NewVersion(parsedContent.Last.VersionString)
	if err != nil {
		return nil, err
	}

	log.Infof("Found cache, will return cached version %s", parsedContent.Next.Version)
	return &parsedContent, nil
}
