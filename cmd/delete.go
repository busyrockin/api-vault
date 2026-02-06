package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/busyrockin/api-vault/core"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Remove a stored credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if !confirm(fmt.Sprintf("Delete credential %q?", name)) {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}

		db, err := openVault()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.DeleteCredential(name); err != nil {
			if errors.Is(err, core.ErrNotFound) {
				return fmt.Errorf("credential %q not found", name)
			}
			return fmt.Errorf("delete credential: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Deleted credential %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
