package analyzer

import (
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	log "github.com/sirupsen/logrus"
)

type Analyzer struct {
	CommitFormat string
}

type Rules struct {
	Tag     string
	Release string
}

type AnalyzeCommit interface {
	Analyze(commit gitutil.Commit, tag string) (AnalyzedCommit, bool)
	GetRules() []Rules
}

type AnalyzedCommit struct {
	Commit                      gitutil.Commit
	ParsedMessage               string
	Scope                       string
	ParsedBreakingChangeMessage string
}

func New(format string) *Analyzer {
	return &Analyzer{
		CommitFormat: format,
	}

}

func (a *Analyzer) Analyze(commits []gitutil.Commit) map[string][]AnalyzedCommit {

	var commitAnalayzer AnalyzeCommit
	switch a.CommitFormat {
	case "angular":
		log.Infof("analyze angular format")
		commitAnalayzer = NewAngular()
	}

	analyzedCommits := make(map[string][]AnalyzedCommit)
	analyzedCommits["major"] = make([]AnalyzedCommit, 0)
	analyzedCommits["minor"] = make([]AnalyzedCommit, 0)
	analyzedCommits["patch"] = make([]AnalyzedCommit, 0)

	for _, commit := range commits {
		for _, rule := range commitAnalayzer.GetRules() {
			analyzedCommit, hasBreakingChange := commitAnalayzer.Analyze(commit, rule.Tag)
			if hasBreakingChange {
				analyzedCommits["major"] = append(analyzedCommits["major"], analyzedCommit)
			} else {
				analyzedCommits[rule.Release] = append(analyzedCommits[rule.Release], analyzedCommit)
			}

		}
	}

	return analyzedCommits
}
