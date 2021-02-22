// Package analyzer provides different commit analyzer
package analyzer

import (
	"bufio"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Nightapes/go-semantic-release/internal/shared"
)

type conventional struct {
	rules []Rule
	regex string
	log   *log.Entry
	config  config.AnalyzerConfig
}

// CONVENTIONAL identifier
const CONVENTIONAL = "conventional"
var conventionalFooterTokenSep = defaultTokenSeparators

func newConventional(config config.AnalyzerConfig) *conventional {
	return &conventional{
		config: config,
		regex: `^(?P<type>\w*)(?:\((?P<scope>.*)\))?(?P<breaking>\!)?: (?P<subject>.*)`,
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
			},
			{
				Tag:       "perf",
				TagString: "Performance improvements",
				Release:   "patch",
				Changelog: true,
			},
			{
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
			},
			{
				Tag:       "refactor",
				TagString: "Code refactor",
				Release:   "none",
				Changelog: false,
			},
			{
				Tag:       "test",
				TagString: "Testing",
				Release:   "none",
				Changelog: false,
			},
			{
				Tag:       "chore",
				TagString: "Changes to the build process or auxiliary tools and libraries such as documentation generation",
				Release:   "none",
				Changelog: false,
			},
			{
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

func (a *conventional) analyze(commit shared.Commit, rule Rule) *shared.AnalyzedCommit {
	tokenSep := append(a.config.TokenSeparators, conventionalFooterTokenSep[:]...)

	firstSplit := strings.SplitN(commit.Message, "\n", 2)
	header := firstSplit[0]
	body := ""
	if len(firstSplit) > 1 {
		body = firstSplit[1]
	}

	matches := getRegexMatchedMap(a.regex, header)

	if len(matches) == 0 || matches["type"] != rule.Tag{
		a.log.Tracef("%s does not match %s, skip", commit.Message, rule.Tag)
		return nil
	}

	msgBlockMap := getConventionalMessageBlockMap(body, tokenSep)

	analyzed := &shared.AnalyzedCommit{
		Commit:        commit,
		Tag:           rule.Tag,
		TagString:     rule.TagString,
		Scope:         shared.Scope(matches["scope"]),
		Subject:       strings.TrimSpace(matches["subject"]),
		MessageBlocks: msgBlockMap,
	}

	isBreaking := matches["breaking"] == "!" || strings.Contains(commit.Message, defaultBreakingChangePrefix)
	analyzed.IsBreaking = isBreaking

	oldFormatMessage := strings.TrimSpace(matches["subject"] + "\n" + body)
	if !isBreaking {
		analyzed.ParsedMessage = strings.Trim(oldFormatMessage, " ")
		a.log.Tracef("%s: found %s", commit.Message, rule.Tag)
		return analyzed
	}

	a.log.Infof(" %s, BREAKING CHANGE found", commit.Message)
	breakingChange := strings.SplitN(oldFormatMessage, defaultBreakingChangePrefix, 2)

	if len(breakingChange) > 1 {
		analyzed.ParsedMessage = strings.TrimSpace(breakingChange[0])
		analyzed.ParsedBreakingChangeMessage = strings.TrimSpace(breakingChange[1])
	} else {
		analyzed.ParsedBreakingChangeMessage = breakingChange[0]
	}

	return analyzed
}

func getConventionalMessageBlockMap(txtBlock string, tokenSep []string) map[string][]shared.MessageBlock{
	msgBlockMap := make(map[string][]shared.MessageBlock)
	footers := make([]string, 0)
	body := ""
	footerBlock := ""
	line := ""
	footerFound := false
	// Look through each line
	scanner := bufio.NewScanner(strings.NewReader(txtBlock))
	for scanner.Scan() {
		line = scanner.Text()
		if token, _ := findFooterToken(line, tokenSep); len(token) > 0 {
			// if footer was already found from before
			if len(footerBlock) > 0{
				footers = append(footers, strings.TrimSpace(footerBlock))
			}
			footerFound = true
			footerBlock = ""
		}

		//'\n' is removed when reading from scanner
		if !footerFound {
			body += line + "\n"
		}else{
			footerBlock += line + "\n"
		}
	}
	if len(footerBlock) > 0 {
		footers = append(footers, strings.TrimSpace(footerBlock))
	}

	body = strings.TrimSpace(body)
	if len(body) > 0{
		msgBlockMap["body"] = []shared.MessageBlock {{
			Label:   "",
			Content: body,
		} }
	}

	footerBlocks := getMessageBlocksFromTexts(footers, tokenSep)
	if len(footerBlocks) > 0 {
		msgBlockMap["footer"] = footerBlocks
	}


	return msgBlockMap
}
