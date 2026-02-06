package main

import (
	"fmt"
	"os"

	"github.com/busyrockin/api-vault/core"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("API Vault v0.1.0")
		fmt.Println("Usage: api-vault <command>")
		fmt.Println("Commands: version, test")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("API Vault v0.1.0-dev")
	case "test":
		runTest()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runTest() {
	const dbPath = "test-vault.db"
	defer os.Remove(dbPath)

	db, err := core.NewDatabase(dbPath, "my-secret-password")
	if err != nil {
		fmt.Printf("❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("🔐 Database opened")

	if err := db.AddCredential("openai", "sk-test-1234567890abcdef", "openai"); err != nil {
		fmt.Printf("❌ Failed to add credential: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Credential added")

	key, err := db.GetCredential("openai")
	if err != nil {
		fmt.Printf("❌ Failed to get credential: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Retrieved key: %s\n", key)

	creds, err := db.ListCredentials()
	if err != nil {
		fmt.Printf("❌ Failed to list credentials: %v\n", err)
		os.Exit(1)
	}
	for _, c := range creds {
		fmt.Printf("✅ Listed: %s (%s)\n", c.Name, c.APIType)
	}

	db.Close()
	fmt.Println("✅ Test complete — vault encrypted, decrypted, and cleaned up")
}
