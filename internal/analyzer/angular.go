// Package analyzer provides different commit analyzer
package analyzer

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Nightapes/go-semantic-release/internal/shared"
)

type angular struct {
	rules []Rule
	regex string
	log   *log.Entry
}

// ANGULAR identifier
const ANGULAR = "angular"

func newAngular() *angular {
	return &angular{
		regex: `^(TAG)(?:\((.*)\))?: (?s)(.*)`,
		log:   log.WithField("analyzer", ANGULAR),
		rules: []Rule{
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
				Tag:       "perf",
				TagString: "Performance improvements",
				Release:   "patch",
				Changelog: true,
			},
			{
				Tag:       "docs",
				TagString: "Documentation changes",
				Release:   "none",
				Changelog: false,
			},
			{
				Tag:       "style",
				TagString: "Style",
				Release:   "none",
				Changelog: false,
			},
			{
				Tag:       "refactor",
				TagString: "Code refactor",
				Release:   "none",
				Changelog: false,
			},
			{
				Tag:       "test",
				TagString: "Testing",
				Release:   "none",
				Changelog: false,
			},
			{
				Tag:       "chore",
				TagString: "Changes to the build process or auxiliary tools and libraries such as documentation generation",
				Release:   "none",
				Changelog: false,
			},
			{
				Tag:       "build",
				TagString: "Changes to CI/CD",
				Release:   "none",
				Changelog: false,
			},
		},
	}
}

func (a *angular) getRules() []Rule {
	return a.rules
}

func (a *angular) analyze(commit shared.Commit, rule Rule) (*shared.AnalyzedCommit, bool) {
	re := regexp.MustCompile(strings.Replace(a.regex, "TAG", rule.Tag, -1))
	matches := re.FindStringSubmatch(commit.Message)
	if matches == nil {
		a.log.Tracef("%s does not match %s, skip", commit.Message, rule.Tag)
		return nil, false
	}

	analyzed := &shared.AnalyzedCommit{
		Commit:    commit,
		Tag:       rule.Tag,
		TagString: rule.TagString,
		Scope:     shared.Scope(matches[2]),
	}

	message := strings.Join(matches[3:], "")
	if !strings.Contains(message, "BREAKING CHANGE:") {
		analyzed.ParsedMessage = strings.Trim(message, " ")
		a.log.Tracef("%s: found %s", commit.Message, rule.Tag)
		return analyzed, false
	}

	a.log.Tracef(" %s, BREAKING CHANGE found", commit.Message)
	breakingChange := strings.SplitN(message, "BREAKING CHANGE:", 2)

	if len(breakingChange) > 1 {
		analyzed.ParsedMessage = strings.TrimSpace(breakingChange[0])
		analyzed.ParsedBreakingChangeMessage = strings.TrimSpace(breakingChange[1])
		return analyzed, true
	}

	analyzed.ParsedBreakingChangeMessage = breakingChange[0]
	return analyzed, true
}
