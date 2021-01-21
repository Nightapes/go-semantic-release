package analyzer_test

import (
	"testing"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConventional(t *testing.T) {

	testConfigs := []struct {
		testCase        string
		commits         []shared.Commit
		analyzedCommits map[shared.Release][]shared.AnalyzedCommit
	}{
		{
			testCase: "feat",
			analyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": {
					{
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
					{
						Commit: shared.Commit{
							Message: "feat: no scope",
							Author:  "me",
							Hash:    "12345667",
						},
						Scope:         "",
						ParsedMessage: "no scope",
						Tag:           "feat",
						TagString:     "Features",
						Print:         true,
					},
				},
				"major": {},
				"patch": {},
				"none":  {},
			},
			commits: []shared.Commit{
				{
					Message: "feat(internal/changelog): my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
				{
					Message: "feat: no scope",
					Author:  "me",
					Hash:    "12345667",
				},
			},
		},
		{
			testCase: "feat breaking change",
			analyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": {
					{
						Commit: shared.Commit{
							Message: "feat: my first commit",
							Author:  "me",
							Hash:    "12345667",
						},
						Scope:         "",
						ParsedMessage: "my first commit",
						Tag:           "feat",
						TagString:     "Features",
						Print:         true,
					},
				},
				"major": {
					{
						Commit: shared.Commit{
							Message: "feat!: my first break",
							Author:  "me",
							Hash:    "12345668",
						},
						Scope:                       "",
						ParsedMessage:               "",
						Tag:                         "feat",
						TagString:                   "Features",
						Print:                       true,
						ParsedBreakingChangeMessage: "my first break",
					},
				},
				"patch": {},
				"none":  {},
			},
			commits: []shared.Commit{
				{
					Message: "feat: my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
				{
					Message: "feat!: my first break",
					Author:  "me",
					Hash:    "12345668",
				},
			},
		},
		{
			testCase: "feat breaking change footer",
			commits: []shared.Commit{
				{
					Message: "feat: my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
				{
					Message: "feat: my first break \n\nBREAKING CHANGE: change api to v2\n",
					Author:  "me",
					Hash:    "12345668",
				},
				{
					Message: "feat!: my first break \n\nBREAKING CHANGE: hey from the change",
					Author:  "me",
					Hash:    "12345669",
				},
			},
			analyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": {
					{
						Commit: shared.Commit{
							Message: "feat: my first commit",
							Author:  "me",
							Hash:    "12345667",
						},
						Scope:         "",
						ParsedMessage: "my first commit",
						Tag:           "feat",
						TagString:     "Features",
						Print:         true,
					},
				},
				"major": {
					{
						Commit: shared.Commit{
							Message: "feat: my first break \n\nBREAKING CHANGE: change api to v2\n",
							Author:  "me",
							Hash:    "12345668",
						},
						Scope:                       "",
						ParsedMessage:               "my first break",
						Tag:                         "feat",
						TagString:                   "Features",
						Print:                       true,
						ParsedBreakingChangeMessage: "change api to v2",
					},
					{
						Commit: shared.Commit{
							Message: "feat!: my first break \n\nBREAKING CHANGE: hey from the change",
							Author:  "me",
							Hash:    "12345669",
						},
						Scope:                       "",
						ParsedMessage:               "my first break",
						Tag:                         "feat",
						TagString:                   "Features",
						Print:                       true,
						ParsedBreakingChangeMessage: "hey from the change",
					},
				},
				"patch": {},
				"none":  {},
			},
		},
		{
			testCase: "invalid",
			analyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": {},
				"major": {},
				"patch": {},
				"none":  {},
			},
			commits: []shared.Commit{
				{
					Message: "internal/changelog: my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
				{
					Message: "Merge feat: my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
			},
		},
		{
			testCase: "feat and build",
			analyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": {
					{
						Commit: shared.Commit{
							Message: "feat: my first commit",
							Author:  "me",
							Hash:    "12345667",
						},
						Scope:         "",
						ParsedMessage: "my first commit",
						Tag:           "feat",
						TagString:     "Features",
						Print:         true,
					},
				},
				"none": {
					{
						Commit: shared.Commit{
							Message: "build: my first build",
							Author:  "me",
							Hash:    "12345668",
						},
						Scope:                       "",
						ParsedMessage:               "my first build",
						Tag:                         "build",
						TagString:                   "Changes to CI/CD",
						Print:                       false,
						ParsedBreakingChangeMessage: "",
					},
				},
				"patch": {},
				"major": {},
			},
			commits: []shared.Commit{
				{
					Message: "feat: my first commit",
					Author:  "me",
					Hash:    "12345667",
				},
				{
					Message: "build: my first build",
					Author:  "me",
					Hash:    "12345668",
				},
			},
		},
	}

	conventional, err := analyzer.New("conventional", config.ChangelogConfig{})
	assert.NoError(t, err)

	for _, test := range testConfigs {
		analyzedCommits := conventional.Analyze(test.commits)
		assert.Equalf(t, test.analyzedCommits["major"], analyzedCommits["major"], "Testcase %s should have major commits", test.testCase)
		assert.Equalf(t, test.analyzedCommits["minor"], analyzedCommits["minor"], "Testcase %s should have minor commits", test.testCase)
		assert.Equalf(t, test.analyzedCommits["patch"], analyzedCommits["patch"], "Testcase %s should have patch commits", test.testCase)
		assert.Equalf(t, test.analyzedCommits["none"], analyzedCommits["none"], "Testcase %s should have none commits", test.testCase)
	}

}
