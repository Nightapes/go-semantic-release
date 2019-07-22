package util

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

//CreateBearerHTTPClient with given token
func CreateBearerHTTPClient(ctx context.Context, token string) *http.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token},
	)

	client := oauth2.NewClient(ctx, tokenSource)

	return client
}

// GetAccessToken lookup for the providers accesstoken
func GetAccessToken(providerName string) (string, error) {
	var token string
	var exists bool
	envName := fmt.Sprintf("%s_ACCESS_TOKEN", strings.ToUpper(providerName))

	log.Debugf("check if %s environment variable is set", envName)

	if token, exists = os.LookupEnv(envName); !exists {
		return "", fmt.Errorf("could not find %s in the enviroment variables. Please check if it is set", envName)
	} else if token == "" {
		return "", fmt.Errorf("token %s is set in environment variables but is empty", envName)
	}
	return token, nil
}

// PrepareAssets prepare all files before uploading
func PrepareAssets(repository string, assets []config.Asset) ([]*string, error) {
	filesToUpload := []*string{}
	for _, asset := range assets {
		if asset.Name == "" {
			return nil, fmt.Errorf("asset name declaration is empty, please check your configuration file")
		} else if asset.Compress {
			log.Debugf("Asset %s will now be compressed", asset.Name)
			log.Debugf("Repo url %s", repository)
			zipNameWithPath, err := zipFile(repository, asset.Name)
			if err != nil {
				return filesToUpload, err
			}
			filesToUpload = append(filesToUpload, &zipNameWithPath)
		} else {
			tmpFileName := fmt.Sprintf("%s/%s", repository, asset.Name)
			filesToUpload = append(filesToUpload, &tmpFileName)
		}
		log.Debugf("Add asset %s to files to upload", asset.Name)
	}
	return filesToUpload, nil
}

// ZipFile compress given file in zip format
func zipFile(repository string, file string) (string, error) {

	fileToZip, err := os.Open(repository + "/" + file)
	if err != nil {
		return "", err
	}
	defer fileToZip.Close()

	zipFileName := fmt.Sprintf("%s/%s.zip", strings.TrimSuffix(repository, "/"), file)
	zipFile, err := os.Create(zipFileName)

	if err != nil {
		return "", err
	}
	log.Debugf("Created zipfile %s", zipFile.Name())

	defer zipFile.Close()

	fileToZipInfo, err := fileToZip.Stat()
	if err != nil {
		return "", err
	}

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileToZipHeader, err := zip.FileInfoHeader(fileToZipInfo)
	if err != nil {
		return "", err
	}

	fileToZipHeader.Name = fileToZipInfo.Name()

	fileToZipWriter, err := zipWriter.CreateHeader(fileToZipHeader)
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(fileToZipWriter, fileToZip); err != nil {
		return "", err
	}

	return zipFileName, nil
}
