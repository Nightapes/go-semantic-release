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

// Container struct
type Container struct {
	Assets     []*Asset
	repository string
	algorithm  string
}

//New container for assets
func New(repository, algorithm string) *Container {
	return &Container{
		Assets:     []*Asset{},
		repository: repository,
		algorithm:  algorithm,
	}
}

// Add assets to the list
func (a *Container) Add(assetConfigs ...config.Asset) error {
	for _, assetConfig := range assetConfigs {
		asset, err := NewAsset(a.repository, assetConfig, a.algorithm)
		if err != nil {
			return err
		}
		a.Assets = append(a.Assets, asset)
	}
	return nil
}

func (a *Container) All() []*Asset {
	return a.Assets
}

func (a *Container) GenerateChecksum() error {
	checksumFile, err := ioutil.TempFile(os.TempDir(), "checksum.*.txt")
	if err != nil {
		return errors.Wrap(err, "Could not generate tmp file for checksum")
	}
	defer checksumFile.Close()
	lines := []string{}
	for _, asset := range a.Assets {
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

	a.Assets = append(a.Assets, &Asset{
		path:         filePath,
		name:         "checksum.txt",
		isCompressed: false,
		algorithm:    "",
	})
	return w.Flush()

}
