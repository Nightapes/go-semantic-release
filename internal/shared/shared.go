package shared

import (
	"github.com/Masterminds/semver"
)

//ReleaseVersion struct
type ReleaseVersion struct {
	Last   ReleaseVersionEntry
	Next   ReleaseVersionEntry
	Branch string
	Draft  bool
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

//ChangelogTemplateConfig struct
type ChangelogTemplateConfig struct {
	CommitURL  string
	CompareURL string
	Hash       string
	Version    string
}
