// Package analyzer provides different commit analyzer
package analyzer

import (
	"fmt"

	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

// Analyzer struct
type Analyzer struct {
	analyzeCommits analyzeCommits
	Config         config.ChangelogConfig
}

// Rule for commits
type Rule struct {
	Tag       string
	TagString string
	Release   shared.Release
	Changelog bool
}

type analyzeCommits interface {
	analyze(commit shared.Commit, tag Rule) (*shared.AnalyzedCommit, bool)
	getRules() []Rule
}

// New Analyzer struct for given commit format
func New(format string, config config.ChangelogConfig) (*Analyzer, error) {
	analyzer := &Analyzer{
		Config: config,
	}

	switch format {
	case ANGULAR:
		analyzer.analyzeCommits = newAngular()
		log.Debugf("Commit format set to %s", ANGULAR)
	case CONVENTIONAL:
		analyzer.analyzeCommits = newConventional()
		log.Debugf("Commit format set to %s", CONVENTIONAL)
	default:
		return nil, fmt.Errorf("invalid commit format: %s", format)
	}
	return analyzer, nil
}

// GetRules from current mode
func (a *Analyzer) GetRules() []Rule {
	return a.analyzeCommits.getRules()
}

// Analyze commits and return commits split by major,minor,patch
func (a *Analyzer) Analyze(commits []shared.Commit) map[shared.Release][]shared.AnalyzedCommit {
	analyzedCommits := make(map[shared.Release][]shared.AnalyzedCommit)
	analyzedCommits["major"] = make([]shared.AnalyzedCommit, 0)
	analyzedCommits["minor"] = make([]shared.AnalyzedCommit, 0)
	analyzedCommits["patch"] = make([]shared.AnalyzedCommit, 0)
	analyzedCommits["none"] = make([]shared.AnalyzedCommit, 0)

	for _, commit := range commits {
		for _, rule := range a.analyzeCommits.getRules() {
			analyzedCommit, hasBreakingChange := a.analyzeCommits.analyze(commit, rule)
			if analyzedCommit == nil {
				continue
			}
			if a.Config.PrintAll || rule.Changelog {
				analyzedCommit.Print = true
			}
			if hasBreakingChange {
				analyzedCommits["major"] = append(analyzedCommits["major"], *analyzedCommit)
				break
			}
			analyzedCommits[rule.Release] = append(analyzedCommits[rule.Release], *analyzedCommit)
			break
		}
	}
	log.Debugf("Analyzed commits: major=%d minor=%d patch=%d none=%d", len(analyzedCommits["major"]), len(analyzedCommits["minor"]), len(analyzedCommits["patch"]), len(analyzedCommits["none"]))
	return analyzedCommits
}
