package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	List(ctx context.Context) ([]model.Comment, error)
	ListByTicketID(ctx context.Context, ticketID uuid.UUID) ([]model.Comment, error)
	Update(ctx context.Context, comment *model.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	Close() error
}

type commentRepository struct {
	createStmt         *sql.Stmt
	getByIDStmt        *sql.Stmt
	listStmt           *sql.Stmt
	listByTicketIDStmt *sql.Stmt
	updateStmt         *sql.Stmt
	deleteStmt         *sql.Stmt
}

const commentColumns = "id, ticket_id, author_id, body, is_internal, created_at"

const createCommentSQL = `
INSERT INTO comments (id, ticket_id, author_id, body, is_internal, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING ` + commentColumns

const getCommentByIDSQL = `
SELECT ` + commentColumns + `
FROM comments
WHERE id = $1`

const listCommentsSQL = `
SELECT ` + commentColumns + `
FROM comments
ORDER BY created_at ASC`

const listCommentsByTicketSQL = `
SELECT ` + commentColumns + `
FROM comments
WHERE ticket_id = $1
ORDER BY created_at ASC`

const updateCommentSQL = `
UPDATE comments
SET body = $1, is_internal = $2
WHERE id = $3`

const deleteCommentSQL = `
DELETE FROM comments
WHERE id = $1`

func NewCommentRepository(db *sql.DB) (CommentRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("new comment repository: db is nil")
	}

	repo := &commentRepository{}
	var err error

	repo.createStmt, err = db.PrepareContext(context.Background(), createCommentSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare create comment: %w", err)
	}
	repo.getByIDStmt, err = db.PrepareContext(context.Background(), getCommentByIDSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare get comment by id: %w", err)
	}
	repo.listStmt, err = db.PrepareContext(context.Background(), listCommentsSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare list comments: %w", err)
	}
	repo.listByTicketIDStmt, err = db.PrepareContext(context.Background(), listCommentsByTicketSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare list comments by ticket: %w", err)
	}
	repo.updateStmt, err = db.PrepareContext(context.Background(), updateCommentSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare update comment: %w", err)
	}
	repo.deleteStmt, err = db.PrepareContext(context.Background(), deleteCommentSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare delete comment: %w", err)
	}

	return repo, nil
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	if comment == nil {
		return fmt.Errorf("create comment: comment is nil")
	}

	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	comment.CreatedAt = time.Now().UTC()

	err := r.createStmt.QueryRowContext(
		ctx,
		comment.ID,
		comment.TicketID,
		comment.AuthorID,
		comment.Body,
		comment.IsInternal,
		comment.CreatedAt,
	).Scan(
		&comment.ID,
		&comment.TicketID,
		&comment.AuthorID,
		&comment.Body,
		&comment.IsInternal,
		&comment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	comment, err := scanComment(r.getByIDStmt.QueryRowContext(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("get comment by id: %w", err)
	}
	return comment, nil
}

func (r *commentRepository) List(ctx context.Context) ([]model.Comment, error) {
	rows, err := r.listStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	comments, err := collectComments(rows)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	return comments, nil
}

func (r *commentRepository) ListByTicketID(ctx context.Context, ticketID uuid.UUID) ([]model.Comment, error) {
	rows, err := r.listByTicketIDStmt.QueryContext(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list comments by ticket: %w", err)
	}
	defer rows.Close()

	comments, err := collectComments(rows)
	if err != nil {
		return nil, fmt.Errorf("list comments by ticket: %w", err)
	}
	return comments, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *model.Comment) error {
	if comment == nil {
		return fmt.Errorf("update comment: comment is nil")
	}

	result, err := r.updateStmt.ExecContext(ctx, comment.Body, comment.IsInternal, comment.ID)
	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	return rowsAffected(result, "update comment")
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.deleteStmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	return rowsAffected(result, "delete comment")
}

func (r *commentRepository) Close() error {
	return closeStmts(
		r.createStmt,
		r.getByIDStmt,
		r.listStmt,
		r.listByTicketIDStmt,
		r.updateStmt,
		r.deleteStmt,
	)
}

func collectComments(rows *sql.Rows) ([]model.Comment, error) {
	comments := []model.Comment{}
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, *comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func scanComment(row interface{ Scan(dest ...any) error }) (*model.Comment, error) {
	var comment model.Comment
	err := row.Scan(
		&comment.ID,
		&comment.TicketID,
		&comment.AuthorID,
		&comment.Body,
		&comment.IsInternal,
		&comment.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
