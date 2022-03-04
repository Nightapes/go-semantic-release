package util_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"

	"github.com/Nightapes/go-semantic-release/internal/releaser/util"
)

func TestCreateBearerHTTPClient(t *testing.T) {
	client := util.CreateBearerHTTPClient(context.Background(), "")

	assert.True(t, client != nil, "Client is empty")
}

type testDoubleToken struct {
	providerName, token string
	valid               bool
}

var testDoubles = []testDoubleToken{
	{providerName: "test0", token: "foo", valid: true},
	{providerName: "test1", token: "", valid: false},
}

func TestGetAccessToken(t *testing.T) {
	for _, testObject := range testDoubles {
		envName := fmt.Sprintf("%s_ACCESS_TOKEN", strings.ToUpper(testObject.providerName))
		if err := os.Setenv(envName, testObject.token); err != nil {
			fmt.Println(err.Error())
		}

		_, err := util.GetAccessToken(envName)

		assert.Equal(t, testObject.valid, err == nil)
		os.Unsetenv(envName)
	}
}

func TestShouldRetry(t *testing.T) {
	assert.True(t, util.ShouldRetry(&http.Response{StatusCode: 429}))
	assert.False(t, util.ShouldRetry(&http.Response{StatusCode: 200}))
}

func TestIsValidResult(t *testing.T) {
	assert.NoError(t, util.IsValidResult(&http.Response{StatusCode: 200}))
	assert.NoError(t, util.IsValidResult(&http.Response{StatusCode: 201}))
	assert.NoError(t, util.IsValidResult(&http.Response{StatusCode: 202}))
	assert.NoError(t, util.IsValidResult(&http.Response{StatusCode: 204}))

	u, err := url.Parse("https://127.0.0.1")
	assert.NoError(t, err)
	assert.Error(t, util.IsValidResult(&http.Response{StatusCode: 500, Request: &http.Request{
		Method: "POST",
		URL:    u,
	}}))
}

func TestPathEscape(t *testing.T) {
	assert.Equal(t, "test%2Ftest", util.PathEscape("test/test"))
	assert.Equal(t, "test", util.PathEscape("test"))
	assert.Equal(t, "test%2Etest", util.PathEscape("test.test"))
}

type example struct {
	Test string `json:"test"`
}

func TestDoAndRoundTrip(t *testing.T) {
	tokenHeader := util.NewAddHeaderTransport(nil, "PRIVATE-TOKEN", "aToken")
	acceptHeader := util.NewAddHeaderTransport(tokenHeader, "Accept", "application/json")
	httpClient := &http.Client{
		Transport: acceptHeader,
		Timeout:   time.Second * 60,
	}

	testsDoMethod := []struct {
		statusCode       int
		body             string
		responseBody     interface{}
		responseBodyType interface{}
		hasError         bool
		path             string
	}{
		{
			statusCode: 200,
			body:       `{"test" : "hallo"}`,
			responseBody: &example{
				Test: "hallo",
			},
			responseBodyType: &example{},
			hasError:         false,
			path:             "",
		},
		{
			statusCode:       400,
			body:             `{"test" : "hallo"}`,
			responseBody:     &example{},
			responseBodyType: &example{},
			hasError:         false,
			path:             "",
		},
		{
			statusCode:       200,
			body:             `{"test" : "hallo"}`,
			hasError:         true,
			responseBody:     &example{},
			responseBodyType: &example{},
			path:             "4/broken",
		},
	}

	for _, testOject := range testsDoMethod {
		testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			log.Infof("Got call from %s %s", req.Method, req.URL.String())

			assert.Equal(t, req.Header.Get("PRIVATE-TOKEN"), "aToken")
			assert.Equal(t, req.Header.Get("Accept"), "application/json")

			rw.WriteHeader(testOject.statusCode)
			rw.Header().Set("Content-Type", "application/json")

			if _, err := rw.Write([]byte(testOject.body)); err != nil {
				log.Info(err)
			}

		}))

		defer testServer.Close()

		req, err := http.NewRequest("POST", testServer.URL+testOject.path, nil)
		assert.NoError(t, err)

		resp, err := util.Do(httpClient, req, testOject.responseBodyType)

		assert.Equal(t, testOject.hasError, err != nil)

		if !testOject.hasError {
			assert.Equal(t, testOject.statusCode, resp.StatusCode)
			assert.Equal(t, testOject.responseBody, testOject.responseBodyType)
		}
	}
}
