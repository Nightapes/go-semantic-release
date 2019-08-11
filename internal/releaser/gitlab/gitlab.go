package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Nightapes/go-semantic-release/internal/releaser/util"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"

	log "github.com/sirupsen/logrus"
)

// GITLAB identifer for gitlab interface
const GITLAB = "gitlab"

// Client type struct
type Client struct {
	config  *config.GitLabProvider
	context context.Context
	client  *http.Client
	baseURL string
	apiURL  string
	token   string
	Release string
	log     *log.Entry
}

// New initialize a new gitlabRelease
func New(config *config.GitLabProvider, accessToken string) (*Client, error) {
	ctx := context.Background()
	tokenHeader := util.NewAddHeaderTransport(nil, "PRIVATE-TOKEN", accessToken)
	acceptHeader := util.NewAddHeaderTransport(tokenHeader, "Accept", "application/json")
	httpClient := &http.Client{
		Transport: acceptHeader,
		Timeout:   time.Second * 60,
	}

	logger := log.WithField("releaser", GITLAB)

	logger.Debugf("validate gitlab provider config")

	if config.Repo == "" {
		return nil, fmt.Errorf("gitlab Repro is not set")
	}

	config.Repo = strings.Trim(config.Repo, "/")

	if config.CustomURL == "" {
		config.CustomURL = "https://gitlab.com"
	}

	config.CustomURL = strings.Trim(config.CustomURL, "/")
	logger.Debugf("Use gitlab url %s", config.CustomURL)

	return &Client{
		token:   accessToken,
		config:  config,
		context: ctx,
		baseURL: config.CustomURL,
		apiURL:  config.CustomURL + "/api/v4",
		client:  httpClient,
		log:     logger,
	}, nil
}

//GetCommitURL for gitlab
func (g *Client) GetCommitURL() string {
	return fmt.Sprintf("%s/%s/commit/{{hash}}", g.baseURL, g.config.Repo)
}

//GetCompareURL for gitlab
func (g *Client) GetCompareURL(oldVersion, newVersion string) string {
	return fmt.Sprintf("%s/%s/compare/%s...%s", g.baseURL, g.config.Repo, oldVersion, newVersion)
}

//ValidateConfig for gitlab
func (g *Client) ValidateConfig() error {
	return nil

}

// CreateRelease creates release on remote
func (g *Client) CreateRelease(releaseVersion *shared.ReleaseVersion, generatedChangelog *shared.GeneratedChangelog) error {

	tag := releaseVersion.Next.Version.String()
	g.Release = tag
	g.log.Infof("create release with version %s", tag)
	url := fmt.Sprintf("%s/projects/%s/releases", g.apiURL, util.PathEscape(g.config.Repo))
	g.log.Infof("Send release to  %s", url)

	bodyBytes, err := json.Marshal(Release{
		TagName:     tag,
		Name:        generatedChangelog.Title,
		Description: generatedChangelog.Content,
		Ref:         releaseVersion.Branch,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("could not create request: %s", err.Error())
	}

	resp, err := util.Do(g.client, req, nil)

	if err != nil {
		return fmt.Errorf("could not create release: %s", err.Error())
	}

	if err := util.IsValidResult(resp); err != nil {
		return err
	}

	log.Infof("Crated release")

	return nil
}

// UploadAssets uploads specified assets
func (g *Client) UploadAssets(repoDir string, assets []config.Asset) error {
	filesToUpload, err := util.PrepareAssets(repoDir, assets)
	if err != nil {
		return err
	}
	for _, f := range filesToUpload {

		file, err := os.Open(*f)
		if err != nil {
			return err
		}
		defer file.Close()

		fileInfo, _ := file.Stat()

		result, err := g.uploadFile(fileInfo.Name(), file)
		if err != nil {
			return fmt.Errorf("could not upload asset %s: %s", file.Name(), err.Error())
		}

		downloadURL := fmt.Sprintf("%s/%s%s", g.baseURL, g.config.Repo, result.URL)

		log.Infof("Uploaded file %s to gitlab can be downloaded under %s", file.Name(), downloadURL)

		path := fmt.Sprintf("%s/projects/%s/releases/%s/assets/links?name=%s&url=%s", g.apiURL, util.PathEscape(g.config.Repo), g.Release, util.PathEscape(fileInfo.Name()), downloadURL)

		req, err := http.NewRequest("POST", path, nil)
		if err != nil {
			return err
		}

		log.Infof("Link file %s with release %s", file.Name(), g.Release)

		resp, err := util.Do(g.client, req, nil)
		if err != nil {
			return err
		}

		if err = util.IsValidResult(resp); err != nil {
			return err
		}

		log.Infof("Link file with release %s is done", g.Release)
	}
	return nil
}

func (g *Client) uploadFile(fileName string, file *os.File) (*ProjectFile, error) {

	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)

	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		return nil, err
	}
	w.Close()

	url := fmt.Sprintf("%s/projects/%s/uploads", g.apiURL, util.PathEscape(g.config.Repo))

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(b)
	req.ContentLength = int64(b.Len())
	req.Header.Set("Content-Type", w.FormDataContentType())

	uf := &ProjectFile{}
	resp, err := util.Do(g.client, req, uf)
	if err != nil {
		return nil, err
	}

	if err = util.IsValidResult(resp); err != nil {
		return nil, err
	}

	return uf, nil
}
