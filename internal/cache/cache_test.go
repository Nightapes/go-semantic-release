package cache_test

import (
	"testing"

	"github.com/Nightapes/go-semantic-release/internal/cache"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
)

func TestReadCacheNotFound(t *testing.T) {

	_, err := cache.Read("notfound/dir")
	assert.Errorf(t, err, "Read non exsiting file")

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

	content := cache.ReleaseVersion{
		Last: cache.ReleaseVersionEntry{
			Commit:  "12345",
			Version: "1.0.0",
		},
		Next: cache.ReleaseVersionEntry{
			Commit:  "12346",
			Version: "1.1.0",
		},
		Branch: "master",
		Draft:  true,
	}

	defer os.RemoveAll(dir)

	writeError := cache.Write(dir, content)
	assert.NoErrorf(t, writeError, "Should write file")
	result, readError := cache.Read(dir)
	assert.NoErrorf(t, readError, "Should read file")

	assert.Equal(t, &content, result)

}

func TestWriteNotFound(t *testing.T) {

	err := cache.Write("notfound/dir", cache.ReleaseVersion{})
	assert.Errorf(t, err, "Write non exsiting file")

}
