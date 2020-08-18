package hooks

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Nightapes/go-semantic-release/internal/shared"
	"github.com/Nightapes/go-semantic-release/pkg/config"
)

//Hooks struct
type Hooks struct {
	version *shared.ReleaseVersion
	config  *config.ReleaseConfig
}

// New hooks struct
func New(config *config.ReleaseConfig, version *shared.ReleaseVersion) *Hooks {
	return &Hooks{
		config:  config,
		version: version,
	}
}

// PreRelease runs before creating release
func (h *Hooks) PreRelease() error {
	log.Infof("Run pre release hooks")
	for _, cmd := range h.config.Hooks.PreRelease {
		log.Debugf("Run %s", cmd)
		err := h.runCommand(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// PostRelease runs after creating release
func (h *Hooks) PostRelease() error {
	log.Infof("Run post release hooks")
	for _, cmd := range h.config.Hooks.PostRelease {
		err := h.runCommand(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Hooks) runCommand(command string) error {

	cmdReplaced := strings.ReplaceAll(command, "$RELEASE_VERSION", h.version.Next.Version.String())

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", cmdReplaced)
	} else {
		cmd = exec.Command("sh", "-c", cmdReplaced)
	}

	cmd.Env = os.Environ()
 	cmd.Env = append(cmd.Env, "RELEASE_VERSION="+h.version.Next.Version.String())
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	h.printOutput(cmdReader, strings.Fields(cmdReplaced)[0])

	cmdErrReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	h.printOutput(cmdErrReader, strings.Fields(cmdReplaced)[0])


	return cmd.Run()
}

func (h *Hooks) printOutput(read io.ReadCloser, cmd string) {
	scanner := bufio.NewScanner(read)
	go func() {
		for scanner.Scan() {
			log.WithField("cmd", cmd).Infof("%s\n", scanner.Text())
		}
	}()
}
