package cmd

import (
	"errors"
	"fmt"

	"github.com/busyrockin/api-vault/core"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Retrieve a decrypted API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openVault()
		if err != nil {
			return err
		}
		defer db.Close()

		key, err := db.GetCredential(args[0])
		if err != nil {
			if errors.Is(err, core.ErrNotFound) {
				return fmt.Errorf("credential %q not found", args[0])
			}
			return fmt.Errorf("get credential: %w", err)
		}

		fmt.Print(key)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
