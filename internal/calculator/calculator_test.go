package calculator_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/calculator"
	"github.com/stretchr/testify/assert"
)

func createVersion(version string) *semver.Version {
	ver, _ := semver.NewVersion(version)
	return ver
}

func TestCalculator_IncPrerelease(t *testing.T) {

	testConfigs := []struct {
		testCase       string
		preReleaseType string
		lastVersion    *semver.Version
		nextVersion    string
		hasError       bool
	}{
		{
			testCase:       "version without preRelease",
			preReleaseType: "alpha",
			lastVersion:    createVersion("1.0.0"),
			nextVersion:    "1.0.0-alpha.0",
		},
		{
			testCase:       "version with preRelease",
			preReleaseType: "alpha",
			lastVersion:    createVersion("1.0.0-alpha.0"),
			nextVersion:    "1.0.0-alpha.1",
		},
		{
			testCase:       "version with preRelease, change type",
			preReleaseType: "beta",
			lastVersion:    createVersion("1.0.0-alpha.0"),
			nextVersion:    "1.0.0-beta.0",
		},
		{
			testCase:       "version with preRelease but broken",
			preReleaseType: "alpha",
			lastVersion:    createVersion("1.0.0-alpha.br0ken"),
			nextVersion:    "1.0.0-alpha.0",
		},
		{
			testCase:       "version with preRelease but broken 2",
			preReleaseType: "alpha",
			lastVersion:    createVersion("1.0.0-alphabr0ken"),
			nextVersion:    "1.0.0-alpha.0",
		},
	}

	c := calculator.New()

	for _, test := range testConfigs {
		next, err := c.IncPrerelease(test.preReleaseType, test.lastVersion)
		assert.Equalf(t, test.hasError, err != nil, "Testcase %s should have error: %t -> %s", test.testCase, test.hasError, err)
		assert.Equal(t, test.nextVersion, next.String())
	}

}

func TestCalculator_CalculateNewVersion(t *testing.T) {

	testConfigs := []struct {
		testCase        string
		releaseType     string
		lastVersion     *semver.Version
		nextVersion     string
		isDraft         bool
		isFirst         bool
		analyzedCommits map[analyzer.Release][]analyzer.AnalyzedCommit
	}{
		{
			testCase:    "version with preRelease alpha",
			releaseType: "alpha",
			lastVersion: createVersion("1.0.0"),
			nextVersion: "1.0.0-alpha.0",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{},
				"minor": []analyzer.AnalyzedCommit{
					analyzer.AnalyzedCommit{},
				},
				"patch": []analyzer.AnalyzedCommit{},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: false,
			isDraft: true,
		},
		{
			testCase:    "version with preRelease beta",
			releaseType: "beta",
			lastVersion: createVersion("1.0.0"),
			nextVersion: "1.0.0-beta.0",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{},
				"minor": []analyzer.AnalyzedCommit{
					analyzer.AnalyzedCommit{},
				},
				"patch": []analyzer.AnalyzedCommit{},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: false,
			isDraft: true,
		},
		{
			testCase:    "version without commits",
			releaseType: "alpha",
			lastVersion: createVersion("1.0.0"),
			nextVersion: "1.0.0",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{},
				"minor": []analyzer.AnalyzedCommit{},
				"patch": []analyzer.AnalyzedCommit{},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: false,
			isDraft: false,
		},
		{
			testCase:    "version with commits and first release",
			releaseType: "release",
			lastVersion: createVersion("1.0.0"),
			nextVersion: "1.0.0",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"minor": []analyzer.AnalyzedCommit{},
				"patch": []analyzer.AnalyzedCommit{},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: true,
			isDraft: false,
		},
		{
			testCase:    "version with commits and rc release",
			releaseType: "rc",
			lastVersion: createVersion("1.0.0"),
			nextVersion: "1.0.0-rc.0",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"minor": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"patch": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: false,
			isDraft: false,
		},
		{
			testCase:    "version with commits and rc release",
			releaseType: "rc",
			lastVersion: createVersion("1.0.0-rc.0"),
			nextVersion: "1.0.0-rc.1",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{},
				"minor": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"patch": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: false,
			isDraft: false,
		},
		{
			testCase:    "version with commits and major release",
			releaseType: "release",
			lastVersion: createVersion("1.0.0"),
			nextVersion: "2.0.0",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"minor": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"patch": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: false,
			isDraft: false,
		},
		{
			testCase:    "version with commits and minor release",
			releaseType: "release",
			lastVersion: createVersion("1.0.0"),
			nextVersion: "1.1.0",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{},
				"minor": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"patch": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: false,
			isDraft: false,
		},
		{
			testCase:    "version with commits and minor patch",
			releaseType: "release",
			lastVersion: createVersion("1.0.0"),
			nextVersion: "1.0.1",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"major": []analyzer.AnalyzedCommit{},
				"minor": []analyzer.AnalyzedCommit{},
				"patch": []analyzer.AnalyzedCommit{analyzer.AnalyzedCommit{}},
				"none":  []analyzer.AnalyzedCommit{},
			},
			isFirst: false,
			isDraft: false,
		},
	}

	c := calculator.New()

	for _, test := range testConfigs {
		next, draft := c.CalculateNewVersion(test.analyzedCommits, test.lastVersion, test.releaseType, test.isFirst)
		assert.Equalf(t, test.isDraft, draft, "Should have draft %t for testcase %s", test.isDraft, test.testCase)
		assert.Equalf(t, test.nextVersion, next.String(), "Should have version %s for testcase %s", test.nextVersion, test.testCase)
	}

}
