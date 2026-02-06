package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/busyrockin/api-vault/core"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name> <api-key>",
	Short: "Store a new API credential",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, apiKey := args[0], args[1]
		apiType, _ := cmd.Flags().GetString("type")

		db, err := openVault()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.AddCredential(name, apiKey, apiType); err != nil {
			if errors.Is(err, core.ErrDuplicate) {
				return fmt.Errorf("credential %q already exists", name)
			}
			return fmt.Errorf("add credential: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Stored credential %q\n", name)
		fmt.Fprintln(os.Stderr, "Warning: API key may be visible in shell history")
		return nil
	},
}

func init() {
	addCmd.Flags().StringP("type", "t", "", "API type (e.g., openai, anthropic, github)")
	rootCmd.AddCommand(addCmd)
}
