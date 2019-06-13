// Package semanticrelease provides public methods to include in own code
package semanticrelease

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/cache"
	"github.com/Nightapes/go-semantic-release/internal/changelog"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/internal/releaser"
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
func (s *SemanticRelease) GetNextVersion(repo string, force bool) (string, error) {
	util, err := gitutil.New(repo)
	if err != nil {
		return "", err
	}

	hash, err := util.GetHash()
	if err != nil {
		return "", err
	}

	log.Debugf("Ignore .version file if exits, %t", force)
	if !force {
		content, err := cache.Read()

		if err == nil && content.Commit == hash {
			log.Infof("Found cache, will return cached version %s", content.NextVersion)
			return content.NextVersion, err
		}
		log.Debugf("Mismatch git and version file  %s - %s", content.Commit, hash)
	}

	lastVersion, lastVersionHash, err := util.GetLastVersion()
	if err != nil {
		return "", err
	}
	var newVersion semver.Version

	if lastVersion == nil {
		defaultVersion, _ := semver.NewVersion("1.0.0")
		newVersion = *defaultVersion
		lastVersion = defaultVersion
	} else {
		newVersion = *lastVersion
	}

	commits, err := util.GetCommits(lastVersionHash)
	if err != nil {
		return "", err
	}

	log.Debugf("Found %d commits till last release", len(commits))

	a, err := analyzer.New(s.config.CommitFormat, s.config.Changelog)
	if err != nil {
		return "", err
	}
	result := a.Analyze(commits)

	currentBranch, err := util.GetBranch()
	if err != nil {
		return "", err
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
		return "", err
	}
	c := changelog.New(s.config, a.GetRules())
	c.GenerateChanglog(newVersion.String(), "https://github.com/Nightapes/go-semantic-release/commit/{{hash}}", result)

	return newVersion.String(), err
}

//SetVersion for git repository
func (s *SemanticRelease) SetVersion(version string, repo string) error {

	util, err := gitutil.New(repo)
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

// GetChangelog from last version till now
func (s *SemanticRelease) GetChangelog(repo string) (string, error) {
	nextVersion, err := s.GetNextVersion(repo, false)
	if err != nil {
		log.Debugf("Could not get next version")
		return "", err
	}

	util, err := gitutil.New(repo)
	if err != nil {
		return "", err
	}

	_, lastVersionHash, err := util.GetLastVersion()
	if err != nil {
		return "", err
	}

	commits, err := util.GetCommits(lastVersionHash)
	if err != nil {
		return "", err
	}

	log.Debugf("Found %d commits till last release", len(commits))

	a, err := analyzer.New(s.config.CommitFormat, s.config.Changelog)
	if err != nil {
		return "", err
	}
	result := a.Analyze(commits)

	c := changelog.New(s.config, a.GetRules())
	_, content, err := c.GenerateChanglog(nextVersion, s.config.GetRepositoryURL()+"{{hash}}", result)
	if err != nil {
		return "", err
	}
	return content, nil

}

// WriteChangeLog wirtes changelog content to the given file
func (s *SemanticRelease) WriteChangeLog(changelogContent, file string) error {
	return ioutil.WriteFile(file, []byte(changelogContent), 0644)
}

// Release pusblish release to provider
func (s *SemanticRelease) Release(repo string) error {
	util, err := gitutil.New(repo)
	if err != nil {
		return err
	}
	currentBranch, err := util.GetBranch()

	if _, ok := s.config.Branch[currentBranch]; !ok {
		log.Debugf("Will not perform a new release. Current %s branch is not configured in release config", currentBranch)
		return nil
	}

	nextVersion, err := s.GetNextVersion(repo, false)
	if err != nil {
		log.Debugf("Could not get next version")
		return err
	}

	changelog, err := s.GetChangelog(repo)
	if err != nil {
		log.Debugf("Could not get changelog")
		return err
	}

	releaseTitle := fmt.Sprintf("%s v%s", s.config.ReleaseTitle, nextVersion)

	releaser, err := releaser.New(s.config).GetReleaser()
	if err != nil {
		return err
	}

	if err = releaser.CreateRelease(nextVersion, releaseTitle, changelog, "master"); err != nil {
		return err
	}

	return nil
}
