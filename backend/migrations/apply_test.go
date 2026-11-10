package migrations

import (
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSplitInitSQLKeepsDoBlocksIntact(t *testing.T) {
	body, err := Files.ReadFile("001_init.sql")
	if err != nil {
		t.Fatal(err)
	}

	statements := splitSQL(string(body))
	if len(statements) < 5 {
		t.Fatalf("expected several statements, got %d", len(statements))
	}

	doBlocks := 0
	for _, stmt := range statements {
		if strings.Count(stmt, "$$")%2 != 0 {
			t.Fatalf("unbalanced dollar quotes:\n%s", stmt)
		}
		if strings.Contains(stmt, "DO $$") {
			doBlocks++
			if !strings.Contains(stmt, "END $$") {
				t.Fatalf("DO block was split:\n%s", stmt)
			}
		}
	}
	if doBlocks != 3 {
		t.Fatalf("DO blocks = %d, want 3", doBlocks)
	}
}

func TestSeedIsIdempotentAndPasswordsMatch(t *testing.T) {
	body, err := Files.ReadFile("002_seed.sql")
	if err != nil {
		t.Fatal(err)
	}
	script := string(body)

	for _, email := range []string{
		"requester@example.com",
		"agent@example.com",
		"admin@example.com",
	} {
		if !strings.Contains(script, email) {
			t.Errorf("seed missing %s", email)
		}
	}
	if !strings.Contains(script, "ON CONFLICT (email) DO NOTHING") {
		t.Fatal("seed is not idempotent")
	}

	hashes := regexp.MustCompile(`\$2[ab]\$\d{2}\$[./A-Za-z0-9]{53}`).FindAll(body, -1)
	if len(hashes) != 3 {
		t.Fatalf("bcrypt hashes = %d, want 3", len(hashes))
	}
	for _, hash := range hashes {
		if err := bcrypt.CompareHashAndPassword(hash, []byte("password123")); err != nil {
			t.Fatalf("seed hash is not bcrypt(password123): %v", err)
		}
	}
}
