package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
	"github.com/yourname/helpdesk/internal/repository"
)

const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
)

// statusTransitions maps the current status to the next status and the roles
// allowed to make that change. A requester may only change a ticket they opened.
var statusTransitions = map[string]map[string]map[string]bool{
	StatusOpen: {
		StatusInProgress: {RoleAgent: true, RoleAdmin: true},
		StatusClosed:     {RoleRequester: true, RoleAgent: true, RoleAdmin: true},
	},
	StatusInProgress: {
		StatusOpen:     {RoleAgent: true, RoleAdmin: true},
		StatusResolved: {RoleAgent: true, RoleAdmin: true},
	},
	StatusResolved: {
		StatusInProgress: {RoleAgent: true, RoleAdmin: true},
		StatusClosed:     {RoleRequester: true, RoleAgent: true, RoleAdmin: true},
	},
	StatusClosed: {
		StatusOpen: {RoleRequester: true, RoleAgent: true, RoleAdmin: true},
	},
}

type TicketService struct {
	tickets repository.TicketRepository
	users   repository.UserRepository
}

func NewTicketService(tickets repository.TicketRepository, users repository.UserRepository) (*TicketService, error) {
	if tickets == nil || users == nil {
		return nil, fmt.Errorf("new ticket service: repository is nil")
	}
	return &TicketService{tickets: tickets, users: users}, nil
}

func (s *TicketService) CreateTicket(ctx context.Context, actorID uuid.UUID, ticket *model.Ticket) (*model.Ticket, error) {
	if ticket == nil || actorID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if err := validateTicketInput(ticket); err != nil {
		return nil, err
	}

	if _, err := s.users.GetByID(ctx, actorID); err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	ticket.ID = uuid.Nil
	ticket.RequesterID = actorID
	ticket.AssigneeID = nil
	ticket.Status = StatusOpen

	if err := s.tickets.Create(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

func (s *TicketService) GetTicket(ctx context.Context, actorID uuid.UUID, role string, ticketID uuid.UUID) (*model.Ticket, error) {
	if actorID == uuid.Nil || ticketID == uuid.Nil || !validRole(role) {
		return nil, ErrInvalidInput
	}

	ticket, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if role == RoleRequester && ticket.RequesterID != actorID {
		return nil, ErrForbidden
	}
	return ticket, nil
}

func (s *TicketService) ListTickets(ctx context.Context, actorID uuid.UUID, role string) ([]model.Ticket, error) {
	if actorID == uuid.Nil || !validRole(role) {
		return nil, ErrInvalidInput
	}

	tickets, err := s.tickets.List(ctx)
	if err != nil {
		return nil, err
	}
	if role != RoleRequester {
		return tickets, nil
	}

	owned := make([]model.Ticket, 0)
	for _, ticket := range tickets {
		if ticket.RequesterID == actorID {
			owned = append(owned, ticket)
		}
	}
	return owned, nil
}

func (s *TicketService) UpdateStatus(ctx context.Context, actorID uuid.UUID, role string, ticketID uuid.UUID, status string) (*model.Ticket, error) {
	if actorID == uuid.Nil || ticketID == uuid.Nil || !validRole(role) {
		return nil, ErrInvalidInput
	}
	if !validStatus(status) {
		return nil, ErrInvalidStatus
	}

	ticket, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if role == RoleRequester && ticket.RequesterID != actorID {
		return nil, ErrForbidden
	}
	if ticket.Status == status {
		return ticket, nil
	}
	if !transitionAllowed(ticket.Status, status, role) {
		return nil, ErrInvalidTransition
	}

	ticket.Status = status
	if err := s.tickets.Update(ctx, ticket); err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ticket, nil
}

func validateTicketInput(ticket *model.Ticket) error {
	ticket.Title = strings.TrimSpace(ticket.Title)
	ticket.Description = strings.TrimSpace(ticket.Description)
	ticket.Category = strings.TrimSpace(ticket.Category)
	ticket.Priority = strings.TrimSpace(ticket.Priority)

	if utf8.RuneCountInString(ticket.Title) < 3 || utf8.RuneCountInString(ticket.Title) > 200 {
		return ErrInvalidInput
	}
	if ticket.Description == "" || utf8.RuneCountInString(ticket.Description) > 5000 {
		return ErrInvalidInput
	}
	if utf8.RuneCountInString(ticket.Category) < 2 || utf8.RuneCountInString(ticket.Category) > 50 {
		return ErrInvalidInput
	}
	switch ticket.Priority {
	case "low", "medium", "high", "urgent":
		return nil
	default:
		return ErrInvalidInput
	}
}

func validStatus(status string) bool {
	switch status {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}

func transitionAllowed(from, to, role string) bool {
	next, ok := statusTransitions[from]
	if !ok {
		return false
	}
	roles, ok := next[to]
	if !ok {
		return false
	}
	return roles[role]
}
