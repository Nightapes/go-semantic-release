package changelog

import (
	"bytes"
	"strings"
	"text/template"
	"time"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/pkg/config"

	log "github.com/sirupsen/logrus"
)

const defaultChangelogTitle string = `v{{.Version}} ({{.Now.Format "2006-01-02"}})`
const defaultChangelog string = `# v{{$.Version}} ({{.Now.Format "2006-01-02"}})
{{ range $key := .Order }}
{{ $commits := index $.Commits $key}} {{if $commits -}}
### {{ $key }}
{{ range $index,$commit := $commits -}}
* **{{$.Backtick}}{{$commit.Scope}}{{$.Backtick}}** {{$commit.ParsedMessage}} {{if $.HasURL}} ([{{ printf "%.7s" $commit.Commit.Hash}}]({{ replace $.URL "{{hash}}" $commit.Commit.Hash}}))  {{end}}
{{ end -}}
{{ end -}}
{{ end -}}
`

type changelogContent struct {
	Commits  map[string][]analyzer.AnalyzedCommit
	Order    []string
	Version  string
	Now      time.Time
	Backtick string
	HasURL   bool
	URL      string
}

//Changelog struct
type Changelog struct {
	config *config.ReleaseConfig
	rules  []analyzer.Rule
}

//New Changelog struct for generating changelog from commits
func New(config *config.ReleaseConfig, rules []analyzer.Rule) *Changelog {
	return &Changelog{
		config: config,
		rules:  rules,
	}
}

// GenerateChanglog from given commits
func (c *Changelog) GenerateChanglog(version, url string, analyzedCommits map[string][]analyzer.AnalyzedCommit) (string, string, error) {

	commitsPerScope := map[string][]analyzer.AnalyzedCommit{}
	order := make([]string, 0)

	for _, rule := range c.rules {
		log.Debugf("Add %s to list", rule.TagString)
		if rule.Changelog || c.config.Changelog.PrintAll {
			order = append(order, rule.TagString)
		}
	}

	for _, commits := range analyzedCommits {
		for _, commit := range commits {
			if commit.Print {
				if _, ok := commitsPerScope[commit.TagString]; !ok {
					commitsPerScope[commit.TagString] = make([]analyzer.AnalyzedCommit, 0)
				}
				commitsPerScope[commit.TagString] = append(commitsPerScope[commit.TagString], commit)
			}
		}
	}

	changelogContent := changelogContent{
		Version:  version,
		Commits:  commitsPerScope,
		Now:      time.Now(),
		Backtick: "`",
		Order:    order,
		HasURL:   url != "",
		URL:      url,
	}

	title, err := generateTemplate(defaultChangelogTitle, changelogContent)
	if err != nil {
		return "", "", err
	}
	content, err := generateTemplate(defaultChangelog, changelogContent)

	return title, content, err
}

func generateTemplate(text string, values changelogContent) (string, error) {

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
