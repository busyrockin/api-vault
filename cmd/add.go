package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/busyrockin/api-vault/core"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Store a new API credential",
	Long:  "Store a credential with --secret and/or --public key, plus optional --url and --env.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		apiType, _ := cmd.Flags().GetString("type")
		secret, _ := cmd.Flags().GetString("secret")
		public, _ := cmd.Flags().GetString("public")
		url, _ := cmd.Flags().GetString("url")
		env, _ := cmd.Flags().GetString("env")

		if secret == "" && public == "" {
			return fmt.Errorf("at least one of --secret or --public is required")
		}

		cred := &core.Credential{Name: name, APIType: apiType}
		if secret != "" {
			cred.SecretKey = &secret
		}
		if public != "" {
			cred.PublicKey = &public
		}
		if url != "" {
			cred.URL = &url
		}
		if env != "" {
			cred.Environment = &env
		}

		db, err := openVault()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.AddCredentialV2(cred); err != nil {
			if errors.Is(err, core.ErrDuplicate) {
				return fmt.Errorf("credential %q already exists", name)
			}
			return fmt.Errorf("add credential: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Stored credential %q\n", name)
		if secret != "" {
			fmt.Fprintln(os.Stderr, "Warning: secret may be visible in shell history")
		}
		return nil
	},
}

func init() {
	addCmd.Flags().StringP("type", "t", "", "API type (e.g., openai, supabase, github)")
	addCmd.Flags().String("secret", "", "Secret/private API key")
	addCmd.Flags().String("public", "", "Public/anon key")
	addCmd.Flags().String("url", "", "Service URL")
	addCmd.Flags().StringP("env", "e", "", "Environment (e.g., prod, staging)")
	rootCmd.AddCommand(addCmd)
}
