package assets

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Asset struct
type Asset struct {
	name         string
	path         string
	algorithm    string
	isCompressed bool
}

//NewAsset from a config
func NewAsset(repository string, assetConfig config.Asset, algorithm string) (*Asset, error) {

	filePath := assetConfig.Path
	if assetConfig.Name != "" && assetConfig.Path == "" {
		filePath = assetConfig.Name
		log.Warn("Name is deprecated. Please update your config. See https://nightapes.github.io/go-semantic-release/")
	}

	realPath := path.Join(repository, filePath)

	file, err := os.Open(realPath)
	if err != nil {
		file.Close()
		return nil, errors.Wrapf(err, "Could not open file %s", realPath)
	}
	defer file.Close()

	name := assetConfig.Rename
	if assetConfig.Rename == "" {
		info, _ := file.Stat()
		name = info.Name()
	}

	asset := &Asset{
		path:         realPath,
		name:         name,
		isCompressed: assetConfig.Compress,
		algorithm:    algorithm,
	}

	return asset, nil
}

func (a *Asset) getChecksum() (string, error) {
	log.Debugf("Calculating checksum for %s", a.path)
	file, err := os.Open(a.path)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to open file %s to calculate checksum", a.name)
	}
	defer file.Close() // nolint: errcheck
	var hash hash.Hash
	switch a.algorithm {
	case "crc32":
		hash = crc32.NewIEEE()
	case "md5":
		hash = md5.New()
	case "sha1":
		hash = sha1.New()
	case "sha224":
		hash = sha256.New224()
	case "sha384":
		hash = sha512.New384()
	case "sha256":
		hash = sha256.New()
	case "sha512":
		hash = sha512.New()
	default:
		hash = sha256.New()
	}
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetPath where the file is located, if zipped true, it will compress it and give you the new location
func (a *Asset) GetPath() (string, error) {
	if a.isCompressed {
		return a.zipFile()
	}
	return a.path, nil
}

// GetName of asset
func (a *Asset) GetName() string {
	return a.name
}

// IsCompressed return true if file was zipped
func (a *Asset) IsCompressed() bool {
	return a.isCompressed
}

// ZipFile compress given file in zip format
func (a *Asset) zipFile() (string, error) {

	path := a.path
	fileToZip, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "Could not open file %s", path)
	}
	defer fileToZip.Close()

	zipFile, err := ioutil.TempFile(os.TempDir(), "asset.*.zip")

	if err != nil {
		return "", errors.Wrap(err, "Could not generate tmp file")
	}
	log.Debugf("Created zipfile %s", zipFile.Name())

	fileToZipInfo, err := fileToZip.Stat()
	if err != nil {
		return "", errors.Wrap(err, "Could not read file infos")
	}

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileToZipHeader, err := zip.FileInfoHeader(fileToZipInfo)
	if err != nil {
		return "", errors.Wrap(err, "Could not add file infos to zip handler")
	}

	fileToZipHeader.Name = fileToZipInfo.Name()

	fileToZipWriter, err := zipWriter.CreateHeader(fileToZipHeader)
	if err != nil {
		return "", errors.Wrap(err, "Could not create zip header")
	}

	if _, err = io.Copy(fileToZipWriter, fileToZip); err != nil {
		return "", errors.Wrap(err, "Could not zip file")
	}
	if err := zipFile.Close(); err != nil {
		return "", errors.Wrap(err, "Could not close file")
	}
	return filepath.Abs(fileToZipInfo.Name())
}
