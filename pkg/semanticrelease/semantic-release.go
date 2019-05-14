// Package semanticrelease provides public methods to include in own code
package semanticrelease

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/internal/storage"
	log "github.com/sirupsen/logrus"
)

// GetNextVersion from .version or calculate new from commits
func GetNextVersion(repro string) error {
	util, err := gitutil.New(repro)
	if err != nil {
		return err
	}

	hash, err := util.GetHash()
	if err != nil {
		return err
	}

	content, err := storage.Read()

	if err == nil && content.Commit == hash {
		fmt.Printf(content.NextVersion)
		return nil
	}

	log.Debugf("Mismatch git and version file  %s - %s", content.Commit, hash)

	lastVersion, lastVersionHash, err := util.GetLastVersion()
	if err != nil {
		return err
	}

	if lastVersion == nil {
		defaultVersion, _ := semver.NewVersion("1.0.0")
		err := SetVersion(defaultVersion.String(), repro)
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

	if len(result["major"]) > 0 {
		newVersion = lastVersion.IncMajor()
		return nil
	} else if len(result["minor"]) > 0 {
		newVersion = lastVersion.IncMinor()
	} else if len(result["patch"]) > 0 {
		newVersion = lastVersion.IncPatch()
	}

	err = SetVersion(newVersion.String(), repro)
	if err != nil {
		return err
	}
	fmt.Printf("%s", newVersion.String())

	return err
}

//SetVersion for git repository
func SetVersion(version string, repro string) error {

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

	newVersionContent := storage.VersionFileContent{
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

	return storage.Write(newVersionContent)
}
