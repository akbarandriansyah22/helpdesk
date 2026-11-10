package service

import (
	"database/sql"
	"errors"
)

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrInvalidRole         = errors.New("invalid role")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrInvalidTransition   = errors.New("status transition is not allowed")
	ErrInvalidAssignee     = errors.New("assignee must be an agent or admin")
	ErrAssigneeNotFound    = errors.New("assignee not found")
	ErrCannotChangeOwnRole = errors.New("cannot change your own role")
)

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
