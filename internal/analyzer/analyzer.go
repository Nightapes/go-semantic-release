// Package analyzer provides different commit analyzer
package analyzer

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
)

const breakingChangeKeywords = "BREAKING CHANGE"
const defaultBreakingChangePrefix = breakingChangeKeywords + ":"
const footerTokenRegex = "^(?P<token>[^\\s*-][\\w\\- ]+[^\\s])<SEP>.*"

var defaultTokenSeparators = [2]string{": ", " #"}

// Analyzer struct
type Analyzer struct {
	analyzeCommits  analyzeCommits
	ChangelogConfig config.ChangelogConfig
	AnalyzerConfig  config.AnalyzerConfig
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
		AnalyzerConfig:  analyzerConfig,
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
func getMessageBlocksFromTexts(txtArray, separators []string) []shared.MessageBlock {
	blocks := make([]shared.MessageBlock, len(txtArray))
	for i, line := range txtArray {
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
	if token, sep := findFooterToken(msg, separators); len(token) > 0 {
		msgBlock.Label = token
		content := strings.Replace(msg, token+sep, "", 1)
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

//
// getDefaultMessageBlockMap parses a text block and splits in to different sections.
// default logic to distinguish different parts is:
//  - Body starts right after the header (without beginning with a token)
//  - Body ends when a footer is discovered or text ends
//  - A footer is detected when it starts with a token ending with a separator
//  - A footer ends when another footer is found or text ends
//
func getDefaultMessageBlockMap(txtBlock string, tokenSep []string) map[string][]shared.MessageBlock {
	msgBlockMap := make(map[string][]shared.MessageBlock)
	footers := make([]string, 0)
	body, footerBlock, line := "", "", ""
	footerFound := false
	// Look through each line
	scanner := bufio.NewScanner(strings.NewReader(txtBlock))
	for scanner.Scan() {
		line = scanner.Text()
		if token, _ := findFooterToken(line, tokenSep); len(token) > 0 {
			// if footer was already found from before
			if len(footerBlock) > 0 {
				footers = append(footers, strings.TrimSpace(footerBlock))
			}
			footerFound = true
			footerBlock = ""
		}

		//'\n' is removed when reading from scanner
		if !footerFound {
			body += line + "\n"
		} else {
			footerBlock += line + "\n"
		}
	}
	if len(footerBlock) > 0 {
		footers = append(footers, strings.TrimSpace(footerBlock))
	}

	body = strings.TrimSpace(body)
	if len(body) > 0 {
		msgBlockMap["body"] = []shared.MessageBlock{{
			Label:   "",
			Content: body,
		}}
	}

	footerBlocks := getMessageBlocksFromTexts(footers, tokenSep)
	if len(footerBlocks) > 0 {
		msgBlockMap["footer"] = footerBlocks
	}

	return msgBlockMap
}
