package util_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Nightapes/go-semantic-release/pkg/config"

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

type testDoubleFiles struct {
	testFiles []config.Asset
	valid     bool
}

var files = []testDoubleFiles{
	testDoubleFiles{
		testFiles: []config.Asset{
			config.Asset{
				Name:     "file0",
				Compress: true,
			},
			config.Asset{
				Name:     "file1",
				Compress: true,
			},
		},
		valid: true,
	},
	testDoubleFiles{
		testFiles: []config.Asset{
			config.Asset{
				Name:     "",
				Compress: true,
			},
			config.Asset{
				Name:     "",
				Compress: false,
			},
		},
		valid: false,
	},
}

func TestPrepareAssets(t *testing.T) {
	for _, testObject := range files {
		workDir, _ := os.Getwd()
		filesToDelete := []string{}

		for _, testFile := range testObject.testFiles {

			if testFile.Name != "" {
				filesToDelete = append(filesToDelete, testFile.Name)

				file, err := os.Create(testFile.Name)
				if err != nil {
					fmt.Print(err.Error())
				}
				defer file.Close()
				if testFile.Compress {
					filesToDelete = append(filesToDelete, testFile.Name+".zip")
				}
			}

		}
		preparedFiles, err := util.PrepareAssets(workDir, testObject.testFiles)

		if err == nil {
			assert.Equal(t, 2, len(preparedFiles))
		}

		assert.Equal(t, testObject.valid, err == nil)

		for _, file := range filesToDelete {
			if err := os.Remove(file); err != nil {
				fmt.Println(err.Error())
			}

		}

	}

}
