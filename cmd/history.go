package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history <name>",
	Short: "Show rotation history for a credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		limit, _ := cmd.Flags().GetInt("limit")

		db, err := openVault()
		if err != nil {
			return err
		}
		defer db.Close()

		records, err := db.GetRotationHistory(name, limit)
		if err != nil {
			return fmt.Errorf("history: %w", err)
		}

		if len(records) == 0 {
			fmt.Printf("No rotation history for %q\n", name)
			return nil
		}

		for _, r := range records {
			fmt.Printf("%s  %-10s  by %-6s  fields: %s",
				r.RotatedAt.Format("2006-01-02 15:04:05"),
				r.PluginName,
				r.RotatedBy,
				strings.Join(r.RotatedFields, ", "),
			)
			if r.NewKeyID != "" {
				fmt.Printf("  key_id: %s", r.NewKeyID)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	historyCmd.Flags().IntP("limit", "n", 10, "Maximum number of records to show")
	rootCmd.AddCommand(historyCmd)
}
