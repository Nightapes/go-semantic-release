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

func (a *angular) analyze(commit gitutil.Commit, rule Rule) (AnalyzedCommit, bool, error) {

	analyzed := AnalyzedCommit{
		Commit:    commit,
		Tag:       rule.Tag,
		TagString: rule.TagString,
	}

	re := regexp.MustCompile(strings.Replace(a.regex, "TAG", rule.Tag, -1))
	matches := re.FindAllStringSubmatch(commit.Message, -1)
	if len(matches) >= 1 {
		if len(matches[0]) >= 3 {

			analyzed.Scope = Scope(matches[0][2])

			message := strings.Join(matches[0][3:], "")
			if !strings.Contains(message, "BREAKING CHANGE:") {
				analyzed.ParsedMessage = strings.Trim(message, " ")

				log.Tracef("%s: found %s", commit.Message, rule.Tag)
				return analyzed, false, nil
			}
			breakingChange := strings.SplitN(message, "BREAKING CHANGE:", 2)

			analyzed.ParsedMessage = strings.Trim(breakingChange[0], " ")
			analyzed.ParsedBreakingChangeMessage = strings.Trim(breakingChange[1], " ")

			log.Tracef(" %s, BREAKING CHANGE found", commit.Message)
			return analyzed, true, nil
		}
	}
	log.Tracef("%s does not match %s, skip", commit.Message, rule.Tag)
	return analyzed, false, fmt.Errorf("not found")

}
