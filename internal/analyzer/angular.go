// Package analyzer provides different commit analyzer
package analyzer

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Nightapes/go-semantic-release/internal/gitutil"
)

type angular struct {
	rules []Rules
	regex string
}

func newAngular() *angular {
	return &angular{
		regex: `(TAG)(?:\((.*)\))?: (.*)`,
		rules: []Rules{
			{
				Tag:       "feat",
				Release:   "minor",
				Changelog: true,
			},
			{
				Tag:       "fix",
				Release:   "patch",
				Changelog: true,
			}, {
				Tag:       "perf",
				Release:   "patch",
				Changelog: true,
			}, {
				Tag:       "docs",
				Release:   "none",
				Changelog: false,
			},
			{
				Tag:       "style",
				Release:   "none",
				Changelog: false,
			}, {
				Tag:       "refactor",
				Release:   "none",
				Changelog: false,
			}, {
				Tag:       "test",
				Release:   "none",
				Changelog: false,
			}, {
				Tag:       "chore",
				Release:   "none",
				Changelog: false,
			}, {
				Tag:       "build",
				Release:   "none",
				Changelog: false,
			},
		},
	}
}

func (a *angular) getRules() []Rules {
	return a.rules
}

func (a *angular) analyze(commit gitutil.Commit, tag string) (AnalyzedCommit, bool, error) {

	analyzed := AnalyzedCommit{
		Commit: commit,
		Tag:    tag,
	}

	re := regexp.MustCompile(strings.Replace(a.regex, "TAG", tag, -1))
	matches := re.FindAllStringSubmatch(commit.Message, -1)
	if len(matches) >= 1 {
		if len(matches[0]) >= 3 {
			analyzed.Scope = matches[0][2]

			message := strings.Join(matches[0][3:], "")
			splitted := strings.SplitN(message, "BREAKING CHANGE:", 1)

			if len(splitted) == 1 {
				analyzed.ParsedMessage = splitted[0]
				log.Tracef("%s: found %s", commit.Message, tag)
				return analyzed, false, nil
			}
			analyzed.ParsedMessage = splitted[0]
			analyzed.ParsedBreakingChangeMessage = splitted[1]
			log.Tracef(" %s, BREAKING CHANGE found", commit.Message)
			return analyzed, true, nil

		}
	}
	log.Tracef("%s does not match %s, skip", commit.Message, tag)
	return analyzed, false, fmt.Errorf("Not found")

}
