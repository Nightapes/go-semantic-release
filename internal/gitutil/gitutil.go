// Package gitutil provides helper methods for git
package gitutil

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Commit struct
type Commit struct {
	Message string
	Author  string
	Hash    string
}

// GitUtil struct
type GitUtil struct {
	Repository *git.Repository
}

// New GitUtil struct and open git repository
func New(folder string) (*GitUtil, error) {
	r, err := git.PlainOpen(folder)
	if err != nil {
		return nil, err
	}
	utils := &GitUtil{
		Repository: r,
	}
	return utils, nil

}

// GetHash from git HEAD
func (g *GitUtil) GetHash() (string, error) {
	ref, err := g.Repository.Head()
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}

// GetBranch from git HEAD
func (g *GitUtil) GetBranch() (string, error) {
	ref, err := g.Repository.Head()
	if err != nil {
		return "", err
	}

	if !ref.Name().IsBranch() {
		branches, err := g.Repository.Branches()
		if err != nil {
			return "", err
		}

		var currentBranch string
		found := branches.ForEach(func(p *plumbing.Reference) error {

			if p.Name().IsBranch() && p.Name().Short() != "origin" {
				currentBranch = p.Name().Short()
				return fmt.Errorf("break")
			}
			return nil
		})

		if found != nil {
			log.Debugf("Found branch from HEAD %s", currentBranch)
			return currentBranch, nil
		}

		return "", fmt.Errorf("no branch found, found %s, please checkout a branch (git checkout -b <BRANCH>)", ref.Name().String())
	}
	log.Debugf("Found branch %s", ref.Name().Short())
	return ref.Name().Short(), nil
}

// GetLastVersion from git tags
func (g *GitUtil) GetLastVersion() (*semver.Version, string, error) {

	var tags []*semver.Version

	gitTags, err := g.Repository.Tags()

	if err != nil {
		return nil, "", err
	}

	err = gitTags.ForEach(func(p *plumbing.Reference) error {
		v, err := semver.NewVersion(p.Name().Short())
		log.Tracef("%+v with hash: %s", p.Target(), p.Hash())

		if err == nil {
			tags = append(tags, v)
		} else {
			log.Debugf("Tag %s is not a valid version, skip", p.Name().Short())
		}
		return nil
	})

	if err != nil {
		return nil, "", err
	}

	sort.Sort(sort.Reverse(semver.Collection(tags)))

	if len(tags) == 0 {
		log.Debugf("Found no tags")
		return nil, "", nil
	}

	log.Debugf("Found old version %s", tags[0].String())

	tag, err := g.Repository.Tag(tags[0].Original())
	if err != nil {
		return nil, "", err
	}

	log.Debugf("Found old hash %s", tag.Hash().String())
	return tags[0], tag.Hash().String(), nil
}

// GetCommits from git hash to HEAD
func (g *GitUtil) GetCommits(lastTagHash string) ([]Commit, error) {

	ref, err := g.Repository.Head()
	if err != nil {
		return nil, err
	}

	cIter, err := g.Repository.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}

	var commits []Commit
	var foundEnd bool

	err = cIter.ForEach(func(c *object.Commit) error {
		if c.Hash.String() == lastTagHash {
			log.Debugf("Found commit with hash %s, will stop here", c.Hash.String())
			foundEnd = true

		}
		if !foundEnd {
			log.Tracef("Found commit with hash %s", c.Hash.String())
			commit := Commit{
				Message: c.Message,
				Author:  c.Committer.Name,
				Hash:    c.Hash.String(),
			}
			commits = append(commits, commit)
		}
		return nil
	})

	return commits, err
}
