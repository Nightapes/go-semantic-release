package analyzer

import (
	"regexp"
	"strings"

	"github.com/Nightapes/go-semantic-release/internal/gitutil"
)

type Angular struct {
	rules []Rules
	regex string
}

func NewAngular() *Angular {
	return &Angular{
		regex: `(TAG)(?:\((.*)\))?: (.*)`,
		rules: []Rules{
			{
				Tag:     "feat",
				Release: "minor",
			},
			{
				Tag:     "fix",
				Release: "patch",
			}, {
				Tag:     "perf",
				Release: "patch",
			},
		},
	}
}

func (a *Angular) GetRules() []Rules {
	return a.rules
}

func (a *Angular) Analyze(commit gitutil.Commit, tag string) (AnalyzedCommit, bool) {

	analyzed := AnalyzedCommit{
		Commit: commit,
	}

	re := regexp.MustCompile(strings.Replace(a.regex, "TAG", tag, -1))
	matches := re.FindAllStringSubmatch(commit.Message+" "+commit.Message, -1)
	if len(matches) >= 1 {
		if len(matches[0]) >= 3 {
			analyzed.Scope = matches[0][2]

			message := strings.Join(matches[0][3:], "")
			splitted := strings.SplitN(message, "BREAKING CHANGE:", 1)

			if len(splitted) == 1 {
				analyzed.ParsedMessage = splitted[0]
				return analyzed, false
			}
			analyzed.ParsedMessage = splitted[0]
			analyzed.ParsedBreakingChangeMessage = splitted[1]
			return analyzed, true

		}
	}
	return analyzed, false

}
