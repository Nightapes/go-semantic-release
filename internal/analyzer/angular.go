// Package analyzer provides different commit analyzer
package analyzer

import (
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Nightapes/go-semantic-release/internal/shared"
)

type angular struct {
	rules []Rule
	regex string
	log   *log.Entry
	config  config.AnalyzerConfig
}

// ANGULAR identifier
const ANGULAR = "angular"
var angularFooterTokenSep = defaultTokenSeparators

func newAngular() *angular {
	return &angular{
		regex: `^(?P<type>\w*)(?:\((?P<scope>.*)\))?: (?P<subject>.*)`,
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

func (a *angular) analyze(commit shared.Commit, rule Rule) *shared.AnalyzedCommit {
	tokenSep := append(a.config.TokenSeparators, conventionalFooterTokenSep[:]...)

	firstSplit := strings.SplitN(commit.Message, "\n", 2)
	header := firstSplit[0]
	body := ""
	if len(firstSplit) > 1 {
		body = firstSplit[1]
	}
	matches := getRegexMatchedMap(a.regex, header)

	if len(matches) == 0 || matches["type"] != rule.Tag{
		a.log.Tracef("%s does not match %s, skip", commit.Message, rule.Tag)
		return nil
	}

	msgBlockMap := getDefaultMessageBlockMap(body, tokenSep)

	analyzed := &shared.AnalyzedCommit{
		Commit:    commit,
		Tag:       rule.Tag,
		TagString: rule.TagString,
		Scope:         shared.Scope(matches["scope"]),
		Subject:       strings.TrimSpace(matches["subject"]),
		MessageBlocks: msgBlockMap,
	}

	isBreaking := strings.Contains(commit.Message, defaultBreakingChangePrefix)
	analyzed.IsBreaking = isBreaking

	oldFormatMessage := strings.TrimSpace(matches["subject"] + "\n" + body)

	if !isBreaking {
		analyzed.ParsedMessage = strings.Trim(oldFormatMessage, " ")
		a.log.Tracef("%s: found %s", commit.Message, rule.Tag)
		return analyzed
	}

	a.log.Tracef(" %s, BREAKING CHANGE found", commit.Message)
	breakingChange := strings.SplitN(oldFormatMessage, defaultBreakingChangePrefix, 2)

	if len(breakingChange) > 1 {
		analyzed.ParsedMessage = strings.TrimSpace(breakingChange[0])
		analyzed.ParsedBreakingChangeMessage = strings.TrimSpace(breakingChange[1])
	} else {
		analyzed.ParsedBreakingChangeMessage = breakingChange[0]
	}
	return analyzed
}
