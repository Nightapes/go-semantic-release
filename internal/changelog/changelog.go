package changelog

import (
	"bytes"
	"text/template"
	"time"

	"github.com/Nightapes/go-semantic-release/internal/analyzer"
	"github.com/Nightapes/go-semantic-release/pkg/config"
)

const defaultChangelogTitle string = `v{{.Version}} ({{.Now.Format "2006-01-02"}})`
const defaultChangelog string = `{{ $version := .Version -}}
{{ $backtick := .Backtick -}}
# v{{.Version}} ({{.Now.Format "2006-01-02"}})
{{ range $key, $commits := .Commits }}
### {{ $key }}

{{range $index,$commit := $commits}}* **{{$backtick}}{{$commit.Scope}}:{{$backtick}}** {{$commit.ParsedMessage}}
{{ end -}}
{{ end -}}
`

type changelogContent struct {
	Commits  map[string][]analyzer.AnalyzedCommit
	Version  string
	Now      time.Time
	Backtick string
}

//CommitFormat struct
type Changelog struct {
	config *config.ReleaseConfig
}

//New Changelog struct for generating changelog from commits
func New(config *config.ReleaseConfig) *Changelog {
	return &Changelog{
		config: config,
	}
}

// GenerateChanglog from given commits
func (c *Changelog) GenerateChanglog(version string, analyzedCommits map[string][]analyzer.AnalyzedCommit) (string, string, error) {

	commitsPerScope := map[string][]analyzer.AnalyzedCommit{}
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
	}

	title, err := generateTemplate(defaultChangelogTitle, changelogContent)
	content, err := generateTemplate(defaultChangelog, changelogContent)

	return title, content, err
}

func generateTemplate(text string, values changelogContent) (string, error) {
	var tpl bytes.Buffer
	tmpl, err := template.New("template").Parse(text)
	if err != nil {
		return "", nil
	}
	err = tmpl.Execute(&tpl, values)
	if err != nil {
		return "", nil
	}
	return tpl.String(), nil
}
