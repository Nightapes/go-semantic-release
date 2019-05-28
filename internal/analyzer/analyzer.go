// Package analyzer provides different commit analyzer
package analyzer

import (
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

//Analyzer struct
type Analyzer struct {
	CommitFormat string
	Config       config.ChangelogConfig
}

//Rule for commits
type Rule struct {
	Tag       string
	TagString string
	Release   string
	Changelog bool
}

type analyzeCommit interface {
	analyze(commit gitutil.Commit, tag Rule) (AnalyzedCommit, bool, error)
	getRules() []Rule
}

//AnalyzedCommit struct
type AnalyzedCommit struct {
	Commit                      gitutil.Commit
	ParsedMessage               string
	Scope                       string
	ParsedBreakingChangeMessage string
	Tag                         string
	TagString                   string
	Print                       bool
}

//New Analyzer struct for given commit format
func New(format string, config config.ChangelogConfig) *Analyzer {
	return &Analyzer{
		CommitFormat: format,
		Config:       config,
	}

}

// Analyze commits and return commits splitted by major,minor,patch
func (a *Analyzer) Analyze(commits []gitutil.Commit) map[string][]AnalyzedCommit {

	var commitAnalayzer analyzeCommit
	switch a.CommitFormat {
	case "angular":
		log.Debugf("Commit format set to angular")
		commitAnalayzer = newAngular()
	}

	analyzedCommits := make(map[string][]AnalyzedCommit)
	analyzedCommits["major"] = make([]AnalyzedCommit, 0)
	analyzedCommits["minor"] = make([]AnalyzedCommit, 0)
	analyzedCommits["patch"] = make([]AnalyzedCommit, 0)
	analyzedCommits["none"] = make([]AnalyzedCommit, 0)

	for _, commit := range commits {
		for _, rule := range commitAnalayzer.getRules() {
			analyzedCommit, hasBreakingChange, err := commitAnalayzer.analyze(commit, rule)
			if err == nil {
				if a.Config.PrintAll {
					analyzedCommit.Print = true
				} else {
					analyzedCommit.Print = rule.Changelog
				}
				if hasBreakingChange {
					analyzedCommits["major"] = append(analyzedCommits["major"], analyzedCommit)
				} else {
					analyzedCommits[rule.Release] = append(analyzedCommits[rule.Release], analyzedCommit)
				}
				break
			}

		}
	}

	log.Debugf("Analyzed commits: major=%d minor=%d patch=%d none=%d", len(analyzedCommits["major"]), len(analyzedCommits["minor"]), len(analyzedCommits["patch"]), len(analyzedCommits["none"]))

	return analyzedCommits
}
