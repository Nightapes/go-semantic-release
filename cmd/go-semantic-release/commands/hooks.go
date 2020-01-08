package commands

import (
	"github.com/Nightapes/go-semantic-release/internal/hooks"
	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(hooksCmd)
}

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Run all hooks",
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

		releaseConfig := readConfig(config)

		s, err := semanticrelease.New(releaseConfig, repository, !ignoreConfigChecks)
		if err != nil {
			return err
		}

		provider, err := s.GetCIProvider()

		if err != nil {
			return err
		}

		releaseVersion, err := s.GetNextVersion(provider, force)
		if err != nil {
			return err
		}

		hook := hooks.New(releaseConfig, releaseVersion)

		err = hook.PreRelease()
		if err != nil {
			return err
		}

		return hook.PostRelease()
	},
}
