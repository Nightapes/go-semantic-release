package integrations

import (
	"os"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestIntegrations_updateNPM(t *testing.T) {
	file, err := os.CreateTemp("", "package")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	err = os.WriteFile(file.Name(), []byte(`{
"name": "test",
"version": "0.0.0",
"license": "MIT",
"scripts": {
  "ng": "ng",
  "nx": "nx"
  }
}`), 0777)
	if err != nil {
		t.Fatal(err)
	}

	testVersion, err := semver.NewVersion("1.2.0")
	if err != nil {
		t.Fatal(err)
	}

	i := New(&config.Integrations{NPM: config.IntegrationNPM{
		Enabled: true,
		Path:    file.Name(),
	}}, &shared.ReleaseVersion{
		Next: shared.ReleaseVersionEntry{
			Version: testVersion,
		},
	})

	assert.NoError(t, i.updateNPM())
	updatedFile, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `{
"name": "test",
"version": "1.2.0",
"license": "MIT",
"scripts": {
  "ng": "ng",
  "nx": "nx"
  }
}`, string(updatedFile))

}
