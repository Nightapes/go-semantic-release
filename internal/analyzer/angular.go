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
	rules []Rule
	regex string
}

func newAngular() *angular {
	return &angular{
		regex: `(TAG)(?:\((.*)\))?: (.*)`,
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
			}, {
				Tag:       "perf",
				TagString: "Performance improvments",
				Release:   "patch",
				Changelog: true,
			}, {
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
			}, {
				Tag:       "refactor",
				TagString: "Code refactor",
				Release:   "none",
				Changelog: false,
			}, {
				Tag:       "test",
				TagString: "Testing",
				Release:   "none",
				Changelog: false,
			}, {
				Tag:       "chore",
				TagString: "Changes to the build process or auxiliary tools and libraries such as documentation generation",
				Release:   "none",
				Changelog: false,
			}, {
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

func (a *angular) analyze(commit gitutil.Commit, rule Rule) (AnalyzedCommit, bool, bool, error) {

	analyzed := AnalyzedCommit{
		Commit:    commit,
		Tag:       rule.Tag,
		TagString: rule.TagString,
	}

	re := regexp.MustCompile(strings.Replace(a.regex, "TAG", rule.Tag, -1))
	matches := re.FindAllStringSubmatch(commit.Message, -1)
	if len(matches) >= 1 {
		if len(matches[0]) >= 3 {

			analyzed.Scope = matches[0][2]

			message := strings.Join(matches[0][3:], "")
			breakingChange := strings.SplitN(message, "BREAKING CHANGE:", 1)
			draft := strings.SplitN(message, "DRAFT:", 1)

			if len(breakingChange) == 1 && len(draft) == 1 {
				analyzed.ParsedMessage = breakingChange[0]
				log.Tracef("%s: found %s", commit.Message, rule.Tag)
				return analyzed, false, false, nil

			}

			if len(breakingChange) > 1 {

				analyzed.ParsedMessage = breakingChange[0]
				analyzed.ParsedBreakingChangeMessage = breakingChange[1]

				log.Tracef(" %s, BREAKING CHANGE found", commit.Message)

				return analyzed, true, false, nil

			} else if len(draft) > 1 {
				analyzed.ParsedMessage = draft[0]
				analyzed.ParsedDraftMessage = draft[1]
				log.Tracef(" %s, DRAFT found", commit.Message)

			}

		}
	}
	log.Tracef("%s does not match %s, skip", commit.Message, rule.Tag)
	return analyzed, false, false, fmt.Errorf("not found")

}
