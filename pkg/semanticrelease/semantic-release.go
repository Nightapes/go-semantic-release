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
			fmt.Printf(content.NextVersion)
			return nil
		}
		log.Debugf("Mismatch git and version file  %s - %s", content.Commit, hash)
	}

	lastVersion, lastVersionHash, err := util.GetLastVersion()
	if err != nil {
		return err
	}

	if lastVersion == nil {
		defaultVersion, _ := semver.NewVersion("1.0.0")
		err := s.SetVersion(defaultVersion.String(), repro)
		if err != nil {
			return err
		}
		fmt.Printf("%s", defaultVersion.String())
		return nil
	}

	commits, err := util.GetCommits(lastVersionHash)
	if err != nil {
		return err
	}

	log.Debugf("Found %d commits till last release", len(commits))

	a := analyzer.New("angular")
	result := a.Analyze(commits)

	var newVersion semver.Version

	currentBranch, err := util.GetBranch()
	if err != nil {
		return err
	}
	newVersion = *lastVersion
	if lastVersion.Prerelease() == "" {
		if len(result["major"]) > 0 {
			newVersion = lastVersion.IncMajor()
		} else if len(result["minor"]) > 0 {
			newVersion = lastVersion.IncMinor()
		} else if len(result["patch"]) > 0 {
			newVersion = lastVersion.IncPatch()
		}
	}

	log.Debugf("Test %+v", s.config)
	for branch, releaseType := range s.config.Branch {
		if currentBranch == branch || strings.HasPrefix(currentBranch, branch) {
			log.Debugf("Found branch config for branch %s with release type %s", currentBranch, releaseType)
			switch releaseType {
			case "rc":
				if newVersion.Prerelease() == "" || !strings.HasPrefix(newVersion.Prerelease(), "rc") {
					newVersion, _ = newVersion.SetPrerelease("rc.0")
				} else {
					parts := strings.Split(newVersion.Prerelease(), ".")
					if len(parts) == 2 {
						i, err := strconv.Atoi(parts[1])
						if err != nil {
							newVersion, _ = newVersion.SetPrerelease("rc.0")
							log.Warnf("Could not parse release tag %s, use version %s", newVersion.Prerelease(), newVersion.String())
						} else {
							newVersion, _ = newVersion.SetPrerelease("rc." + strconv.Itoa((i + 1)))
						}
					} else {
						newVersion, _ = newVersion.SetPrerelease("rc.0")
						log.Warnf("Could not parse release tag %s, use version %s", newVersion.Prerelease(), newVersion.String())
					}
				}
			}

		}
	}

	err = s.SetVersion(newVersion.String(), repro)
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
		NextVersion: newVersion.String(),
		Branch:      branch,
	}

	lastVersion, _, err := util.GetLastVersion()
	if err != nil {
		return err
	}

	if lastVersion != nil {
		newVersionContent.Version = lastVersion.String()
	}

	return cache.Write(newVersionContent)
}
