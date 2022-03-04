package commands

import (
	"github.com/Nightapes/go-semantic-release/internal/integrations"
	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	integrationsCmd.Flags().Bool("checks", false, "Check for missing values and envs")
	integrationsCmd.Flags().StringP("out", "o", "CHANGELOG.md", "Name of the file")
	rootCmd.AddCommand(integrationsCmd)
}

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Call integrations from config file manual",
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

		releaseConfig := readConfig(config)

		s, err := semanticrelease.New(releaseConfig, repository, configChecks)
		if err != nil {
			return err
		}

		provider, err := s.GetCIProvider()
		if err != nil {
			return err
		}

		releaseVersion, err := s.GetNextVersion(provider, force, "")
		if err != nil {
			return err
		}
		log.Debugf("Found %d commits till last release", len(releaseVersion.Commits))

		i := integrations.New(&releaseConfig.Integrations, releaseVersion)

		return i.Run()
	},
}
