package releaser

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"io/ioutil"
	"net/http"
)

// Releasers struct type
type Releasers struct {
	config *config.ReleaseConfig
}

// Releaser interface for providers
type Releaser interface {
	CreateRelease(releaseName, releaseMessage string) error
}

// New initialize a Relerser
func New(c *config.ReleaseConfig) *Releasers {
	return &Releasers{
		config: c,
	}
}

//GetReleaser returns an initialized releaser
func (r *Releasers) GetReleaser(releaserType string) (Releaser, error) {
	switch releaserType {
	case GITHUB:
		return NewGitHubReleaser(r.config), nil
	}
	return nil, fmt.Errorf("Could not initialize a releaser from this type: %s", releaserType)
}

// tbd. http helper function

func makeReleaseRequest(url, authToken string, jsonRelease []byte) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonRelease))
	request.Header.Set("Authorization", authToken)
	request.Header.Set("content-type", "application/json")

	client := http.Client{}
	defer client.CloseIdleConnections()

	response, err := client.Do(request)

	if err != nil {
		return []byte{}, err

	}
	bodyContent, _ := ioutil.ReadAll(response.Body)

	if response.StatusCode >= http.StatusMultipleChoices {
		return []byte{}, fmt.Errorf("Could not create new release. HTTP %d: %s", response.StatusCode, string(bodyContent))
	}

	return bodyContent, nil
}

func uploadReleaseAssets(url, authToken string, assets []string) error {
	body := []byte{}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))

	request.Header.Set("Authorization", authToken)
	client := http.Client{}
	defer client.CloseIdleConnections()

	response, err := client.Do(request)

	if err != nil {
		return err

	}
	bodyContent, _ := ioutil.ReadAll(response.Body)

	if response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Could not create new release. HTTP %d: %s", response.StatusCode, string(bodyContent))
	}

	return nil
}

func prepareAssets(tempDir string, asset []config.Asset) ([]string, error) {
	buf := new(bytes.Buffer)
	tempAssets := []string{}
	for _, asset := range asset {
		if asset.Compress {
			fileContent, err := ioutil.ReadFile(asset.Name)
			if err != nil {
				return []string{}, err
			}

			w := zip.NewWriter(buf)
			zip, err := w.Create(tempDir + asset.Name)

			if err != nil {
				return []string{}, err
			}

			_, err = zip.Write(fileContent)
			if err != nil {
				return []string{}, err
			}
			tempAssets = append(tempAssets, tempDir+asset.Name)

		} else {
			tempAssets = append(tempAssets, asset.Name)
		}
	}
	return tempAssets, nil
}
