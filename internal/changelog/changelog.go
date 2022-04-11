package changelog

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"

	log "github.com/sirupsen/logrus"
)

const defaultCommitList string = `{{ range $index,$commit := .BreakingChanges -}}
{{ if eq $index 0 -}}
## BREAKING CHANGES
{{ end -}}
* {{ if $commit.Scope }}**{{$.Backtick}}{{$commit.Scope}}{{$.Backtick}}**{{ end }} {{$commit.ParsedBreakingChangeMessage}}  
introduced by commit: 
{{$commit.Subject}} {{if $.HasURL}} ([{{ printf "%.7s" $commit.Commit.Hash}}]({{ replace $.URL "{{hash}}" $commit.Commit.Hash}})){{end}}
{{ end -}}
{{ range $key := .Order  -}}
{{ $commits := index $.Commits $key -}}
{{ if $commits -}}
### {{ $key }}
{{ range $index,$commit := $commits -}}
* {{ if $commit.Scope }}**{{$.Backtick}}{{$commit.Scope}}{{$.Backtick}}** {{end}}{{$commit.Subject}}{{if $.HasURL}} ([{{ printf "%.7s" $commit.Commit.Hash}}]({{ replace $.URL "{{hash}}" $commit.Commit.Hash}})){{end}}
{{ if not $.ShowBodyAsHeader -}}
{{ if $commit.MessageBlocks.body -}}
{{ range $indexBlock,$bodyBlock := $commit.MessageBlocks.body -}}
{{ addPrefixToLines  $bodyBlock.Content "  > "}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}`
const defaultCommitListSubTemplate = `{{ define "commitList" }}` + defaultCommitList + "{{ end }}"
const defaultChangelogTitle = `v{{.Version}} ({{.Now.Format "2006-01-02"}})`
const defaultChangelog = `# v{{$.Version}} ({{.Now.Format "2006-01-02"}})
{{ if .ShowBodyAsHeader -}}

{{ range $key := .CommitsContent.Order  -}}
{{ $commits := index $.CommitsContent.Commits $key -}}
{{ if $commits -}}
{{ range $index,$commit := $commits -}}
{{ if $commit.MessageBlocks.body -}}
{{ range $indexBlock,$bodyBlock := $commit.MessageBlocks.body -}}
{{ $bodyBlock.Content }}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}
{{ end -}}

{{ template "commitList" .CommitsContent -}}

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

{{ if .HasNPM}}
## NodeJS Package

New NodeJS package is released under [{{.NPMPackageName}}]({{.NPMRepository}})

### Usage

{{$.Backtick}}yarn add {{.NPMPackageName}}@{{.Version}}{{$.Backtick}}

or

{{$.Backtick}}npm install -save {{.NPMPackageName}}@{{.Version}}{{$.Backtick}}

{{ end -}}

{{ if .ShowAuthors -}}
# Special Thanks

{{range $i,$a := .Authors}}{{if gt $i 0 }}, {{end}}{{$a}}{{end}}
{{ end -}}
`

type changelogContent struct {
	Commits          string
	CommitsContent   commitsContent
	Version          string
	Now              time.Time
	Backtick         string
	ShowBodyAsHeader bool
	HasDocker        bool
	HasDockerLatest  bool
	DockerRepository string
	HasNPM           bool
	IsYarn           bool
	NPMRepository    string
	NPMPackageName   string
	Authors          []string
	ShowAuthors      bool
}

type commitsContent struct {
	Commits          map[string][]shared.AnalyzedCommit
	BreakingChanges  []shared.AnalyzedCommit
	Order            []string
	ShowBodyAsHeader bool
	Backtick         string
	HasURL           bool
	URL              string
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

// GenerateChangelog from given commits
func (c *Changelog) GenerateChangelog(templateConfig shared.ChangelogTemplateConfig, analyzedCommits map[shared.Release][]shared.AnalyzedCommit) (*shared.GeneratedChangelog, error) {

	commitsPerScope := map[string][]shared.AnalyzedCommit{}
	var commitsBreakingChange []shared.AnalyzedCommit
	order := make([]string, 0)

	for _, rule := range c.rules {
		c.log.Tracef("Add %s to list", rule.TagString)
		if rule.Changelog || c.config.Changelog.PrintAll {
			order = append(order, rule.TagString)
		}
	}

	authors := map[string]bool{}

	for _, commits := range analyzedCommits {
		for _, commit := range commits {
			authors[commit.Commit.Author] = true
			if commit.Print {
				if commit.IsBreaking {
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
		Commits:          commitsPerScope,
		BreakingChanges:  commitsBreakingChange,
		Backtick:         "`",
		Order:            order,
		ShowBodyAsHeader: c.config.Changelog.ShowBodyAsHeader,
		HasURL:           templateConfig.CommitURL != "",
		URL:              templateConfig.CommitURL,
	}

	authorsNames := make([]string, len(authors))
	i := 0
	for k := range authors {
		authorsNames[i] = k
		i++
	}

	sort.Sort(sort.StringSlice(authorsNames))

	changelogContent := changelogContent{
		CommitsContent:   commitsContent,
		Version:          templateConfig.Version,
		Now:              c.releaseTime,
		Backtick:         "`",
		HasDocker:        c.config.Changelog.Docker.Repository != "",
		HasDockerLatest:  c.config.Changelog.Docker.Latest,
		DockerRepository: c.config.Changelog.Docker.Repository,
		HasNPM:           c.config.Changelog.NPM.PackageName != "",
		NPMPackageName:   c.config.Changelog.NPM.PackageName,
		NPMRepository:    c.config.Changelog.NPM.Repository,
		ShowBodyAsHeader: c.config.Changelog.ShowBodyAsHeader,
		ShowAuthors:      c.config.Changelog.ShowAuthors && len(authors) > 0,
		Authors:          authorsNames,
	}

	chglogTemplate := defaultCommitListSubTemplate + defaultChangelog
	if c.config.Changelog.TemplatePath != "" {
		content, err := ioutil.ReadFile(c.config.Changelog.TemplatePath)
		if err != nil {
			return nil, err
		}
		chglogTemplate = string(content)
	}

	templateTitle := defaultChangelogTitle
	if c.config.Changelog.TemplateTitle != "" {
		templateTitle = c.config.Changelog.TemplateTitle
	}

	log.Debugf("Render title")
	renderedTitle, err := generateTemplate(templateTitle, changelogContent, nil)
	if err != nil {
		return nil, err
	}

	log.Debugf("Render commits")
	renderedCommitList, err := generateTemplate(defaultCommitList, commitsContent, nil)
	if err != nil {
		return nil, err
	}

	log.Tracef("Commits %s", renderedCommitList)
	changelogContent.Commits = renderedCommitList

	extraFuncMap := template.FuncMap{
		"commitUrl": func() string { return templateConfig.CommitURL },
	}
	log.Debugf("Render changelog")
	renderedContent, err := generateTemplate(chglogTemplate, changelogContent, extraFuncMap)

	return &shared.GeneratedChangelog{Title: renderedTitle, Content: renderedContent}, err
}

func generateTemplate(text string, values interface{}, extraFuncMap template.FuncMap) (string, error) {

	funcMap := template.FuncMap{
		"replace":          replace,
		"lower":            lower,
		"upper":            upper,
		"capitalize":       capitalize,
		"addPrefixToLines": addPrefixToLines,
	}

	for k, v := range extraFuncMap {
		funcMap[k] = v
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

func lower(input string) string {
	return strings.ToLower(input)
}

func upper(input string) string {
	return strings.ToUpper(input)
}

func capitalize(input string) string {
	if len(input) > 0 {
		return strings.ToUpper(string(input[0])) + input[1:]
	}
	return ""
}

// Adds a prefix to each line of the given text block
// this can be helpful in rendering correct indentation or bullets for multi-line texts
func addPrefixToLines(input, prefix string) string {
	output := ""
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		output += prefix + scanner.Text() + "\n"
	}
	output = strings.TrimRight(output, "\n")
	return output
}
