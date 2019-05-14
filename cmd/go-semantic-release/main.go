// Package main as start point for go build
package main

import (
	"os"

	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app      = kingpin.New("go-semantic-release", "A command-line for releasing software")
	loglevel = app.Flag("loglevel", "Set loglevel.").Default("error").HintOptions("error", "warning", "info", "debug").Short('l').String()

	nextCommand    = app.Command("next", "Print next version")
	nextRepository = nextCommand.Flag("repository", "Path to repository").String()

	setCommand    = app.Command("set", "Set version for current build.")
	setRepository = setCommand.Flag("repository", "Path to repository").String()
	setVersion    = setCommand.Arg("version", "semver version").Required().String()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case nextCommand.FullCommand():
		setLoglevel(*loglevel)
		err := semanticrelease.GetNextVersion(*nextRepository)
		if err != nil {
			log.Fatal(err)
		}

	case setCommand.FullCommand():
		setLoglevel(*loglevel)
		log.Infof("Version %s", *setVersion)
		err := semanticrelease.SetVersion(*setVersion, *setRepository)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func setLoglevel(level string) {
	parsed, err := log.ParseLevel(level)
	if err != nil {
		log.Errorf("Invalid loglevel %s", level)
	} else {
		log.SetLevel(parsed)
	}

}
