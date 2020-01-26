package calculator

import (
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/shared"

	log "github.com/sirupsen/logrus"
)

// Calculator struct
type Calculator struct{}

// New Calculator struct
func New() *Calculator {
	return &Calculator{}
}

//IncPrerelease increase prerelease by one
func (c *Calculator) IncPrerelease(preReleaseType string, version semver.Version) (semver.Version, error) {
	defaultPrerelease := preReleaseType + ".0"
	if !c.hasPrerelease(version, preReleaseType) {
		return version.SetPrerelease(defaultPrerelease)
	}

	parts := strings.Split(version.Prerelease(), ".")
	if len(parts) == 2 {
		i, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Warnf("Could not parse release tag %s, use version %s", version.Prerelease(), version.String())
			return version.SetPrerelease(defaultPrerelease)
		}
		return version.SetPrerelease(preReleaseType + "." + strconv.Itoa(i+1))

	}
	log.Warnf("Could not parse release tag %s, use version %s", version.Prerelease(), version.String())
	return version.SetPrerelease(defaultPrerelease)
}

func (c *Calculator) hasPrerelease(version semver.Version, preReleaseType string) bool {
	if version.Prerelease() == "" || !strings.HasPrefix(version.Prerelease(), preReleaseType) {
		return false
	}
	return true
}

//CalculateNewVersion from given commits and lastversion
func (c *Calculator) CalculateNewVersion(commits map[shared.Release][]shared.AnalyzedCommit, lastVersion *semver.Version, releaseType string, firstRelease bool) semver.Version {
	switch releaseType {
	case "beta", "alpha", "rc":
		var version = *lastVersion
		if !c.hasPrerelease(*lastVersion, releaseType) {
			version, _ = c.inc(commits, lastVersion)
		}

		if len(commits["major"]) > 0 || len(commits["minor"]) > 0 || len(commits["patch"]) > 0 {
			version, _ := c.IncPrerelease(releaseType, version)
			return version
		}
	case "release":
		if !firstRelease {
			if lastVersion.Prerelease() != "" {
				newVersion, _ := lastVersion.SetPrerelease("")
				return newVersion
			}
			version, done := c.inc(commits, lastVersion)
			if done {
				return version
			}
		}
	}

	return *lastVersion
}

func (c *Calculator) inc(commits map[shared.Release][]shared.AnalyzedCommit, lastVersion *semver.Version) (semver.Version, bool) {
	if len(commits["major"]) > 0 {
		return lastVersion.IncMajor(), true
	} else if len(commits["minor"]) > 0 {
		return lastVersion.IncMinor(), true
	} else if len(commits["patch"]) > 0 {
		return lastVersion.IncPatch(), true
	}
	return semver.Version{}, false
}
