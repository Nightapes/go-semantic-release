package main

import (
	"os"

	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1" // imports as package "cli"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "loglevel",
			Value: "error",
			Usage: "Set loglevel 'LEVEL",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "version commands",
			Subcommands: []cli.Command{
				{
					Name:  "set",
					Usage: "set version `VERSION`",
					Action: func(c *cli.Context) error {
						setLoglevel(c.GlobalString("loglevel"))
						path := c.String("path")
						version := c.Args().First()
						log.Infof("Version %s", version)
						return semanticrelease.SetVersion(version, path)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "config, c",
							Value: "release-config.json",
							Usage: "Load release configuration from `FILE`",
						},
						cli.StringFlag{
							Name:  "path, p",
							Usage: "`PATH` to repro ",
						},
					},
				},
				{
					Name:  "next",
					Usage: "get next `VERSION` or the set one ",
					Action: func(c *cli.Context) error {
						setLoglevel(c.GlobalString("loglevel"))
						path := c.String("path")
						return semanticrelease.GetNextVersion(path)
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "config, c",
							Value: "release-config.json",
							Usage: "Load release configuration from `FILE`",
						},
						cli.StringFlag{
							Name:  "path, p",
							Usage: "`PATH` to repro ",
						},
					},
				},
			},
		},
		{
			Name:    "release",
			Aliases: []string{},
			Usage:   "make release",
			Action: func(c *cli.Context) error {
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Value: "release-config.json",
					Usage: "Load release configuration from `FILE`",
				},
			},
		},
		{
			Name:    "init",
			Aliases: []string{},
			Usage:   "create config",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
	}

	//gitutil.GetCommits(folder)
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
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
