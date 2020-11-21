package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cPath       *string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "logprom",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cPath = rootCmd.PersistentFlags().StringP("config", "c", "", "path of config file")
	rootCmd.MarkFlagRequired("config")
}
