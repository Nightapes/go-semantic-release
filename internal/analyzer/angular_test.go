package analyzer_test

import (
	"testing"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestAngular(t *testing.T) {

	testConfigs := []struct {
		testCase        string
		commits         []gitutil.Commit
		analyzedCommits map[analyzer.Release][]analyzer.AnalyzedCommit
	}{
		{
			testCase: "feat",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"minor": []analyzer.AnalyzedCommit{
					analyzer.AnalyzedCommit{
						Commit: gitutil.Commit{
							Message: "feat(internal/changelog): my first commit",
							Author:  "me",
							Hash:    "12345667",
						},
						Scope:         "internal/changelog",
						ParsedMessage: "my first commit",
						Tag:           "feat",
						TagString:     "Features",
						Print:         true,
					},
				},
				"major": []analyzer.AnalyzedCommit{},
				"patch": []analyzer.AnalyzedCommit{},
				"none":  []analyzer.AnalyzedCommit{},
			},
			commits: []gitutil.Commit{
				gitutil.Commit{
					Message: "feat(internal/changelog): my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
			},
		},
		{
			testCase: "feat breaking change",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"minor": []analyzer.AnalyzedCommit{
					analyzer.AnalyzedCommit{
						Commit: gitutil.Commit{
							Message: "feat(internal/changelog): my first commit",
							Author:  "me",
							Hash:    "12345667",
						},
						Scope:         "internal/changelog",
						ParsedMessage: "my first commit",
						Tag:           "feat",
						TagString:     "Features",
						Print:         true,
					},
				},
				"major": []analyzer.AnalyzedCommit{
					analyzer.AnalyzedCommit{
						Commit: gitutil.Commit{
							Message: "feat(internal/changelog): my first break BREAKING CHANGE: change api to v2",
							Author:  "me",
							Hash:    "12345668",
						},
						Scope:                       "internal/changelog",
						ParsedMessage:               "my first break",
						Tag:                         "feat",
						TagString:                   "Features",
						Print:                       true,
						ParsedBreakingChangeMessage: "change api to v2",
					},
				},
				"patch": []analyzer.AnalyzedCommit{},
				"none":  []analyzer.AnalyzedCommit{},
			},
			commits: []gitutil.Commit{
				gitutil.Commit{
					Message: "feat(internal/changelog): my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
				gitutil.Commit{
					Message: "feat(internal/changelog): my first break BREAKING CHANGE: change api to v2",
					Author:  "me",
					Hash:    "12345668",
				},
			},
		},
		{
			testCase: "invalid",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"minor": []analyzer.AnalyzedCommit{},
				"major": []analyzer.AnalyzedCommit{},
				"patch": []analyzer.AnalyzedCommit{},
				"none":  []analyzer.AnalyzedCommit{},
			},
			commits: []gitutil.Commit{
				gitutil.Commit{
					Message: "internal/changelog: my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
			},
		},
		{
			testCase: "feat and build",
			analyzedCommits: map[analyzer.Release][]analyzer.AnalyzedCommit{
				"minor": []analyzer.AnalyzedCommit{
					analyzer.AnalyzedCommit{
						Commit: gitutil.Commit{
							Message: "feat(internal/changelog): my first commit",
							Author:  "me",
							Hash:    "12345667",
						},
						Scope:         "internal/changelog",
						ParsedMessage: "my first commit",
						Tag:           "feat",
						TagString:     "Features",
						Print:         true,
					},
				},
				"none": []analyzer.AnalyzedCommit{
					analyzer.AnalyzedCommit{
						Commit: gitutil.Commit{
							Message: "build(internal/changelog): my first build",
							Author:  "me",
							Hash:    "12345668",
						},
						Scope:                       "internal/changelog",
						ParsedMessage:               "my first build",
						Tag:                         "build",
						TagString:                   "Changes to CI/CD",
						Print:                       false,
						ParsedBreakingChangeMessage: "",
					},
				},
				"patch": []analyzer.AnalyzedCommit{},
				"major": []analyzer.AnalyzedCommit{},
			},
			commits: []gitutil.Commit{
				gitutil.Commit{
					Message: "feat(internal/changelog): my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
				gitutil.Commit{
					Message: "build(internal/changelog): my first build",
					Author:  "me",
					Hash:    "12345668",
				},
			},
		},
	}

	angular, err := analyzer.New("angular", config.ChangelogConfig{})
	assert.NoError(t, err)

	for _, test := range testConfigs {
		analyzedCommits := angular.Analyze(test.commits)
		assert.Equalf(t, test.analyzedCommits["major"], analyzedCommits["major"], "Testcase %s should have major commits", test.testCase)
		assert.Equalf(t, test.analyzedCommits["minor"], analyzedCommits["minor"], "Testcase %s should have minor commits", test.testCase)
		assert.Equalf(t, test.analyzedCommits["patch"], analyzedCommits["patch"], "Testcase %s should have patch commits", test.testCase)
		assert.Equalf(t, test.analyzedCommits["none"], analyzedCommits["none"], "Testcase %s should have none commits", test.testCase)
	}

}
