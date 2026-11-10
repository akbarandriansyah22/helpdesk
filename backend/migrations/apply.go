package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed 001_init.sql 002_seed.sql
var Files embed.FS

// Apply runs every embedded SQL file in name order.
// The scripts are idempotent, so Apply is safe on every startup.
func Apply(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("apply migrations: db is nil")
	}

	names, err := fs.Glob(Files, "*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		body, err := Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		statements := splitSQL(string(body))
		for i, stmt := range statements {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("apply %s statement %d: %w", name, i+1, err)
			}
		}
	}
	return nil
}

// splitSQL splits a script on semicolons that are outside quotes and dollar quotes.
func splitSQL(src string) []string {
	var statements []string
	var b strings.Builder
	inSingle := false
	dollarTag := ""

	for i := 0; i < len(src); {
		if dollarTag != "" {
			if strings.HasPrefix(src[i:], dollarTag) {
				b.WriteString(dollarTag)
				i += len(dollarTag)
				dollarTag = ""
				continue
			}
			b.WriteByte(src[i])
			i++
			continue
		}

		if inSingle {
			b.WriteByte(src[i])
			if src[i] == '\'' {
				if i+1 < len(src) && src[i+1] == '\'' {
					b.WriteByte(src[i+1])
					i += 2
					continue
				}
				inSingle = false
			}
			i++
			continue
		}

		if src[i] == '-' && i+1 < len(src) && src[i+1] == '-' {
			for i < len(src) && src[i] != '\n' {
				b.WriteByte(src[i])
				i++
			}
			continue
		}

		if src[i] == '\'' {
			inSingle = true
			b.WriteByte(src[i])
			i++
			continue
		}

		if tag, ok := dollarTagAt(src, i); ok {
			dollarTag = tag
			b.WriteString(tag)
			i += len(tag)
			continue
		}

		if src[i] == ';' {
			appendStatement(&statements, b.String())
			b.Reset()
			i++
			continue
		}

		b.WriteByte(src[i])
		i++
	}

	appendStatement(&statements, b.String())
	return statements
}

func appendStatement(statements *[]string, raw string) {
	stmt := strings.TrimSpace(raw)
	if stmt == "" || onlyComment(stmt) {
		return
	}
	*statements = append(*statements, stmt)
}

func onlyComment(stmt string) bool {
	for _, line := range strings.Split(stmt, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			return false
		}
	}
	return true
}

func dollarTagAt(src string, i int) (string, bool) {
	if i >= len(src) || src[i] != '$' {
		return "", false
	}
	j := i + 1
	for j < len(src) {
		if src[j] == '$' {
			return src[i : j+1], true
		}
		c := src[j]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return "", false
		}
		j++
	}
	return "", false
}
