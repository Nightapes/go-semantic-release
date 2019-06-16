package semanticrelease

import (
	"io/ioutil"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/changelog"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/internal/releaser"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

// SemanticRelease struct
type SemanticRelease struct {
	config     *config.ReleaseConfig
	gitutil    *gitutil.GitUtil
	analyzer   *analyzer.Analyzer
	releaser   releaser.Releaser
	repository string
}

// New SemanticRelease struct
func New(c *config.ReleaseConfig, repository string) (*SemanticRelease, error) {
	util, err := gitutil.New(repository)
	if err != nil {
		return nil, err
	}

	analyzer, err := analyzer.New(c.CommitFormat, c.Changelog)
	if err != nil {
		return nil, err
	}

	releaser, err := releaser.New(c).GetReleaser()
	if err != nil {
		return nil, err
	}

	return &SemanticRelease{
		config:     c,
		gitutil:    util,
		releaser:   releaser,
		analyzer:   analyzer,
		repository: repository,
	}, nil
}

// GetNextVersion from .version or calculate new from commits
func (s *SemanticRelease) GetNextVersion(force bool) (*shared.ReleaseVersion, error) {
	hash, err := s.gitutil.GetHash()
	if err != nil {
		return nil, err
	}

	log.Debugf("Ignore .version file if exits, %t", force)
	if !force {
		releaseVersion, err := s.readFromCache(hash)
		if err != nil {
			return nil, err
		}

		if releaseVersion != nil {
			return releaseVersion, nil
		}
	}

	currentBranch, err := s.gitutil.GetBranch()
	if err != nil {
		return nil, err
	}

	lastVersion, lastVersionHash, err := s.gitutil.GetLastVersion()
	if err != nil {
		return nil, err
	}

	var newVersion semver.Version

	if lastVersion == nil {
		defaultVersion, _ := semver.NewVersion("1.0.0")
		newVersion = *defaultVersion
		lastVersion = defaultVersion
	} else {
		newVersion = *lastVersion
	}

	commits, err := s.gitutil.GetCommits(lastVersionHash)
	if err != nil {
		return nil, err
	}

	log.Debugf("Found %d commits till last release", len(commits))

	a, err := analyzer.New(s.config.CommitFormat, s.config.Changelog)
	if err != nil {
		return nil, err
	}
	result := a.Analyze(commits)

	for branch, releaseType := range s.config.Branch {
		if currentBranch == branch || strings.HasPrefix(currentBranch, branch) {
			log.Debugf("Found branch config for branch %s with release type %s", currentBranch, releaseType)
			switch releaseType {
			case "rc", "beta", "alpha":
				newVersion = s.incPrerelease(releaseType, newVersion)
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

	releaseVersion := shared.ReleaseVersion{
		Next: shared.ReleaseVersionEntry{
			Commit:  hash,
			Version: &newVersion,
		},
		Last: shared.ReleaseVersionEntry{
			Commit:  lastVersionHash,
			Version: lastVersion,
		},
		Branch: currentBranch,
	}

	log.Infof("New version %s -> %s", lastVersion.String(), newVersion.String())
	err = s.saveToCache(releaseVersion)
	if err != nil {
		return nil, err
	}
	return &releaseVersion, err
}

//SetVersion for git repository
func (s *SemanticRelease) SetVersion(version string) error {

	newVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	lastVersion, lastVersionHash, err := s.gitutil.GetLastVersion()
	if err != nil {
		return err
	}
	if lastVersion == nil {
		lastVersion, _ = semver.NewVersion("1.0.0")
	}

	hash, err := s.gitutil.GetHash()
	if err != nil {
		return err
	}

	currentBranch, err := s.gitutil.GetBranch()
	if err != nil {
		return err
	}

	return s.saveToCache(shared.ReleaseVersion{
		Next: shared.ReleaseVersionEntry{
			Commit:  hash,
			Version: newVersion,
		},
		Last: shared.ReleaseVersionEntry{
			Commit:  lastVersionHash,
			Version: lastVersion,
		},
		Branch: currentBranch,
	})
}

// GetChangelog from last version till now
func (s *SemanticRelease) GetChangelog(releaseVersion *shared.ReleaseVersion) (*shared.GeneratedChangelog, error) {
	commits, err := s.gitutil.GetCommits(releaseVersion.Last.Commit)
	if err != nil {
		return nil, err
	}

	result := s.analyzer.Analyze(commits)

	log.Debugf("Found %d commits till last release", len(commits))

	c := changelog.New(s.config, s.analyzer.GetRules())
	return c.GenerateChanglog(shared.ChangelogTemplateConfig{
		Version:    releaseVersion.Next.Version.String(),
		Hash:       releaseVersion.Last.Commit,
		CommitURL:  s.releaser.GetCommitURL(),
		CompareURL: s.releaser.GetCompareURL(releaseVersion.Last.Version.String(), releaseVersion.Next.Version.String()),
	}, result)

}

// WriteChangeLog wirtes changelog content to the given file
func (s *SemanticRelease) WriteChangeLog(changelogContent, file string) error {
	return ioutil.WriteFile(file, []byte(changelogContent), 0644)
}

// Release pusblish release to provider
func (s *SemanticRelease) Release(force bool) error {
	currentBranch, err := s.gitutil.GetBranch()
	if err != nil {
		return err
	}

	if _, ok := s.config.Branch[currentBranch]; !ok {
		log.Debugf("Will not perform a new release. Current %s branch is not configured in release config", currentBranch)
		return nil
	}

	releaseVersion, err := s.GetNextVersion(force)
	if err != nil {
		log.Debugf("Could not get next version")
		return err
	}

	generatedChanglog, err := s.GetChangelog(releaseVersion)
	if err != nil {
		log.Debugf("Could not get changelog")
		return err
	}

	releaser, err := releaser.New(s.config).GetReleaser()
	if err != nil {
		return err
	}

	err = releaser.ValidateConfig()
	if err != nil {
		return err
	}

	if err = releaser.CreateRelease(releaseVersion, generatedChanglog); err != nil {
		return err
	}

	return nil
}
