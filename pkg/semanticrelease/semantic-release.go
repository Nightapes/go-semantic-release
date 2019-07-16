package semanticrelease

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/changelog"
	"github.com/Nightapes/go-semantic-release/internal/ci"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/internal/releaser"
	"github.com/Nightapes/go-semantic-release/internal/releaser/util"
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
	provider, err := ci.GetCIProvider(s.gitutil, ci.ReadAllEnvs())

	if err != nil {
		fakeVersion, _ := semver.NewVersion("0.0.0-fake.0")
		log.Warnf("Will not calculate version, set fake version. Could not find CI Provider, if running locally, set env CI=true")
		return &shared.ReleaseVersion{
			Next: shared.ReleaseVersionEntry{
				Commit:  "",
				Version: fakeVersion,
			},
		}, nil
	}

	log.Debugf("Ignore .version file if exits, %t", force)
	if !force {
		releaseVersion, err := s.readFromCache(provider.Commit)
		if err != nil {
			return nil, err
		}

		if releaseVersion != nil {
			return releaseVersion, nil
		}
	}

	lastVersion, lastVersionHash, err := s.gitutil.GetLastVersion()
	if err != nil {
		return nil, err
	}

	var newVersion semver.Version
	firstRelease := false

	if lastVersion == nil {
		defaultVersion, _ := semver.NewVersion("1.0.0")
		newVersion = *defaultVersion
		lastVersion = defaultVersion
		firstRelease = true
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
	isDraft := false
	for branch, releaseType := range s.config.Branch {
		if provider.Branch == branch || strings.HasPrefix(provider.Branch, branch) {
			log.Debugf("Found branch config for branch %s with release type %s", provider.Branch, releaseType)
			switch releaseType {
			case "beta", "alpha":
				isDraft = true
				newVersion = s.incPrerelease(releaseType, newVersion)
			case "rc":
				newVersion = s.incPrerelease(releaseType, newVersion)
			case "release":
				if !firstRelease {
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
	}

	releaseVersion := shared.ReleaseVersion{
		Next: shared.ReleaseVersionEntry{
			Commit:  provider.Commit,
			Version: &newVersion,
		},
		Last: shared.ReleaseVersionEntry{
			Commit:  lastVersionHash,
			Version: lastVersion,
		},
		Branch: provider.Branch,
		Draft:  isDraft,
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

	provider, err := ci.GetCIProvider(s.gitutil, ci.ReadAllEnvs())

	if err != nil {
		return fmt.Errorf("will not set version. Could not find CI Provider, if running locally, set env CI=true")
	}

	lastVersion, lastVersionHash, err := s.gitutil.GetLastVersion()
	if err != nil {
		return err
	}
	if lastVersion == nil {
		lastVersion, _ = semver.NewVersion("1.0.0")
	}

	return s.saveToCache(shared.ReleaseVersion{
		Next: shared.ReleaseVersionEntry{
			Commit:  provider.Commit,
			Version: newVersion,
		},
		Last: shared.ReleaseVersionEntry{
			Commit:  lastVersionHash,
			Version: lastVersion,
		},
		Branch: provider.Branch,
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

	provider, err := ci.GetCIProvider(s.gitutil, ci.ReadAllEnvs())

	if err != nil {
		log.Debugf("Will not perform a new release. Could not find CI Provider")
		return nil
	}

	if provider.IsPR {
		log.Debugf("Will not perform a new release. This is a pull request")
		return nil
	}

	if _, ok := s.config.Branch[provider.Branch]; !ok {
		log.Debugf("Will not perform a new release. Current %s branch is not configured in release config", provider.Branch)
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

	if err = releaser.UploadAssets(s.repository, s.config.Assets); err != nil {
		return err
	}

	return nil
}

// ZipFiles zip files configured in release config
func (s *SemanticRelease) ZipFiles() error {
	for _, file := range s.config.Assets {
		if file.Compress {
			if _, err := util.PrepareAssets(s.repository, s.config.Assets); err != nil {
				return err
			}
		}
	}
	return nil
}
