package core

import (
	"os"
	"path/filepath"
	"testing"
)

func tempDB(t *testing.T) (*Database, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(path, "test-password")
	if err != nil {
		t.Fatalf("NewDatabase: %v", err)
	}
	return db, path
}

func TestRoundTrip(t *testing.T) {
	db, _ := tempDB(t)
	defer db.Close()

	if err := db.AddCredential("openai", "sk-test-123", "openai"); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	got, err := db.GetCredential("openai")
	if err != nil {
		t.Fatalf("GetCredential: %v", err)
	}
	if got != "sk-test-123" {
		t.Fatalf("got %q, want %q", got, "sk-test-123")
	}
}

func TestDuplicate(t *testing.T) {
	db, _ := tempDB(t)
	defer db.Close()

	db.AddCredential("key", "val", "type")
	err := db.AddCredential("key", "val2", "type")
	if err != ErrDuplicate {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

func TestNotFound(t *testing.T) {
	db, _ := tempDB(t)
	defer db.Close()

	_, err := db.GetCredential("nope")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteAndVerify(t *testing.T) {
	db, _ := tempDB(t)
	defer db.Close()

	db.AddCredential("tmp", "secret", "generic")
	if err := db.DeleteCredential("tmp"); err != nil {
		t.Fatalf("DeleteCredential: %v", err)
	}
	if err := db.DeleteCredential("tmp"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestListCredentials(t *testing.T) {
	db, _ := tempDB(t)
	defer db.Close()

	db.AddCredential("alpha", "key-a", "openai")
	db.AddCredential("beta", "key-b", "anthropic")

	creds, err := db.ListCredentials()
	if err != nil {
		t.Fatalf("ListCredentials: %v", err)
	}
	if len(creds) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(creds))
	}
	if creds[0].Name != "alpha" || creds[1].Name != "beta" {
		t.Fatalf("unexpected order: %v, %v", creds[0].Name, creds[1].Name)
	}
}

func TestWrongPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "locked.db")
	db, err := NewDatabase(path, "correct-password")
	if err != nil {
		t.Fatalf("NewDatabase: %v", err)
	}
	db.AddCredential("secret", "val", "type")
	db.Close()

	_, err = NewDatabase(path, "wrong-password")
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
}

func TestFileUnreadableWithoutPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "encrypted.db")
	db, _ := NewDatabase(path, "strong-pass")
	db.AddCredential("test", "secret-key", "generic")
	db.Close()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	// Encrypted SQLCipher files don't start with "SQLite format 3"
	if len(raw) > 16 && string(raw[:16]) == "SQLite format 3\x00" {
		t.Fatal("database file is not encrypted — starts with plain SQLite header")
	}
}
