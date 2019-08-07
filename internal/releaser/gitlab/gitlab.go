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
}

// New initialize a new gitlabRelease
func New(config *config.GitLabProvider) (*Client, error) {
	accessToken, err := util.GetAccessToken(GITLAB)
	if err != nil {
		return nil, err
	}

	if config.CustomURL == "" {
		config.CustomURL = "https://gitlab.com"
	}

	baseURL, err := util.CheckURL(config.CustomURL)
	log.Debugf("Use gitlab url %s", baseURL)
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
		baseURL: baseURL,
		apiURL:  baseURL + "api/v4",
		client:  httpClient,
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
	log.Debugf("validate gitlab provider config")

	if g.config.Repo == "" {
		return fmt.Errorf("gitlab Repro is not set")
	}

	return nil

}

// CreateRelease creates release on remote
func (g *Client) CreateRelease(releaseVersion *shared.ReleaseVersion, generatedChangelog *shared.GeneratedChangelog) error {

	tag := releaseVersion.Next.Version.String()
	g.release = tag
	log.Debugf("create release with version %s", tag)

	bodyBytes, err := json.Marshal(Release{
		TagName:     tag,
		Name:        generatedChangelog.Title,
		Description: generatedChangelog.Content,
		Ref:         releaseVersion.Branch,
	})
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(bodyBytes)

	url := fmt.Sprintf("%s/projects/%s/releases", g.apiURL, util.PathEscape(g.config.Repo))
	log.Debugf("Send release to  %s", url)

	resp, err := g.client.Post(url, "application/json", bodyReader)

	if err != nil {

		if !strings.Contains(err.Error(), "already_exists") && resp.StatusCode >= http.StatusUnprocessableEntity {
			return fmt.Errorf("could not create release: %v", err)
		}
		log.Infof("A release with tag %s already exits, will not perform a release or update", tag)
	} else {

		defer resp.Body.Close()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Debugf("Release repsone: %+v", string(bodyBytes))

		if err := util.IsValidResult(resp); err != nil {
			return err
		}

		log.Infof("Crated release")
	}

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
		defer file.Close()
		if err != nil {
			return err
		}
		fileInfo, _ := file.Stat()

		result, err := g.uploadFile(fileInfo.Name(), file)
		if err != nil {
			return fmt.Errorf("releaser: gitlab: Could not upload asset %s: %s", file.Name(), err.Error())
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

		log.Infof("Link file with release %s done", g.release)
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

	if err = util.IsValidResult(resp); err != nil {
		return nil, err
	}

	return uf, nil
}
