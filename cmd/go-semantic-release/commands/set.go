package commands

import (
	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"

	"github.com/spf13/cobra"
)

func init() {
	setCmd.Flags().Bool("checks", false, "Check for missing values and envs")
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set [version]",
	Args:  cobra.ExactArgs(1),
	Short: "Set next release version",
	RunE: func(cmd *cobra.Command, args []string) error {

		config, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		repository, err := cmd.Flags().GetString("repository")
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
			return err
		}

		return s.SetVersion(provider, args[0])
	},
}
