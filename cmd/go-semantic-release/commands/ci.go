package commands

import (
	"github.com/Nightapes/go-semantic-release/internal/ci"
	"github.com/Nightapes/go-semantic-release/internal/gitutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ciCmd)
}

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "ci configured artifact from release config",
	RunE: func(cmd *cobra.Command, args []string) error {

		repository, err := cmd.Flags().GetString("repository")
		if err != nil {
			return err
		}

		util, err := gitutil.New(repository)
		if err != nil {
			return err
		}

		config, err := ci.GetCIProvider(util, ci.ReadAllEnvs())
		if err != nil {
			return err
		}
		log.Infof("Found ci %v", config)

		return nil
	},
}
