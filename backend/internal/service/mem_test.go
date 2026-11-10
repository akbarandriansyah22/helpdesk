package service

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
	"github.com/yourname/helpdesk/internal/repository"
)

type memUsers struct {
	mu   sync.Mutex
	byID map[uuid.UUID]model.User
}

func newMemUsers() *memUsers {
	return &memUsers{byID: map[uuid.UUID]model.User{}}
}

func (m *memUsers) put(user model.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byID[user.ID] = user
}

func (m *memUsers) Create(context.Context, *model.User) error {
	return fmt.Errorf("not implemented")
}

func (m *memUsers) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	user, ok := m.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	clone := user
	return &clone, nil
}

func (m *memUsers) GetByEmail(_ context.Context, email string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, user := range m.byID {
		if user.Email == email {
			clone := user
			return &clone, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *memUsers) List(context.Context) ([]model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	users := make([]model.User, 0, len(m.byID))
	for _, user := range m.byID {
		users = append(users, user)
	}
	return users, nil
}

func (m *memUsers) Update(context.Context, *model.User) error {
	return fmt.Errorf("not implemented")
}

func (m *memUsers) Delete(context.Context, uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (m *memUsers) Close() error { return nil }

type memTickets struct {
	mu   sync.Mutex
	byID map[uuid.UUID]model.Ticket
}

func newMemTickets() *memTickets {
	return &memTickets{byID: map[uuid.UUID]model.Ticket{}}
}

func (m *memTickets) put(ticket model.Ticket) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byID[ticket.ID] = cloneTicket(ticket)
}

func (m *memTickets) Create(context.Context, *model.Ticket) error {
	return fmt.Errorf("not implemented")
}

func (m *memTickets) GetByID(_ context.Context, id uuid.UUID) (*model.Ticket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ticket, ok := m.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	clone := cloneTicket(ticket)
	return &clone, nil
}

func (m *memTickets) List(context.Context) ([]model.Ticket, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	tickets := make([]model.Ticket, 0, len(m.byID))
	for _, ticket := range m.byID {
		tickets = append(tickets, cloneTicket(ticket))
	}
	return tickets, nil
}

func (m *memTickets) ListFiltered(_ context.Context, filter repository.TicketListFilter) ([]model.Ticket, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	items := make([]model.Ticket, 0, len(m.byID))
	for _, ticket := range m.byID {
		if filter.Status != "" && ticket.Status != filter.Status {
			continue
		}
		if filter.Priority != "" && ticket.Priority != filter.Priority {
			continue
		}
		if filter.Unassigned {
			if ticket.AssigneeID != nil {
				continue
			}
		} else if filter.AssigneeID != nil && (ticket.AssigneeID == nil || *ticket.AssigneeID != *filter.AssigneeID) {
			continue
		}
		if filter.RequesterID != nil && ticket.RequesterID != *filter.RequesterID {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(ticket.Title), strings.ToLower(filter.Search)) {
			continue
		}
		items = append(items, cloneTicket(ticket))
	}

	sort.SliceStable(items, func(i, j int) bool {
		if strings.EqualFold(filter.Order, "ASC") {
			return ticketLess(items[i], items[j], filter.SortBy)
		}
		return ticketLess(items[j], items[i], filter.SortBy)
	})

	total := len(items)
	start := filter.Offset
	if start > total {
		start = total
	}
	end := start + filter.Limit
	if end > total || filter.Limit < 1 {
		end = total
	}
	if start > end {
		start = end
	}
	return append([]model.Ticket{}, items[start:end]...), total, nil
}

func (m *memTickets) Update(_ context.Context, ticket *model.Ticket) error {
	if ticket == nil {
		return fmt.Errorf("update ticket: ticket is nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[ticket.ID]; !ok {
		return sql.ErrNoRows
	}
	ticket.UpdatedAt = time.Now().UTC()
	m.byID[ticket.ID] = cloneTicket(*ticket)
	return nil
}

func (m *memTickets) Delete(context.Context, uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (m *memTickets) Close() error { return nil }

func cloneTicket(ticket model.Ticket) model.Ticket {
	if ticket.AssigneeID != nil {
		id := *ticket.AssigneeID
		ticket.AssigneeID = &id
	}
	return ticket
}

func ticketLess(a, b model.Ticket, field string) bool {
	switch field {
	case "title":
		return a.Title < b.Title
	case "status":
		return a.Status < b.Status
	case "priority":
		return a.Priority < b.Priority
	case "updated_at":
		return a.UpdatedAt.Before(b.UpdatedAt)
	default:
		return a.CreatedAt.Before(b.CreatedAt)
	}
}

type memComments struct {
	mu    sync.Mutex
	items []model.Comment
}

func newMemComments() *memComments {
	return &memComments{}
}

func (m *memComments) Create(_ context.Context, comment *model.Comment) error {
	if comment == nil {
		return fmt.Errorf("create comment: comment is nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now().UTC()
	}
	m.items = append(m.items, *comment)
	return nil
}

func (m *memComments) GetByID(context.Context, uuid.UUID) (*model.Comment, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *memComments) List(context.Context) ([]model.Comment, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *memComments) ListByTicketID(_ context.Context, ticketID uuid.UUID) ([]model.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	comments := []model.Comment{}
	for _, comment := range m.items {
		if comment.TicketID == ticketID {
			comments = append(comments, comment)
		}
	}
	return comments, nil
}

func (m *memComments) Update(context.Context, *model.Comment) error {
	return fmt.Errorf("not implemented")
}

func (m *memComments) Delete(context.Context, uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (m *memComments) Close() error { return nil }

var (
	_ repository.UserRepository    = (*memUsers)(nil)
	_ repository.TicketRepository  = (*memTickets)(nil)
	_ repository.CommentRepository = (*memComments)(nil)
)
