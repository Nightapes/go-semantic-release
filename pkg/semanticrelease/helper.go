package semanticrelease

import (
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/cache"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	log "github.com/sirupsen/logrus"
)

func (s *SemanticRelease) incPrerelease(preReleaseType string, version semver.Version) semver.Version {
	defaultPrerelease := preReleaseType + ".0"
	if version.Prerelease() == "" || !strings.HasPrefix(version.Prerelease(), preReleaseType) {
		version, _ = version.SetPrerelease(defaultPrerelease)
	} else {
		parts := strings.Split(version.Prerelease(), ".")
		if len(parts) == 2 {
			i, err := strconv.Atoi(parts[1])
			if err != nil {
				version, _ = version.SetPrerelease(defaultPrerelease)
				log.Warnf("Could not parse release tag %s, use version %s", version.Prerelease(), version.String())
			} else {
				version, _ = version.SetPrerelease(preReleaseType + "." + strconv.Itoa((i + 1)))
			}
		} else {
			version, _ = version.SetPrerelease(defaultPrerelease)
			log.Warnf("Could not parse release tag %s, use version %s", version.Prerelease(), version.String())
		}
	}

	return version
}

func (s *SemanticRelease) saveToCache(releaseVersion shared.ReleaseVersion) error {

	toCache := cache.ReleaseVersion{
		Next: cache.ReleaseVersionEntry{
			Commit:  releaseVersion.Next.Commit,
			Version: releaseVersion.Next.Version.String(),
		},
		Last: cache.ReleaseVersionEntry{
			Commit:  releaseVersion.Last.Commit,
			Version: releaseVersion.Last.Version.String(),
		},
		Branch: releaseVersion.Branch,
	}

	log.Debugf("Save %s with hash %s to cache", releaseVersion.Next.Version.String(), releaseVersion.Next.Commit)
	return cache.Write(s.repository, toCache)
}

func (s *SemanticRelease) readFromCache(currentHash string) (*shared.ReleaseVersion, error) {
	content, err := cache.Read(s.repository)

	if err == nil && content.Next.Commit == currentHash {

		nextVersion, err := semver.NewVersion(content.Next.Version)
		if err != nil {
			return nil, err
		}

		lastVersion, err := semver.NewVersion(content.Last.Version)
		if err != nil {
			return nil, err
		}

		releaseVersion := &shared.ReleaseVersion{
			Next: shared.ReleaseVersionEntry{
				Commit:  content.Next.Commit,
				Version: nextVersion,
			},
			Last: shared.ReleaseVersionEntry{
				Commit:  content.Last.Commit,
				Version: lastVersion,
			},
			Branch: content.Branch,
		}

		log.Infof("Found cache, will return cached version %s", content.Next.Version)
		return releaseVersion, nil

	}
	log.Debugf("Mismatch git and version file  %s - %s", content.Next.Commit, currentHash)
	return nil, nil
}
