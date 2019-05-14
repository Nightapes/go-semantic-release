// Package analyzer provides different commit analyzer
package analyzer

import (
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	log "github.com/sirupsen/logrus"
)

//Analyzer struct
type Analyzer struct {
	CommitFormat string
}

//Rules for commits
type Rules struct {
	Tag     string
	Release string
}

type analyzeCommit interface {
	analyze(commit gitutil.Commit, tag string) (AnalyzedCommit, bool)
	getRules() []Rules
}

//AnalyzedCommit struct
type AnalyzedCommit struct {
	Commit                      gitutil.Commit
	ParsedMessage               string
	Scope                       string
	ParsedBreakingChangeMessage string
}

//New Analyzer struct for given commit format
func New(format string) *Analyzer {
	return &Analyzer{
		CommitFormat: format,
	}

}

// Analyze commits and return commits splitted by major,minor,patch
func (a *Analyzer) Analyze(commits []gitutil.Commit) map[string][]AnalyzedCommit {

	var commitAnalayzer analyzeCommit
	switch a.CommitFormat {
	case "angular":
		log.Infof("analyze angular format")
		commitAnalayzer = newAngular()
	}

	analyzedCommits := make(map[string][]AnalyzedCommit)
	analyzedCommits["major"] = make([]AnalyzedCommit, 0)
	analyzedCommits["minor"] = make([]AnalyzedCommit, 0)
	analyzedCommits["patch"] = make([]AnalyzedCommit, 0)

	for _, commit := range commits {
		for _, rule := range commitAnalayzer.getRules() {
			analyzedCommit, hasBreakingChange := commitAnalayzer.analyze(commit, rule.Tag)
			if hasBreakingChange {
				analyzedCommits["major"] = append(analyzedCommits["major"], analyzedCommit)
			} else {
				analyzedCommits[rule.Release] = append(analyzedCommits[rule.Release], analyzedCommit)
			}

		}
	}

	return analyzedCommits
}
