package analyzer_test

import (
	"testing"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConventional(t *testing.T) {
	t.Parallel()
	testConfigs := []struct {
		testCase            string
		commits             []shared.Commit
		wantAnalyzedCommits map[shared.Release][]shared.AnalyzedCommit
	}{
		{
			testCase: "feat",
			wantAnalyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
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
						Subject: "my first commit",
						MessageBlocks: map[string][]shared.MessageBlock{},
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
						Subject: "no scope",
						MessageBlocks: map[string][]shared.MessageBlock{},
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
			wantAnalyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
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
						Subject: "my first commit",
						MessageBlocks: map[string][]shared.MessageBlock{},
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
						IsBreaking: true,
						Subject: "my first break",
						MessageBlocks: map[string][]shared.MessageBlock{},
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
			wantAnalyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
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
						Subject: "my first commit",
						MessageBlocks: map[string][]shared.MessageBlock{},
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
						IsBreaking: true,
						Subject: "my first break",
						MessageBlocks: map[string][]shared.MessageBlock{
							"body" : { shared.MessageBlock{
								Label:   "BREAKING CHANGE",
								Content: "change api to v2",
								},
							},
						},
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
						IsBreaking: true,
						Subject: "my first break",
						MessageBlocks: map[string][]shared.MessageBlock{
							"body" : {shared.MessageBlock{
								Label:   "BREAKING CHANGE",
								Content: "hey from the change",
								},
							},
						},
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
		},
		{
			testCase: "invalid",
			wantAnalyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
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
			wantAnalyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
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
						Subject:  "my first commit",
						MessageBlocks: map[string][]shared.MessageBlock{},
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
						Subject:  "my first build",
						MessageBlocks: map[string][]shared.MessageBlock{},
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
		{
			testCase: "fix and build",
			wantAnalyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"minor": {},
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
						Subject:  "my first build",
						MessageBlocks: map[string][]shared.MessageBlock{},
					},
				},
				"patch": {{
					Commit: shared.Commit{
						Message: "fix: my first commit",
						Author:  "me",
						Hash:    "12345667",
					},
					Scope:         "",
					ParsedMessage: "my first commit",
					Tag:           "fix",
					TagString:     "Bug fixes",
					Print:         true,
					Subject:  "my first commit",
					MessageBlocks: map[string][]shared.MessageBlock{},
				}},
				"major": {},
			},
			commits: []shared.Commit{
				{
					Message: "fix: my first commit",
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
		{
			testCase: "fix issue with footers",
			wantAnalyzedCommits: map[shared.Release][]shared.AnalyzedCommit{
				"patch": {{
					Commit: shared.Commit{
						Message: "fix: squash bug for logging\n\nNote: now the logs will not print lines twice.\n\nIssue: #123\nSeverity: medium",
						Author:  "me",
						Hash:    "12345667",
					},
					Scope:         "",
					ParsedMessage: "squash bug for logging\n\nNote: now the logs will not print lines twice.\n\nIssue: #123\nSeverity: medium",
					Tag:           "fix",
					TagString:     "Bug fixes",
					Print:         true,
					Subject:  "squash bug for logging",
					MessageBlocks: map[string][]shared.MessageBlock{
						"body": { shared.MessageBlock{
							Label:   "Note",
							Content: "now the logs will not print lines twice.",
						},
						},
						"footer": { shared.MessageBlock{
							Label:   "Issue",
							Content: "#123",
						},
						shared.MessageBlock{
							Label:   "Severity",
							Content: "medium",
						},
						},
					},
				}},
				"minor": {},
				"major": {},
				"none": {},
			},
			commits: []shared.Commit{
				{
					Message: "fix: squash bug for logging\n\nNote: now the logs will not print lines twice.\n\nIssue: #123\nSeverity: medium",
					Author:  "me",
					Hash:    "12345667",
				},
			},
		},
	}

	conventional, err := analyzer.New("conventional", config.AnalyzerConfig{BlockPrefixes: []string{"Note", "Issue", "Severity"}}, config.ChangelogConfig{})
	assert.NoError(t, err)

	for _, test := range testConfigs {
		analyzedCommits := conventional.Analyze(test.commits)
		assert.Equalf(t, test.wantAnalyzedCommits["major"], analyzedCommits["major"], "Testcase %s should have major commits", test.testCase)
		assert.Equalf(t, test.wantAnalyzedCommits["minor"], analyzedCommits["minor"], "Testcase %s should have minor commits", test.testCase)
		assert.Equalf(t, test.wantAnalyzedCommits["patch"], analyzedCommits["patch"], "Testcase %s should have patch commits", test.testCase)
		assert.Equalf(t, test.wantAnalyzedCommits["none"], analyzedCommits["none"], "Testcase %s should have none commits", test.testCase)
	}
}
