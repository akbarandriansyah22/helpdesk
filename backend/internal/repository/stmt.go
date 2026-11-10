package repository

import (
	"database/sql"
	"fmt"
)

func closeStmts(stmts ...*sql.Stmt) error {
	var err error
	for _, stmt := range stmts {
		if stmt == nil {
			continue
		}
		if closeErr := stmt.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

func rowsAffected(result sql.Result, action string) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	if count == 0 {
		return fmt.Errorf("%s: %w", action, sql.ErrNoRows)
	}
	return nil
}
