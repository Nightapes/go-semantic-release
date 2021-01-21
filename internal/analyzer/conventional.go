// Package analyzer provides different commit analyzer
package analyzer

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Nightapes/go-semantic-release/internal/shared"
)

type conventional struct {
	rules []Rule
	regex string
	log   *log.Entry
}

// CONVENTIONAL identifier
const CONVENTIONAL = "conventional"

func newConventional() *conventional {
	return &conventional{
		regex: `^(TAG)(?:\((.*)\))?(\!)?: (?s)(.*)`,
		log:   log.WithField("analyzer", CONVENTIONAL),
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

func (a *conventional) getRules() []Rule {
	return a.rules
}

func (a *conventional) analyze(commit shared.Commit, rule Rule) (shared.AnalyzedCommit, bool, error) {

	analyzed := shared.AnalyzedCommit{
		Commit:    commit,
		Tag:       rule.Tag,
		TagString: rule.TagString,
	}

	re := regexp.MustCompile(strings.Replace(a.regex, "TAG", rule.Tag, -1))
	matches := re.FindAllStringSubmatch(commit.Message, -1)
	if len(matches) >= 1 {
		if len(matches[0]) >= 4 {

			analyzed.Scope = shared.Scope(matches[0][2])

			message := strings.Join(matches[0][4:], "")
			if matches[0][3] == "" && !strings.Contains(message, "BREAKING CHANGE:") {
				analyzed.ParsedMessage = strings.Trim(message, " ")

				a.log.Tracef("%s: found %s", commit.Message, rule.Tag)
				return analyzed, false, nil
			}

			if matches[0][3] == "" {
				breakingChange := strings.SplitN(message, "BREAKING CHANGE:", 2)
				analyzed.ParsedMessage = strings.TrimSpace(breakingChange[0])
				analyzed.ParsedBreakingChangeMessage = strings.TrimSpace(breakingChange[1])
				a.log.Tracef(" %s, BREAKING CHANGE found", commit.Message)
				return analyzed, true, nil
			}

			breakingChange := strings.SplitN(message, "BREAKING CHANGE:", 2)

			if len(breakingChange) > 1 {
				analyzed.ParsedMessage = strings.TrimSpace(breakingChange[0])
				analyzed.ParsedBreakingChangeMessage = strings.TrimSpace(breakingChange[1])
			} else {
				analyzed.ParsedBreakingChangeMessage = breakingChange[0]
			}
			a.log.Infof(" %s, BREAKING CHANGE found", commit.Message)
			return analyzed, true, nil
		}
	}
	a.log.Tracef("%s does not match %s, skip", commit.Message, rule.Tag)
	return analyzed, false, fmt.Errorf("not found")

}
