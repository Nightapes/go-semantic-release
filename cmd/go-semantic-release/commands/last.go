package commands

import (
	"fmt"
	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	lastCmd.Flags().Bool("checks", false, "Check for missing values and envs")
	rootCmd.AddCommand(lastCmd)
}

var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "Get last version",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		repository, err := cmd.Flags().GetString("repository")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("no-cache")
		if err != nil {
			return err
		}

		configChecks, err := cmd.Flags().GetBool("checks")
		if err != nil {
			return err
		}

		s, err := semanticrelease.New(readConfig(config), repository, configChecks)
		if err != nil {
			return err
		}

		provider, err := s.GetCIProvider()

		if err != nil {
			log.Infof("Will not calculate version, set fake version. Could not find CI Provider, if running locally, set env CI=true")
			fmt.Println("0.0.0-fake.0")
			return nil
		}

		releaseVersion, err := s.GetNextVersion(provider, force, "")
		if err != nil {
			return err
		}
		fmt.Println(releaseVersion.Last.Version.String())
		return nil
	},
}
