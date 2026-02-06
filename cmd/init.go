package cmd

import (
	"fmt"
	"os"

	"github.com/busyrockin/api-vault/core"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new encrypted vault",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(vaultPath); err == nil {
			return fmt.Errorf("vault already exists at %s", vaultPath)
		}

		pw, err := readPassword("Choose master password: ")
		if err != nil {
			return err
		}
		pw2, err := readPassword("Confirm master password: ")
		if err != nil {
			return err
		}
		if pw != pw2 {
			return fmt.Errorf("passwords do not match")
		}

		if err := os.MkdirAll(vaultDir, 0700); err != nil {
			return fmt.Errorf("create vault directory: %w", err)
		}

		db, err := core.NewDatabase(vaultPath, pw)
		if err != nil {
			return fmt.Errorf("create vault: %w", err)
		}
		db.Close()

		fmt.Fprintf(os.Stderr, "Vault created at %s\n", vaultPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
