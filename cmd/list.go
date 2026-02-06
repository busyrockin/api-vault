package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored credentials",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openVault()
		if err != nil {
			return err
		}
		defer db.Close()

		creds, err := db.ListCredentials()
		if err != nil {
			return fmt.Errorf("list credentials: %w", err)
		}

		if len(creds) == 0 {
			fmt.Fprintln(os.Stderr, "No credentials stored.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tCREATED")
		for _, c := range creds {
			fmt.Fprintf(w, "%s\t%s\t%s\n", c.Name, c.APIType, c.CreatedAt.Format("2006-01-02"))
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
