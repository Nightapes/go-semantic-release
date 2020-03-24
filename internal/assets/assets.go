package assets

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/pkg/errors"
)

// Set struct
type Set struct {
	Assets     []*Asset
	repository string
	algorithm  string
}

//New container for assets
func New(repository, algorithm string) *Set {
	return &Set{
		Assets:     []*Asset{},
		repository: repository,
		algorithm:  algorithm,
	}
}

// Add assets to the list
func (s *Set) Add(assetConfigs ...config.Asset) error {
	for _, assetConfig := range assetConfigs {
		asset, err := NewAsset(s.repository, assetConfig, s.algorithm)
		if err != nil {
			return err
		}
		s.Assets = append(s.Assets, asset)
	}
	return nil
}

func (s *Set) All() []*Asset {
	return s.Assets
}

func (s *Set) GenerateChecksum() error {
	checksumFile, err := ioutil.TempFile(os.TempDir(), "checksum.*.txt")
	if err != nil {
		return errors.Wrap(err, "Could not generate tmp file for checksum")
	}
	defer checksumFile.Close()
	lines := []string{}
	for _, asset := range s.Assets {
		checksum, err := asset.getChecksum()
		if err != nil {
			return err
		}
		lines = append(lines, fmt.Sprintf("%s %s", checksum, asset.GetName()))
	}

	w := bufio.NewWriter(checksumFile)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	filePath, err := filepath.Abs(checksumFile.Name())
	if err != nil {
		return err
	}

	s.Assets = append(s.Assets, &Asset{
		path:         filePath,
		name:         "checksum.txt",
		isCompressed: false,
		algorithm:    "",
	})
	return w.Flush()

}
