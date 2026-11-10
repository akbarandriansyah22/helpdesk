package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
)

type TicketRepository interface {
	Create(ctx context.Context, ticket *model.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error)
	List(ctx context.Context) ([]model.Ticket, error)
	Update(ctx context.Context, ticket *model.Ticket) error
	Delete(ctx context.Context, id uuid.UUID) error
	Close() error
}

type ticketRepository struct {
	createStmt  *sql.Stmt
	getByIDStmt *sql.Stmt
	listStmt    *sql.Stmt
	updateStmt  *sql.Stmt
	deleteStmt  *sql.Stmt
}

const ticketColumns = "id, title, description, category, priority, status, requester_id, assignee_id, created_at, updated_at"

const createTicketSQL = `
INSERT INTO tickets (
	id, title, description, category, priority, status,
	requester_id, assignee_id, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING ` + ticketColumns

const getTicketByIDSQL = `
SELECT ` + ticketColumns + `
FROM tickets
WHERE id = $1`

const listTicketsSQL = `
SELECT ` + ticketColumns + `
FROM tickets
ORDER BY created_at DESC`

const updateTicketSQL = `
UPDATE tickets
SET title = $1,
	description = $2,
	category = $3,
	priority = $4,
	status = $5,
	requester_id = $6,
	assignee_id = $7,
	updated_at = $8
WHERE id = $9`

const deleteTicketSQL = `
DELETE FROM tickets
WHERE id = $1`

func NewTicketRepository(db *sql.DB) (TicketRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("new ticket repository: db is nil")
	}

	repo := &ticketRepository{}
	var err error

	repo.createStmt, err = db.PrepareContext(context.Background(), createTicketSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare create ticket: %w", err)
	}
	repo.getByIDStmt, err = db.PrepareContext(context.Background(), getTicketByIDSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare get ticket by id: %w", err)
	}
	repo.listStmt, err = db.PrepareContext(context.Background(), listTicketsSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare list tickets: %w", err)
	}
	repo.updateStmt, err = db.PrepareContext(context.Background(), updateTicketSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare update ticket: %w", err)
	}
	repo.deleteStmt, err = db.PrepareContext(context.Background(), deleteTicketSQL)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("prepare delete ticket: %w", err)
	}

	return repo, nil
}

func (r *ticketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	if ticket == nil {
		return fmt.Errorf("create ticket: ticket is nil")
	}

	now := time.Now().UTC()
	if ticket.ID == uuid.Nil {
		ticket.ID = uuid.New()
	}
	if ticket.Status == "" {
		ticket.Status = "open"
	}
	ticket.CreatedAt = now
	ticket.UpdatedAt = now

	var assignee uuid.NullUUID
	err := r.createStmt.QueryRowContext(
		ctx,
		ticket.ID,
		ticket.Title,
		ticket.Description,
		ticket.Category,
		ticket.Priority,
		ticket.Status,
		ticket.RequesterID,
		assigneeValue(ticket.AssigneeID),
		ticket.CreatedAt,
		ticket.UpdatedAt,
	).Scan(
		&ticket.ID,
		&ticket.Title,
		&ticket.Description,
		&ticket.Category,
		&ticket.Priority,
		&ticket.Status,
		&ticket.RequesterID,
		&assignee,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create ticket: %w", err)
	}
	ticket.AssigneeID = assigneePointer(assignee)
	return nil
}

func (r *ticketRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	ticket, err := scanTicket(r.getByIDStmt.QueryRowContext(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("get ticket by id: %w", err)
	}
	return ticket, nil
}

func (r *ticketRepository) List(ctx context.Context) ([]model.Ticket, error) {
	rows, err := r.listStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	tickets := []model.Ticket{}
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("list tickets: %w", err)
		}
		tickets = append(tickets, *ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	return tickets, nil
}

func (r *ticketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	if ticket == nil {
		return fmt.Errorf("update ticket: ticket is nil")
	}

	ticket.UpdatedAt = time.Now().UTC()
	result, err := r.updateStmt.ExecContext(
		ctx,
		ticket.Title,
		ticket.Description,
		ticket.Category,
		ticket.Priority,
		ticket.Status,
		ticket.RequesterID,
		assigneeValue(ticket.AssigneeID),
		ticket.UpdatedAt,
		ticket.ID,
	)
	if err != nil {
		return fmt.Errorf("update ticket: %w", err)
	}
	return rowsAffected(result, "update ticket")
}

func (r *ticketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.deleteStmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("delete ticket: %w", err)
	}
	return rowsAffected(result, "delete ticket")
}

func (r *ticketRepository) Close() error {
	return closeStmts(
		r.createStmt,
		r.getByIDStmt,
		r.listStmt,
		r.updateStmt,
		r.deleteStmt,
	)
}

func scanTicket(row interface{ Scan(dest ...any) error }) (*model.Ticket, error) {
	var (
		ticket   model.Ticket
		assignee uuid.NullUUID
	)
	err := row.Scan(
		&ticket.ID,
		&ticket.Title,
		&ticket.Description,
		&ticket.Category,
		&ticket.Priority,
		&ticket.Status,
		&ticket.RequesterID,
		&assignee,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	ticket.AssigneeID = assigneePointer(assignee)
	return &ticket, nil
}

func assigneeValue(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func assigneePointer(value uuid.NullUUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id := value.UUID
	return &id
}
