package commands

import (
	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.Flags().BoolP("force", "f", false, "")

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

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		s, err := semanticrelease.New(readConfig(config), repository)
		if err != nil {
			return err
		}
		return s.Release(force)
	},
}
