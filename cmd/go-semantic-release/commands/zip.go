package commands

import (
	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(zipCmd)
}

var zipCmd = &cobra.Command{
	Use:   "zip",
	Short: "Zip configured artifact from release config (internal)",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		repository, err := cmd.Flags().GetString("repository")
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

		if err = s.ZipFiles(); err != nil {
			return err
		}

		return nil
	},
}
