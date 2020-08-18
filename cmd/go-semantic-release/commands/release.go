package commands

import (
	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	"github.com/spf13/cobra"
)

func init() {
	releaseCmd.Flags().Bool("no-checks", false, "Ignore missing values and envs")
	rootCmd.AddCommand(releaseCmd)
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Make a release",
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

		ignoreConfigChecks, err := cmd.Flags().GetBool("no-checks")
		if err != nil {
			return err
		}

		s, err := semanticrelease.New(readConfig(config), repository, !ignoreConfigChecks)
		if err != nil {
			return err
		}

		provider, err := s.GetCIProvider()
		if err != nil {
			return err
		}

		return s.Release(provider, force)
	},
}
