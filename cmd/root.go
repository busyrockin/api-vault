package cmd

import "github.com/spf13/cobra"

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "api-vault",
	Short:   "Secure credential vault for AI agents",
	Version: version,
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() error {
	return rootCmd.Execute()
}
