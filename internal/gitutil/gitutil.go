// Package gitutil provides helper methods for git
package gitutil

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
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
		return "", fmt.Errorf("no branch found, found %s, please checkout a branch (git checkout <BRANCH>)", ref.Name().String())
	}

	return ref.Name().Short(), nil
}

// GetLastVersion from git tags
func (g *GitUtil) GetLastVersion() (*semver.Version, string, error) {

	log.Debugf("GetLastVersion")

	tagObjects, err := g.Repository.TagObjects()
	if err != nil {
		return nil, "", err
	}

	var tags []*semver.Version

	err = tagObjects.ForEach(func(t *object.Tag) error {
		v, err := semver.NewVersion(t.Name)

		if err != nil {
			log.Debugf("Tag %s is not a valid version, skip", t.Name)
		} else {
			log.Debugf("Add tag %s", t.Name)
			tags = append(tags, v)
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

	tagObject, err := g.Repository.TagObject(tag.Hash())
	if err != nil {
		return nil, "", err
	}

	log.Debugf("Found old hash %s", tagObject.Target.String())
	return tags[0], tagObject.Target.String(), nil
}

// GetCommits from git hash to HEAD
func (g *GitUtil) GetCommits(lastTagHash string) ([]Commit, error) {

	log.Printf("Read head")
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
			log.Infof("%s == %s", c.Hash.String(), lastTagHash)
			foundEnd = true
		}

		if !foundEnd {
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
