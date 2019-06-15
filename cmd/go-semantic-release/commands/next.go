package commands

import (
	"fmt"

	"github.com/Nightapes/go-semantic-release/pkg/semanticrelease"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(nextCmd)
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Get next release version",
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

		s, err := semanticrelease.New(readConfig(config), repository)
		if err != nil {
			return err
		}

		releaseVersion, err := s.GetNextVersion(force)
		if err != nil {
			return err
		}
		fmt.Println(releaseVersion.Next.Version.String())
		return nil
	},
}
