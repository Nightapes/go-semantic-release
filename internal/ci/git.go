package ci

import (
	"fmt"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
)

//Git struct
type Git struct {
	gitUtil *gitutil.GitUtil
}

//Detect if on Git
func (t Git) detect(envs map[string]string) (*ProviderConfig, error) {

	if _, exists := envs["CI"]; !exists {
		return nil, fmt.Errorf("running not git only")
	}

	hash, err := t.gitUtil.GetHash()
	if err != nil {
		return nil, err
	}

	currentBranch, err := t.gitUtil.GetBranch()
	if err != nil {
		return nil, err
	}

	return &ProviderConfig{
		Service: "Git",
		Name:    "Git only",
		Commit:  hash,
		Branch:  currentBranch,
		IsPR:    false,
	}, nil
}
