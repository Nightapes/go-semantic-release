package semanticrelease

import (
	"github.com/Nightapes/go-semantic-release/internal/integrations"
	"io/ioutil"
	"time"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/assets"
	"github.com/Nightapes/go-semantic-release/internal/cache"
	"github.com/Nightapes/go-semantic-release/internal/calculator"
	"github.com/Nightapes/go-semantic-release/internal/changelog"
	"github.com/Nightapes/go-semantic-release/internal/ci"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/internal/hooks"
	"github.com/Nightapes/go-semantic-release/internal/releaser"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
)

// SemanticRelease struct
type SemanticRelease struct {
	config      *config.ReleaseConfig
	gitUtil     *gitutil.GitUtil
	analyzer    *analyzer.Analyzer
	calculator  *calculator.Calculator
	releaser    releaser.Releaser
	assets      *assets.Set
	repository  string
	checkConfig bool
}

// New SemanticRelease struct
func New(c *config.ReleaseConfig, repository string, checkConfig bool) (*SemanticRelease, error) {
	util, err := gitutil.New(repository)
	if err != nil {
		return nil, err
	}

	analyzer, err := analyzer.New(c.CommitFormat, c.Analyzer, c.Changelog)
	if err != nil {
		return nil, err
	}

	if !checkConfig {
		log.Infof("Ignore config checks!. No guarantee to run without issues")
	}

	assets := assets.New(repository, c.Checksum.Algorithm)

	releaser, err := releaser.New(c, util).GetReleaser(checkConfig)
	if err != nil {
		return nil, err
	}

	return &SemanticRelease{
		config:      c,
		gitUtil:     util,
		releaser:    releaser,
		analyzer:    analyzer,
		repository:  repository,
		assets:      assets,
		checkConfig: checkConfig,
		calculator:  calculator.New(),
	}, nil
}

// GetCIProvider result with ci config
func (s *SemanticRelease) GetCIProvider() (*ci.ProviderConfig, error) {
	return ci.GetCIProvider(s.gitUtil, s.checkConfig, ci.ReadAllEnvs())
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

	lastVersion, lastVersionHash, err := s.gitUtil.GetLastVersion()
	if err != nil {
		return nil, err
	}

	firstRelease := false

	if lastVersion == nil {
		lastVersion, _ = semver.NewVersion("1.0.0")
		log.Infof("This is the first release, will set version to %s", lastVersion.String())
		firstRelease = true
	}

	commits, err := s.gitUtil.GetCommits(lastVersionHash)
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

// SetVersion for git repository
func (s *SemanticRelease) SetVersion(provider *ci.ProviderConfig, version string) error {
	newVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	lastVersion, lastVersionHash, err := s.gitUtil.GetLastVersion()
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
	return c.GenerateChangelog(shared.ChangelogTemplateConfig{
		Version:    releaseVersion.Next.Version.String(),
		Hash:       releaseVersion.Last.Commit,
		CommitURL:  s.releaser.GetCommitURL(),
		CompareURL: s.releaser.GetCompareURL(releaseVersion.Last.Version.String(), releaseVersion.Next.Version.String()),
	}, releaseVersion.Commits)
}

// WriteChangeLog writes changelog content to the given file
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

	if err := s.assets.Add(s.config.Assets...); err != nil {
		return err
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

	generatedChangelog, err := s.GetChangelog(releaseVersion)
	if err != nil {
		log.Debugf("Could not get changelog")
		return err
	}

	integrations := integrations.New(&s.config.Integrations, releaseVersion)
	if err := integrations.Run(); err != nil {
		log.Debugf("Error during integrations run")
		return err
	}

	hook := hooks.New(s.config, releaseVersion)
	if err := hook.PreRelease(); err != nil {
		log.Debugf("Error during pre release hook")
		return err
	}

	if s.config.Checksum.Algorithm != "" {
		if err := s.assets.GenerateChecksum(); err != nil {
			return err
		}
	}

	for _, asset := range s.assets.All() {
		if asset.IsCompressed() {
			if _, err := asset.ZipFile(); err != nil {
				return err
			}
		}
	}

	if err = s.releaser.CreateRelease(releaseVersion, generatedChangelog, s.assets); err != nil {
		return err
	}

	if err := hook.PostRelease(); err != nil {
		log.Debugf("Error during post release hook")
		return err
	}

	return nil
}

// ZipFiles zip files configured in release config
func (s *SemanticRelease) ZipFiles() error {
	assets := assets.New(s.repository, "")

	if err := assets.Add(s.config.Assets...); err != nil {
		return err
	}
	if err := assets.GenerateChecksum(); err != nil {
		return err
	}

	for _, asset := range assets.All() {
		path, err := asset.GetPath()
		if err != nil {
			return err
		}
		log.Infof("File %s under %s is zipped %t", asset.GetName(), path, asset.IsCompressed())
	}
	return nil
}
