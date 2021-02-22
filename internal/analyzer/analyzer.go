// Package analyzer provides different commit analyzer
package analyzer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

const breakingChangeKeywords = "BREAKING CHANGE"
const defaultBreakingChangePrefix =  breakingChangeKeywords + ":"
const footerTokenRegex = "^(?P<token>[^\\s][\\w\\- ]+[^\\s])<SEP>.*"
var defaultTokenSeparators = [2]string{ ": ", " #"}

// Analyzer struct
type Analyzer struct {
	analyzeCommits  analyzeCommits
	ChangelogConfig config.ChangelogConfig
	AnalyzerConfig config.AnalyzerConfig
}

// Rule for commits
type Rule struct {
	Tag       string
	TagString string
	Release   shared.Release
	Changelog bool
}

type analyzeCommits interface {
	analyze(commit shared.Commit, tag Rule) *shared.AnalyzedCommit
	getRules() []Rule
}

// New Analyzer struct for given commit format
func New(format string, analyzerConfig config.AnalyzerConfig, chglogConfig config.ChangelogConfig) (*Analyzer, error) {
	analyzer := &Analyzer{
		AnalyzerConfig: analyzerConfig,
		ChangelogConfig: chglogConfig,
	}

	switch format {
	case ANGULAR:
		analyzer.analyzeCommits = newAngular()
		log.Debugf("Commit format set to %s", ANGULAR)
	case CONVENTIONAL:
		analyzer.analyzeCommits = newConventional(analyzerConfig)
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
			analyzedCommit := a.analyzeCommits.analyze(commit, rule)
			if analyzedCommit == nil {
				continue
			}
			if a.ChangelogConfig.PrintAll || rule.Changelog {
				analyzedCommit.Print = true
			}
			if analyzedCommit.IsBreaking {
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

func getMessageParts(msg string) (header string, bodyBlocks []string){
	firstSplit := strings.SplitN(msg, "\n", 2)
	header = firstSplit[0]
	bodyBlocks = make([]string, 0)

	if len(firstSplit) < 2 {
		return
	}
	// Trim and then split by a blank line
	remaining := strings.Trim(firstSplit[1], "\n")
	bodyBlocks = strings.Split(remaining, "\n\n")

	return
}

//
// getRegexMatchedMap will match a regex with named groups and map the matching
//  results to corresponding group names
//
func getRegexMatchedMap(regEx, url string) (paramsMap map[string]string) {
	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

//
// getMessageBlocksFromTexts converts strings to an array of MessageBlock
//
func getMessageBlocksFromTexts(txtArray,  separators []string) []shared.MessageBlock {
	blocks := make([]shared.MessageBlock, len(txtArray))
	for i, line := range txtArray{
		blocks[i] = parseMessageBlock(line, separators)
	}
	return blocks
}

//
// parseMessageBlock parses a text in to MessageBlock
//
func parseMessageBlock(msg string, separators []string) shared.MessageBlock {
	msgBlock := shared.MessageBlock{
		Label:   "",
		Content: msg,
	}
	if token, sep := findFooterToken(msg, separators); len(token)  > 0{
		msgBlock.Label = token
		content := strings.Replace(msg, token + sep, "", 1)
		msgBlock.Content = strings.TrimSpace(content)
	}
	return msgBlock
}

//
// findFooterToken checks if given text has a token with one of the separators and returns a token
//
func findFooterToken(text string, separators []string) (token string, sep string) {
	for _, sep := range separators {
		regex := strings.Replace(footerTokenRegex, "<SEP>", sep, 1)
		matches := getRegexMatchedMap(regex, text)
		if token, ok := matches["token"]; ok {
			return token, sep
		}
	}
	return "", ""
}
