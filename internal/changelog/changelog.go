package changelog

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"

	log "github.com/sirupsen/logrus"
)

const defaultCommitList string = `{{ range $index,$commit := .BreakingChanges -}}
{{ if eq $index 0 }}
## BREAKING CHANGES
{{ end}}
* **{{$.Backtick}}{{$commit.Scope}}{{$.Backtick}}** {{$commit.ParsedBreakingChangeMessage}}  
introduced by commit: 
{{$commit.ParsedMessage}} {{if $.HasURL}} ([{{ printf "%.7s" $commit.Commit.Hash}}]({{ replace $.URL "{{hash}}" $commit.Commit.Hash}}))  {{end}}
{{ end -}}
{{ range $key := .Order }}
{{ $commits := index $.Commits $key}} {{if $commits -}}
### {{ $key }}
{{ range $index,$commit := $commits -}}
* **{{$.Backtick}}{{$commit.Scope}}{{$.Backtick}}** {{$commit.ParsedMessage}} {{if $.HasURL}} ([{{ printf "%.7s" $commit.Commit.Hash}}]({{ replace $.URL "{{hash}}" $commit.Commit.Hash}}))  {{end}}
{{ end -}}
{{ end -}}
{{ end -}}`
const defaultChangelogTitle string = `v{{.Version}} ({{.Now.Format "2006-01-02"}})`
const defaultChangelog string = `# v{{$.Version}} ({{.Now.Format "2006-01-02"}})
{{ .Commits -}}
{{ if .HasDocker}}
## Docker image

New docker image is released under {{$.Backtick}}{{.DockerRepository}}:{{.Version}}{{$.Backtick}}

### Usage

{{$.Backtick}}docker run {{.DockerRepository}}:{{.Version}}{{$.Backtick}}
{{ if .HasDockerLatest}}
or

{{$.Backtick}}docker run {{.DockerRepository}}:latest{{$.Backtick}}
{{ end -}}
{{ end -}}
`

type changelogContent struct {
	Commits          string
	Version          string
	Now              time.Time
	Backtick         string
	HasDocker        bool
	HasDockerLatest  bool
	DockerRepository string
}

type commitsContent struct {
	Commits         map[string][]shared.AnalyzedCommit
	BreakingChanges []shared.AnalyzedCommit
	Order           []string
	Version         string
	Now             time.Time
	Backtick        string
	HasURL          bool
	URL             string
}

//Changelog struct
type Changelog struct {
	config      *config.ReleaseConfig
	rules       []analyzer.Rule
	releaseTime time.Time
	log         *log.Entry
}

//New Changelog struct for generating changelog from commits
func New(config *config.ReleaseConfig, rules []analyzer.Rule, releaseTime time.Time) *Changelog {
	return &Changelog{
		config:      config,
		rules:       rules,
		releaseTime: releaseTime,
		log:         log.WithField("changelog", config.CommitFormat),
	}
}

// GenerateChanglog from given commits
func (c *Changelog) GenerateChanglog(templateConfig shared.ChangelogTemplateConfig, analyzedCommits map[shared.Release][]shared.AnalyzedCommit) (*shared.GeneratedChangelog, error) {

	commitsPerScope := map[string][]shared.AnalyzedCommit{}
	commitsBreakingChange := []shared.AnalyzedCommit{}
	order := make([]string, 0)

	for _, rule := range c.rules {
		c.log.Tracef("Add %s to list", rule.TagString)
		if rule.Changelog || c.config.Changelog.PrintAll {
			order = append(order, rule.TagString)
		}
	}

	for _, commits := range analyzedCommits {
		for _, commit := range commits {
			if commit.Print {
				if commit.ParsedBreakingChangeMessage != "" {
					commitsBreakingChange = append(commitsBreakingChange, commit)
					continue
				}
				if _, ok := commitsPerScope[commit.TagString]; !ok {
					commitsPerScope[commit.TagString] = make([]shared.AnalyzedCommit, 0)
				}
				commitsPerScope[commit.TagString] = append(commitsPerScope[commit.TagString], commit)
			}
		}
	}

	commitsContent := commitsContent{
		Version:         templateConfig.Version,
		Commits:         commitsPerScope,
		Now:             c.releaseTime,
		BreakingChanges: commitsBreakingChange,
		Backtick:        "`",
		Order:           order,
		HasURL:          templateConfig.CommitURL != "",
		URL:             templateConfig.CommitURL,
	}

	changelogContent := changelogContent{
		Version:          templateConfig.Version,
		Now:              c.releaseTime,
		Backtick:         "`",
		HasDocker:        c.config.Changelog.Docker.Repository != "",
		HasDockerLatest:  c.config.Changelog.Docker.Latest,
		DockerRepository: c.config.Changelog.Docker.Repository,
	}
	template := defaultChangelog
	if c.config.Changelog.TemplatePath != "" {
		content, err := ioutil.ReadFile(c.config.Changelog.TemplatePath)
		if err != nil {
			return nil, err
		}
		template = string(content)
	}

	templateTitle := defaultChangelogTitle
	if c.config.Changelog.TemplateTitle != "" {
		templateTitle = c.config.Changelog.TemplateTitle
	}

	log.Debugf("Render title")
	renderedTitle, err := generateTemplate(templateTitle, changelogContent)
	if err != nil {
		return nil, err
	}

	log.Debugf("Render commits")
	renderedCommitList, err := generateTemplate(defaultCommitList, commitsContent)
	if err != nil {
		return nil, err
	}

	log.Tracef("Commits %s", renderedCommitList)

	changelogContent.Commits = renderedCommitList
	log.Debugf("Render changelog")
	renderedContent, err := generateTemplate(template, changelogContent)

	return &shared.GeneratedChangelog{Title: renderedTitle, Content: renderedContent}, err
}

func generateTemplate(text string, values interface{}) (string, error) {

	funcMap := template.FuncMap{
		"replace": replace,
	}

	var tpl bytes.Buffer
	tmpl, err := template.New("template").Funcs(funcMap).Parse(text)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&tpl, values)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}
