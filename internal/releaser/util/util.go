package util

import (
	"context"
	"fmt"
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
