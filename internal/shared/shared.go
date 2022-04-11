package shared

import (
	"github.com/Masterminds/semver"
)

//ReleaseVersion struct
type ReleaseVersion struct {
	Last    ReleaseVersionEntry          `yaml:"last"`
	Next    ReleaseVersionEntry          `yaml:"next"`
	Branch  string                       `yaml:"branch"`
	Commits map[Release][]AnalyzedCommit `yaml:"commits"`
}

//ReleaseVersionEntry struct
type ReleaseVersionEntry struct {
	Commit        string          `yaml:"commit"`
	VersionString string          `yaml:"version"`
	Version       *semver.Version `yaml:"-"`
}

//GeneratedChangelog struct
type GeneratedChangelog struct {
	Title   string
	Content string
}

//ChangelogTemplateConfig struct
type ChangelogTemplateConfig struct {
	CommitURL  string
	CompareURL string
	Hash       string
	Version    string
}

//AnalyzedCommit struct
type AnalyzedCommit struct {
	Commit                      Commit                    `yaml:"commit"`
	ParsedMessage               string                    `yaml:"parsedMessage"`
	ParsedBreakingChangeMessage string                    `yaml:"parsedBreakingChangeMessage"`
	Tag                         string                    `yaml:"tag"`
	TagString                   string                    `yaml:"tagString"`
	Scope                       Scope                     `yaml:"scope"`
	Subject                     string                    `yaml:"subject"`
	MessageBlocks               map[string][]MessageBlock `yaml:"messageBlocks"`
	IsBreaking                  bool                      `yaml:"isBreaking"`
	Print                       bool                      `yaml:"print"`
}

// MessageBlock represents a block in the body section of a commit message
type MessageBlock struct {
	Label   string `yaml:"label"`
	Content string `yaml:"content"`
}

//Scope of the commit, like feat, fix,..
type Scope string

//Release types, like major
type Release string

// Commit struct
type Commit struct {
	Message string `yaml:"message"`
	Author  string `yaml:"author"`
	Hash    string `yaml:"hash"`
}
