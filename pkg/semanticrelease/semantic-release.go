package semanticrelease

import (
	"io/ioutil"
	"time"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/cache"
	"github.com/Nightapes/go-semantic-release/internal/calculator"
	"github.com/Nightapes/go-semantic-release/internal/changelog"
	"github.com/Nightapes/go-semantic-release/internal/ci"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/internal/hooks"
	"github.com/Nightapes/go-semantic-release/internal/releaser"
	"github.com/Nightapes/go-semantic-release/internal/releaser/util"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

// SemanticRelease struct
type SemanticRelease struct {
	config      *config.ReleaseConfig
	gitutil     *gitutil.GitUtil
	analyzer    *analyzer.Analyzer
	calculator  *calculator.Calculator
	releaser    releaser.Releaser
	repository  string
	checkConfig bool
}

// New SemanticRelease struct
func New(c *config.ReleaseConfig, repository string, checkConfig bool) (*SemanticRelease, error) {
	util, err := gitutil.New(repository)
	if err != nil {
		return nil, err
	}

	analyzer, err := analyzer.New(c.CommitFormat, c.Changelog)
	if err != nil {
		return nil, err
	}

	if !checkConfig {
		log.Infof("Ignore config checks!. No guarantee to run without issues")
	}

	releaser, err := releaser.New(c, util).GetReleaser(checkConfig)
	if err != nil {
		return nil, err
	}

	return &SemanticRelease{
		config:      c,
		gitutil:     util,
		releaser:    releaser,
		analyzer:    analyzer,
		repository:  repository,
		checkConfig: checkConfig,
		calculator:  calculator.New(),
	}, nil
}

//GetCIProvider result with ci config
func (s *SemanticRelease) GetCIProvider() (*ci.ProviderConfig, error) {
	return ci.GetCIProvider(s.gitutil, s.checkConfig, ci.ReadAllEnvs())
}

// GetNextVersion from .version or calculate new from commits
func (s *SemanticRelease) GetNextVersion(provider *ci.ProviderConfig, force bool) (*shared.ReleaseVersion, error) {
	log.Debugf("Ignore .version file if exits, %t", force)
	if !force {
		releaseVersion, err := cache.Read(s.repository)
		if err != nil {
			return nil, err
		}

		if releaseVersion.Next.Commit == provider.Commit && releaseVersion != nil {
			return releaseVersion, nil
		}
	}

	lastVersion, lastVersionHash, err := s.gitutil.GetLastVersion()
	if err != nil {
		return nil, err
	}

	firstRelease := false

	if lastVersion == nil {
		lastVersion, _ = semver.NewVersion("1.0.0")
		log.Infof("This is the first release, will set version to %s", lastVersion.String())
		firstRelease = true
	}

	commits, err := s.gitutil.GetCommits(lastVersionHash)
	if err != nil {
		return nil, err
	}

	log.Debugf("Found %d commits till last release", len(commits))

	analyzedCommits := s.analyzer.Analyze(commits)

	var newVersion semver.Version
	foundBranchConfig := false
	for branch, releaseType := range s.config.Branch {
		if provider.Branch == branch {
			log.Debugf("Found branch config for branch %s with release type %s", provider.Branch, releaseType)
			newVersion = s.calculator.CalculateNewVersion(analyzedCommits, lastVersion, releaseType, firstRelease)
			foundBranchConfig = true
			break
		}
	}

	if !foundBranchConfig {
		log.Warnf("No branch config found for branch %s, will return last known version", provider.Branch)
		newVersion = *lastVersion
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
		Branch:  provider.Branch,
		Commits: analyzedCommits,
	}

	if firstRelease {
		releaseVersion.Last.Version, _ = semver.NewVersion("0.0.0")
	}

	log.Infof("New version %s -> %s", lastVersion.String(), newVersion.String())
	err = cache.Write(s.repository, releaseVersion)
	if err != nil {
		return nil, err
	}
	return &releaseVersion, err
}

//SetVersion for git repository
func (s *SemanticRelease) SetVersion(provider *ci.ProviderConfig, version string) error {

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

	return cache.Write(s.repository, shared.ReleaseVersion{
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
	c := changelog.New(s.config, s.analyzer.GetRules(), time.Now())
	return c.GenerateChanglog(shared.ChangelogTemplateConfig{
		Version:    releaseVersion.Next.Version.String(),
		Hash:       releaseVersion.Last.Commit,
		CommitURL:  s.releaser.GetCommitURL(),
		CompareURL: s.releaser.GetCompareURL(releaseVersion.Last.Version.String(), releaseVersion.Next.Version.String()),
	}, releaseVersion.Commits)

}

// WriteChangeLog wirtes changelog content to the given file
func (s *SemanticRelease) WriteChangeLog(changelogContent, file string) error {
	return ioutil.WriteFile(file, []byte(changelogContent), 0644)
}

// Release publish release to provider
func (s *SemanticRelease) Release(provider *ci.ProviderConfig, force bool) error {

	if provider.IsPR {
		log.Infof("Will not perform a new release. This is a pull request")
		return nil
	}

	if _, ok := s.config.Branch[provider.Branch]; !ok {
		log.Infof("Will not perform a new release. Current %s branch is not configured in release config", provider.Branch)
		return nil
	}

	releaseVersion, err := s.GetNextVersion(provider, force)
	if err != nil {
		log.Debugf("Could not get next version")
		return err
	}

	if releaseVersion.Next.Version.Equal(releaseVersion.Last.Version) {
		log.Infof("No new version, no release needed %s <> %s", releaseVersion.Next.Version.String(), releaseVersion.Last.Version.String())
		return nil
	}

	hook := hooks.New(s.config, releaseVersion)

	generatedChangelog, err := s.GetChangelog(releaseVersion)
	if err != nil {
		log.Debugf("Could not get changelog")
		return err
	}

	err = hook.PreRelease()
	if err != nil {
		log.Debugf("Error during pre release hook")
		return err
	}

	if err = s.releaser.CreateRelease(releaseVersion, generatedChangelog); err != nil {
		return err
	}

	if err = s.releaser.UploadAssets(s.repository, s.config.Assets); err != nil {
		return err
	}

	err = hook.PostRelease()
	if err != nil {
		log.Debugf("Error during post release hook")
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
