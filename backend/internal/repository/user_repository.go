package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Close() error
}

type userRepository struct {
	createStmt     *sql.Stmt
	getByIDStmt    *sql.Stmt
	getByEmailStmt *sql.Stmt
	listStmt       *sql.Stmt
	updateStmt     *sql.Stmt
	deleteStmt     *sql.Stmt
}

const userColumns = "id, name, email, password_hash, role, created_at, updated_at"

const createUserSQL = `
INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING ` + userColumns

const getUserByIDSQL = `
SELECT ` + userColumns + `
FROM users
WHERE id = $1`

const getUserByEmailSQL = `
SELECT ` + userColumns + `
FROM users
WHERE email = $1`

const listUsersSQL = `
SELECT ` + userColumns + `
FROM users
ORDER BY created_at DESC`

const updateUserSQL = `
UPDATE users
SET name = $1, email = $2, password_hash = $3, role = $4, updated_at = $5
WHERE id = $6`

const deleteUserSQL = `
DELETE FROM users
WHERE id = $1`

func NewUserRepository(db *sql.DB) (UserRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("new user repository: db is nil")
	}

	repo := &userRepository{}
	var err error

	repo.createStmt, err = db.PrepareContext(context.Background(), createUserSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare create user: %w", err)
	}
	repo.getByIDStmt, err = db.PrepareContext(context.Background(), getUserByIDSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare get user by id: %w", err)
	}
	repo.getByEmailStmt, err = db.PrepareContext(context.Background(), getUserByEmailSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare get user by email: %w", err)
	}
	repo.listStmt, err = db.PrepareContext(context.Background(), listUsersSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare list users: %w", err)
	}
	repo.updateStmt, err = db.PrepareContext(context.Background(), updateUserSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare update user: %w", err)
	}
	repo.deleteStmt, err = db.PrepareContext(context.Background(), deleteUserSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare delete user: %w", err)
	}

	return repo, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if user == nil {
		return fmt.Errorf("create user: user is nil")
	}

	now := time.Now().UTC()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.createStmt.QueryRowContext(
		ctx,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := scanUser(r.getByIDStmt.QueryRowContext(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := scanUser(r.getByEmailStmt.QueryRowContext(ctx, email))
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *userRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.listStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		users = append(users, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if user == nil {
		return fmt.Errorf("update user: user is nil")
	}

	user.UpdatedAt = time.Now().UTC()
	result, err := r.updateStmt.ExecContext(
		ctx,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return rowsAffected(result, "update user")
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.deleteStmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return rowsAffected(result, "delete user")
}

func (r *userRepository) Close() error {
	return closeStmts(
		r.createStmt,
		r.getByIDStmt,
		r.getByEmailStmt,
		r.listStmt,
		r.updateStmt,
		r.deleteStmt,
	)
}

func scanUser(row interface{ Scan(dest ...any) error }) (*model.User, error) {
	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
