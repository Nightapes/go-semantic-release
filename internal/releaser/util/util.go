package util

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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
	}
	return token, nil
}

// ZipFile compress given file in zip format
func ZipFile(repository string, file string) (string, error) {

	zipFileName := fmt.Sprintf("%s/%s", strings.TrimSuffix(repository, "/"), file)
	zipFile, err := os.Create(zipFileName)

	if err != nil {
		return "", err
	}

	defer zipFile.Close()

	fileToZip, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fileToZip.Close()

	fileToZipInfo, err := fileToZip.Stat()
	if err != nil {
		return "", err
	}

	zipWriter := zip.NewWriter(zipFile)

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
