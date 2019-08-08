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
	release string
	log     *log.Entry
}

// New initialize a new gitlabRelease
func New(config *config.GitLabProvider) (*Client, error) {
	accessToken, err := util.GetAccessToken(GITLAB)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	tokenHeader := util.NewAddHeaderTransport(nil, "PRIVATE-TOKEN", accessToken)
	acceptHeader := util.NewAddHeaderTransport(tokenHeader, "Accept", "application/json")
	httpClient := &http.Client{
		Transport: acceptHeader,
		Timeout:   time.Second * 60,
	}

	return &Client{
		token:   accessToken,
		config:  config,
		context: ctx,
		baseURL: config.CustomURL,
		apiURL:  config.CustomURL + "api/v4",
		client:  httpClient,
		log:     log.WithField("releaser", GITLAB),
	}, nil
}

//GetCommitURL for gitlab
func (g *Client) GetCommitURL() string {
	return fmt.Sprintf("%s%s/commit/{{hash}}", g.baseURL, g.config.Repo)
}

//GetCompareURL for gitlab
func (g *Client) GetCompareURL(oldVersion, newVersion string) string {
	return fmt.Sprintf("%s%s/compare/%s...%s", g.baseURL, g.config.Repo, oldVersion, newVersion)
}

//ValidateConfig for gitlab
func (g *Client) ValidateConfig() error {
	g.log.Debugf("validate gitlab provider config")

	if g.config.Repo == "" {
		return fmt.Errorf("gitlab Repro is not set")
	}

	g.config.Repo = strings.Trim(g.config.Repo, "/")

	if g.config.CustomURL == "" {
		g.config.CustomURL = "https://gitlab.com"
	}

	g.config.CustomURL = strings.Trim(g.config.CustomURL, "/")
	g.log.Debugf("Use gitlab url %s", g.config.CustomURL)

	return nil

}

// CreateRelease creates release on remote
func (g *Client) CreateRelease(releaseVersion *shared.ReleaseVersion, generatedChangelog *shared.GeneratedChangelog) error {

	tag := releaseVersion.Next.Version.String()
	g.release = tag
	g.log.Debugf("create release with version %s", tag)
	url := fmt.Sprintf("%s/projects/%s/releases", g.apiURL, util.PathEscape(g.config.Repo))
	g.log.Debugf("Send release to  %s", url)

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
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := util.Do(g.client, req, nil)

	if err != nil {

		return fmt.Errorf("could not create release: %s", err.Error())
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debugf("Release repsone: %+v", string(respBodyBytes))

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

		downloadURL := fmt.Sprintf("%s%s%s", g.baseURL, g.config.Repo, result.URL)

		log.Infof("Uploaded file %s to gitlab can be downloaded under %s", file.Name(), downloadURL)

		path := fmt.Sprintf("%s/projects/%s/releases/%s/assets/links?name=%s&url=%s", g.apiURL, util.PathEscape(g.config.Repo), g.release, util.PathEscape(fileInfo.Name()), downloadURL)

		req, err := http.NewRequest("POST", path, nil)
		if err != nil {
			return err
		}

		log.Infof("Link file %s with release %s", file.Name(), g.release)

		resp, err := util.Do(g.client, req, nil)
		if err != nil {
			return err
		}

		if err = util.IsValidResult(resp); err != nil {
			return err
		}

		log.Infof("Link file with release %s is done", g.release)
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
