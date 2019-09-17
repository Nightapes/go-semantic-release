package gitlab

// Release struct
type Release struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Ref         string `json:"ref"`
	Description string `json:"description,omitempty"`
}

// ReleaseLink struct
type ReleaseLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ProjectFile struct
type ProjectFile struct {
	Alt      string `json:"alt"`
	URL      string `json:"url"`
	Markdown string `json:"markdown"`
}
