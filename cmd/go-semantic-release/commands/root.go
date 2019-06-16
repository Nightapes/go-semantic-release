package commands

import (
	"fmt"
	"os"

	"github.com/Nightapes/go-semantic-release/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-semantic-release",
	Short: "Make simple releases",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := cmd.Flags().GetString("loglevel")
		if err != nil {
			return err
		}
		setLoglevel(level)
		return nil
	},
}

//Execute rootCmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("repository", "r", "", "Path to repository")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "error", "Set loglevel")
	rootCmd.PersistentFlags().StringP("config", "c", ".release.yml", "Path to config file")
	rootCmd.PersistentFlags().Bool("no-cache", false, "Ignore cache, don't use in ci build")
}

func readConfig(file string) *config.ReleaseConfig {
	releaseConfig, err := config.Read(file)
	if err != nil {
		log.Fatal(err)
	}
	return releaseConfig
}

func setLoglevel(level string) {
	parsed, err := log.ParseLevel(level)
	if err != nil {
		log.Errorf("Invalid loglevel %s", level)
	} else {
		log.SetLevel(parsed)
	}

}
