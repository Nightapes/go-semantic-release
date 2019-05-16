// Package semanticrelease provides public methods to include in own code
package semanticrelease

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/cache"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

// SemanticRelease struct
type SemanticRelease struct {
	config *config.ReleaseConfig
}

// New SemanticRelease struct
func New(c *config.ReleaseConfig) *SemanticRelease {
	return &SemanticRelease{
		config: c,
	}
}

// GetNextVersion from .version or calculate new from commits
func (s *SemanticRelease) GetNextVersion(repro string, force bool) error {
	util, err := gitutil.New(repro)
	if err != nil {
		return err
	}

	hash, err := util.GetHash()
	if err != nil {
		return err
	}

	log.Debugf("Ignore .version file if exits, %t", force)
	if !force {
		content, err := cache.Read()

		if err == nil && content.Commit == hash {
			log.Infof("Found cache, will return cached version %s", content.NextVersion)
			fmt.Printf(content.NextVersion)
			return nil
		}
		log.Debugf("Mismatch git and version file  %s - %s", content.Commit, hash)
	}

	lastVersion, lastVersionHash, err := util.GetLastVersion()
	if err != nil {
		return err
	}
	var newVersion semver.Version

	if lastVersion == nil {
		defaultVersion, _ := semver.NewVersion("1.0.0")
		newVersion = *defaultVersion
	} else {
		newVersion = *lastVersion
	}

	commits, err := util.GetCommits(lastVersionHash)
	if err != nil {
		return err
	}

	log.Debugf("Found %d commits till last release", len(commits))

	a := analyzer.New("angular")
	result := a.Analyze(commits)

	currentBranch, err := util.GetBranch()
	if err != nil {
		return err
	}

	for branch, releaseType := range s.config.Branch {
		if currentBranch == branch || strings.HasPrefix(currentBranch, branch) {
			log.Debugf("Found branch config for branch %s with release type %s", currentBranch, releaseType)
			switch releaseType {
			case "rc", "beta", "alpha":
				newVersion = incPrerelease(releaseType, newVersion)
			case "release":
				if len(result["major"]) > 0 {
					newVersion = newVersion.IncMajor()
				} else if len(result["minor"]) > 0 {
					newVersion = newVersion.IncMinor()
				} else if len(result["patch"]) > 0 {
					newVersion = newVersion.IncPatch()
				}
			}
		}
	}

	log.Infof("New version %s -> %s", lastVersion.String(), newVersion.String())
	err = saveToCache(util, lastVersion, &newVersion)
	if err != nil {
		return err
	}
	fmt.Printf("%s", newVersion.String())

	return err
}

//SetVersion for git repository
func (s *SemanticRelease) SetVersion(version string, repro string) error {

	util, err := gitutil.New(repro)
	if err != nil {
		return err
	}

	newVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	lastVersion, _, err := util.GetLastVersion()
	if err != nil {
		return err
	}

	return saveToCache(util, lastVersion, newVersion)
}

func saveToCache(util *gitutil.GitUtil, lastVersion *semver.Version, nextVersion *semver.Version) error {

	hash, err := util.GetHash()
	if err != nil {
		return err
	}

	branch, err := util.GetBranch()
	if err != nil {
		return err
	}

	newVersionContent := cache.VersionFileContent{
		Commit:      hash,
		NextVersion: nextVersion.String(),
		Branch:      branch,
	}

	if lastVersion != nil {
		newVersionContent.Version = lastVersion.String()
	}

	log.Debugf("Save %s with hash %s to cache", nextVersion.String(), hash)
	return cache.Write(newVersionContent)
}

func incPrerelease(preReleaseType string, version semver.Version) semver.Version {
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
