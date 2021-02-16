// Package analyzer provides different commit analyzer
package analyzer

import (
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
const breakingChangeKeywords = "BREAKING CHANGE"
const breakingChangePrefix =  breakingChangeKeywords + ":"

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
	prefixes := append(a.config.BlockPrefixes, breakingChangeKeywords)

	header, txtBlocks := getMessageParts(commit.Message)
	matches := getRegexMatchedMap(a.regex, header)

	if len(matches) == 0 || matches["type"] != rule.Tag{
		a.log.Tracef("%s does not match %s, skip", commit.Message, rule.Tag)
		return nil
	}

	msgBlockMap := make(map[string][]shared.MessageBlock)
	footer := ""
	if len(txtBlocks) > 0 {
		bodyCount := len(txtBlocks)-1
		if len(txtBlocks) == 1 {
			bodyCount = 1
		}
		bodyTxtBlocks := txtBlocks[0:bodyCount]
		if len(txtBlocks) > 1{
			footer = txtBlocks[len(txtBlocks)-1]
		}
		msgBlockMap["body"] = getMessageBlocks(bodyTxtBlocks, prefixes)

		if len(footer) > 0{
			footerLines := strings.Split(footer, "\n")
			msgBlockMap["footer"] = getMessageBlocks(footerLines, prefixes)
		}
	}

	analyzed := &shared.AnalyzedCommit{
		Commit:        commit,
		Tag:           rule.Tag,
		TagString:     rule.TagString,
		Scope:         shared.Scope(matches["scope"]),
		Subject:       strings.TrimSpace(matches["subject"]),
		MessageBlocks: msgBlockMap,
	}

	isBreaking := matches["breaking"] == "!" || strings.Contains(commit.Message, breakingChangePrefix)
	analyzed.IsBreaking = isBreaking

	oldMsgSplit := strings.SplitN(commit.Message, "\n", 2)
	originalBodyBlock := ""
	if len(oldMsgSplit) > 1 {
		originalBodyBlock = oldMsgSplit[1]
	}
	oldFormatMessage := strings.TrimSpace(matches["subject"] + "\n" + originalBodyBlock)
	if !isBreaking {
		analyzed.ParsedMessage = strings.Trim(oldFormatMessage, " ")
		a.log.Tracef("%s: found %s", commit.Message, rule.Tag)
		return analyzed
	}

	a.log.Infof(" %s, BREAKING CHANGE found", commit.Message)
	breakingChange := strings.SplitN(oldFormatMessage, breakingChangePrefix, 2)

	if len(breakingChange) > 1 {
		analyzed.ParsedMessage = strings.TrimSpace(breakingChange[0])
		analyzed.ParsedBreakingChangeMessage = strings.TrimSpace(breakingChange[1])
	} else {
		analyzed.ParsedBreakingChangeMessage = breakingChange[0]
	}

	return analyzed
}

func getMessageBlocks(txtArray,  prefixes []string) []shared.MessageBlock {
	blocks := make([]shared.MessageBlock, len(txtArray))
	for i, line := range txtArray{
		blocks[i] = parseMessageBlock(line, prefixes)
	}
	return blocks
}


