package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/busyrockin/api-vault/core"
	"golang.org/x/term"
)

var (
	vaultDir  string
	vaultPath string
)

func init() {
	home, _ := os.UserHomeDir()
	vaultDir = filepath.Join(home, ".api-vault")
	vaultPath = filepath.Join(vaultDir, "vault.db")
}

func readPassword(prompt string) (string, error) {
	if pw := os.Getenv("API_VAULT_PASSWORD"); pw != "" {
		return pw, nil
	}
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	if len(b) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}
	return string(b), nil
}

func openVault() (*core.Database, error) {
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("vault not found — run 'api-vault init' first")
	}
	pw, err := readPassword("Master password: ")
	if err != nil {
		return nil, err
	}
	db, err := core.NewDatabase(vaultPath, pw)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock vault (wrong password?)")
	}
	return db, nil
}

func confirm(question string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", question)
	var ans string
	fmt.Scanln(&ans)
	return ans == "y" || ans == "Y"
}
