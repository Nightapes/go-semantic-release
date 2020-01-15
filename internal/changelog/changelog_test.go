package changelog_test

import (
	"testing"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/changelog"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestChangelog(t *testing.T) {

	templateConfig := shared.ChangelogTemplateConfig{
		CommitURL:  "https://commit.url",
		CompareURL: "https://compare.url",
		Hash:       "hash",
		Version:    "1.0.0",
	}

	testConfigs := []struct {
		testCase        string
		analyzedCommits map[shared.Release][]shared.AnalyzedCommit
		result          *shared.GeneratedChangelog
		hasError        bool
	}{
		{
			testCase: "feat",
			analyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": []shared.AnalyzedCommit{
					shared.AnalyzedCommit{
						Commit: shared.Commit{
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
			},
			result: &shared.GeneratedChangelog{
				Title:   "v1.0.0 (2019-07-19)",
				Content: "# v1.0.0 (2019-07-19)\n\n\n### Features\n\n* **`internal/changelog`** my first commit ([1234566](https://commit.url))\n\n",
			},
			hasError: false,
		},
		{
			testCase: "feat no scope",
			analyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": []shared.AnalyzedCommit{
					shared.AnalyzedCommit{
						Commit: shared.Commit{
							Message: "feat: my first commit",
							Author:  "me",
							Hash:    "12345667",
						},
						ParsedMessage: "my first commit",
						Tag:           "feat",
						TagString:     "Features",
						Print:         true,
					},
				},
			},
			result: &shared.GeneratedChangelog{
				Title:   "v1.0.0 (2019-07-19)",
				Content: "# v1.0.0 (2019-07-19)\n\n\n### Features\n\n* my first commit ([1234566](https://commit.url))\n\n",
			},
			hasError: false,
		},
		{
			testCase: "feat breaking change",
			analyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": []shared.AnalyzedCommit{
					shared.AnalyzedCommit{
						Commit: shared.Commit{
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
					shared.AnalyzedCommit{
						Commit: shared.Commit{
							Message: "feat(internal/changelog): my first break: BREAKING CHANGE: change api to v2",
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
			},
			result: &shared.GeneratedChangelog{
				Title:   "v1.0.0 (2019-07-19)",
				Content: "# v1.0.0 (2019-07-19)\n\n## BREAKING CHANGES\n\n* **`internal/changelog`** change api to v2  \nintroduced by commit: \nmy first break  ([1234566](https://commit.url))\n\n\n### Features\n\n* **`internal/changelog`** my first commit ([1234566](https://commit.url))\n\n",
			},
			hasError: false,
		},
	}

	cl := changelog.New(&config.ReleaseConfig{}, []analyzer.Rule{
		{
			Tag:       "feat",
			TagString: "Features",
			Release:   "minor",
			Changelog: true,
		},
		{
			Tag:       "fix",
			TagString: "Bug fixes",
			Release:   "patch",
			Changelog: true,
		},
		{
			Tag:       "build",
			TagString: "Build",
			Release:   "none",
			Changelog: false,
		},
	}, time.Date(2019, 7, 19, 0, 0, 0, 0, time.UTC))

	for _, config := range testConfigs {
		generatedChangelog, err := cl.GenerateChanglog(templateConfig, config.analyzedCommits)
		assert.Equalf(t, config.hasError, err != nil, "Testcase %s should have error: %t -> %s", config.testCase, config.hasError, err)
		assert.Equalf(t, config.result, generatedChangelog, "Testcase %s should have generated changelog", config.testCase)
	}

}
