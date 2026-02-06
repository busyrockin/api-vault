package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/busyrockin/api-vault/core"
	"github.com/busyrockin/api-vault/rotation"
	"github.com/spf13/cobra"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <name>",
	Short: "Rotate credentials for a stored service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		db, err := openVault()
		if err != nil {
			return err
		}
		defer db.Close()

		cred, err := db.GetCredentialV2(name)
		if err != nil {
			return fmt.Errorf("credential %q: %w", name, err)
		}

		plugin, ok := rotation.GetGlobalRegistry().Get(cred.APIType)
		if !ok {
			return fmt.Errorf("no rotation plugin for api_type %q (available: %s)",
				cred.APIType, strings.Join(rotation.GetGlobalRegistry().List(), ", "))
		}

		info := rotation.CredentialInfo{
			Name:      cred.Name,
			APIType:   cred.APIType,
			SecretKey: cred.SecretKey,
			PublicKey: cred.PublicKey,
			URL:       cred.URL,
			Config:    cred.Config,
		}

		if err := plugin.Validate(info); err != nil {
			return fmt.Errorf("validation: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := plugin.Rotate(ctx, info, nil)
		if err != nil {
			return fmt.Errorf("rotate: %w", err)
		}

		coreResult := &core.RotationResult{
			NewSecretKey: result.NewSecretKey,
			NewPublicKey: result.NewPublicKey,
			NewURL:       result.NewURL,
			KeyID:        result.KeyID,
			OldKeyGrace:  result.OldKeyGrace,
			Metadata:     result.Metadata,
		}

		if err := db.RotateCredential(name, coreResult, plugin.Name(), "cli"); err != nil {
			return fmt.Errorf("save rotation: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Rotated %q via %s plugin\n", name, plugin.Name())
		if result.KeyID != "" {
			fmt.Fprintf(os.Stderr, "  Key ID: %s\n", result.KeyID)
		}
		if result.OldKeyGrace > 0 {
			fmt.Fprintf(os.Stderr, "  Old key grace period: %s\n", result.OldKeyGrace)
		}

		var fields []string
		if result.NewSecretKey != nil {
			fields = append(fields, "secret_key")
		}
		if result.NewPublicKey != nil {
			fields = append(fields, "public_key")
		}
		if result.NewURL != nil {
			fields = append(fields, "url")
		}
		fmt.Fprintf(os.Stderr, "  Rotated fields: %s\n", strings.Join(fields, ", "))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rotateCmd)
}
