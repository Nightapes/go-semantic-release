package util_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

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
	testDoubleToken{providerName: "test0", token: "foo", valid: true},
	testDoubleToken{providerName: "test1", token: "", valid: false},
}

func TestGetAccessToken(t *testing.T) {
	for _, testObject := range testDoubles {
		envName := fmt.Sprintf("%s_ACCESS_TOKEN", strings.ToUpper(testObject.providerName))
		if err := os.Setenv(envName, testObject.token); err != nil {
			fmt.Println(err.Error())
		}

		_, err := util.GetAccessToken(testObject.providerName)

		assert.Equal(t, testObject.valid, err == nil)
		os.Unsetenv(envName)
	}
}
