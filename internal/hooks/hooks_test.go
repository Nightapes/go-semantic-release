package hooks_test

import (
	"os"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/Nightapes/go-semantic-release/internal/hooks"
	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestPreReleaseHooks(t *testing.T) {

	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	hooks := hooks.New(&config.ReleaseConfig{
		Hooks: config.Hooks{
			PreRelease: []string{
				"go test -test.run=TestHelperProcess -- " + "$RELEASE_VERSION",
			},
		},
	},
		&shared.ReleaseVersion{
			Next: shared.ReleaseVersionEntry{
				Version: createVersion("1.0.0"),
			},
		})
	err := hooks.PreRelease()
	assert.NoError(t, err)
	os.Unsetenv("GO_WANT_HELPER_PROCESS")

}

func TestPreReleaseHooksError(t *testing.T) {

	hooks := hooks.New(&config.ReleaseConfig{
		Hooks: config.Hooks{
			PreRelease: []string{
				"exit 1",
			},
		},
	},
		&shared.ReleaseVersion{
			Next: shared.ReleaseVersionEntry{
				Version: createVersion("1.0.0"),
			},
		})
	err := hooks.PreRelease()
	assert.Error(t, err)
}

func TestPostReleaseHooks(t *testing.T) {

	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	hooks := hooks.New(&config.ReleaseConfig{
		Hooks: config.Hooks{
			PostRelease: []string{
				"go test -test.run=TestHelperProcess -- " + "$RELEASE_VERSION",
			},
		},
	},
		&shared.ReleaseVersion{
			Next: shared.ReleaseVersionEntry{
				Version: createVersion("1.0.0"),
			},
		})
	err := hooks.PostRelease()
	assert.NoError(t, err)
	os.Unsetenv("GO_WANT_HELPER_PROCESS")

}

func TestPostReleaseHooksError(t *testing.T) {

	hooks := hooks.New(&config.ReleaseConfig{
		Hooks: config.Hooks{
			PostRelease: []string{
				"exit 1",
			},
		},
	},
		&shared.ReleaseVersion{
			Next: shared.ReleaseVersionEntry{
				Version: createVersion("1.0.0"),
			},
		})
	err := hooks.PostRelease()
	assert.Error(t, err)
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {

		return
	}
	assert.Equal(t, "1.0.0", os.Args[3])
}

func createVersion(version string) *semver.Version {
	ver, _ := semver.NewVersion(version)
	return ver
}
