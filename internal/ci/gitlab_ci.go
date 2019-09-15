package ci

import (
	"fmt"
)

//GitlabCI struct
type GitlabCI struct{}

//Detect if on GitlabCI
func (t GitlabCI) detect(envs map[string]string) (*ProviderConfig, error) {

	if _, exists := envs["GITLAB_CI"]; !exists {
		return nil, fmt.Errorf("not running on gitlab")
	}

	return &ProviderConfig{
		Service:  "gitlab",
		Name:     "GitLab CI/CD",
		Commit:   envs["CI_COMMIT_SHA"],
		Tag:      envs["CI_COMMIT_TAG"],
		BuildURL: envs["CI_PROJECT_URL"] + "/pipelines/" + envs["CI_PIPELINE_ID"],
		Branch:   envs["CI_COMMIT_REF_NAME"],
		IsPR:     false,
	}, nil
}
