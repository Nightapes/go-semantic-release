package cache_test

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/cache"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
)

func TestReadCacheNotFound(t *testing.T) {

	resp, err := cache.Read("notfound/dir")
	assert.NoErrorf(t, err, "Read non exsiting file")
	assert.NotNil(t, resp)

}

func TestReadCacheInvalidContent(t *testing.T) {

	dir, err := ioutil.TempDir("", "prefix")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	completePath := path.Join(path.Dir(dir), ".version")
	brokenContent := []byte("hello broken\ngo: lang\n")
	err = ioutil.WriteFile(completePath, brokenContent, 0644)
	assert.NoError(t, err)

	_, readError := cache.Read(dir)
	assert.Errorf(t, readError, "Should give error, when broken content")

}

func TestWriteAndReadCache(t *testing.T) {

	dir, err := ioutil.TempDir("", "prefix")

	assert.NoError(t, err)

	content := shared.ReleaseVersion{
		Last: shared.ReleaseVersionEntry{
			Commit:        "12345",
			Version:       createVersion("1.0.0"),
			VersionString: "1.0.0",
		},
		Next: shared.ReleaseVersionEntry{
			Commit:        "12346",
			Version:       createVersion("1.1.0"),
			VersionString: "1.1.0",
		},
		Branch: "master",
		Draft:  true,
		Commits: map[shared.Release][]shared.AnalyzedCommit{
			"major": []shared.AnalyzedCommit{
				shared.AnalyzedCommit{
					Commit: shared.Commit{
						Message: "Message",
						Author:  "Author",
						Hash:    "Hash",
					},
					ParsedMessage:               "add gitlab as relase option",
					Scope:                       "releaser",
					ParsedBreakingChangeMessage: "",
					Tag:                         "feat",
					TagString:                   "Features",
					Print:                       true,
				},
			},
		},
	}

	defer os.RemoveAll(dir)

	writeError := cache.Write(dir, content)
	assert.NoErrorf(t, writeError, "Should write file")
	result, readError := cache.Read(dir)
	assert.NoErrorf(t, readError, "Should read file")

	assert.EqualValues(t, &content, result)

}

func TestWriteNotFound(t *testing.T) {

	err := cache.Write("notfound/dir", shared.ReleaseVersion{
		Last: shared.ReleaseVersionEntry{
			Commit:  "12345",
			Version: createVersion("1.0.0"),
		},
		Next: shared.ReleaseVersionEntry{
			Commit:  "12346",
			Version: createVersion("1.1.0"),
		},
		Branch: "master",
		Draft:  true,
	})
	assert.Errorf(t, err, "Write non exsiting file")

}

func createVersion(version string) *semver.Version {
	ver, _ := semver.NewVersion(version)
	return ver
}
