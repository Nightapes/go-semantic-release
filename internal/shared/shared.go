package shared

import (
	"github.com/Masterminds/semver"
)

//ReleaseVersion struct
type ReleaseVersion struct {
	Last   ReleaseVersionEntry
	Next   ReleaseVersionEntry
	Branch string
}

//ReleaseVersionEntry struct
type ReleaseVersionEntry struct {
	Commit  string
	Version *semver.Version
}

//GeneratedChangelog struct
type GeneratedChangelog struct {
	Title   string
	Content string
}

//GenerateChangelogConfig struct
type ChangelogTemplateConfig struct {
	CommitURL  string
	CompareURL string
	Hash       string
	Version    string
}
