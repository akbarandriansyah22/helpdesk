package service

import (
	"context"
	"fmt"
	"strconv"
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

	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
	maxSearchLen = 200
)

// statusTransitions is the only allowed status graph.
// A missing entry is rejected for every role.
var statusTransitions = map[string]map[string]map[string]bool{
	StatusOpen: {
		StatusInProgress: {RoleAgent: true, RoleAdmin: true},
		StatusClosed:     {RoleRequester: true, RoleAdmin: true},
	},
	StatusInProgress: {
		StatusResolved: {RoleAgent: true, RoleAdmin: true},
	},
	StatusResolved: {
		StatusClosed: {RoleRequester: true, RoleAgent: true, RoleAdmin: true},
		StatusOpen:   {RoleRequester: true},
	},
}

type TicketUpdate struct {
	Status      *string
	Priority    *string
	AssigneeSet bool
	AssigneeID  *uuid.UUID
}

type TicketListQuery struct {
	Status     string
	Priority   string
	AssigneeID string
	Search     string
	SortBy     string
	Order      string
	Page       string
	Limit      string
}

type TicketListResult struct {
	Tickets    []model.Ticket
	Page       int
	Limit      int
	Total      int
	TotalPages int
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

func (s *TicketService) ListTickets(ctx context.Context, actorID uuid.UUID, role string, query TicketListQuery) (*TicketListResult, error) {
	if actorID == uuid.Nil || !validRole(role) {
		return nil, ErrInvalidInput
	}

	filter, page, limit, err := normalizeListQuery(query)
	if err != nil {
		return nil, err
	}
	if role == RoleRequester {
		filter.RequesterID = &actorID
	}

	tickets, total, err := s.tickets.ListFiltered(ctx, filter)
	if err != nil {
		return nil, err
	}
	if tickets == nil {
		tickets = []model.Ticket{}
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return &TicketListResult{
		Tickets:    tickets,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *TicketService) UpdateStatus(ctx context.Context, actorID uuid.UUID, role string, ticketID uuid.UUID, status string) (*model.Ticket, error) {
	if !validStatus(status) {
		return nil, ErrInvalidStatus
	}
	return s.UpdateTicket(ctx, actorID, role, ticketID, TicketUpdate{Status: &status})
}

func (s *TicketService) UpdateTicket(ctx context.Context, actorID uuid.UUID, role string, ticketID uuid.UUID, patch TicketUpdate) (*model.Ticket, error) {
	if actorID == uuid.Nil || ticketID == uuid.Nil || !validRole(role) {
		return nil, ErrInvalidInput
	}
	if patch.Status == nil && patch.Priority == nil && !patch.AssigneeSet {
		return nil, fmt.Errorf("%w: no fields to update", ErrInvalidInput)
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

	changed := false

	if patch.Status != nil {
		status := strings.TrimSpace(*patch.Status)
		if !validStatus(status) {
			return nil, ErrInvalidStatus
		}
		if status != ticket.Status {
			if !transitionAllowed(ticket.Status, status, role) {
				return nil, fmt.Errorf("%w: cannot change status from %s to %s as %s", ErrInvalidTransition, ticket.Status, status, role)
			}
			ticket.Status = status
			changed = true
		}
	}

	if patch.Priority != nil {
		priority := strings.TrimSpace(*patch.Priority)
		if !validPriority(priority) {
			return nil, fmt.Errorf("%w: priority must be one of low, medium, high, urgent", ErrInvalidInput)
		}
		if priority != ticket.Priority {
			ticket.Priority = priority
			changed = true
		}
	}

	if patch.AssigneeSet {
		if role == RoleRequester {
			return nil, fmt.Errorf("%w: requesters cannot assign or unassign tickets", ErrForbidden)
		}
		if patch.AssigneeID == nil {
			if ticket.AssigneeID != nil {
				ticket.AssigneeID = nil
				changed = true
			}
		} else {
			if *patch.AssigneeID == uuid.Nil {
				return nil, fmt.Errorf("%w: assignee_id must be a UUID or null", ErrInvalidInput)
			}
			assignee, err := s.users.GetByID(ctx, *patch.AssigneeID)
			if err != nil {
				if isNotFound(err) {
					return nil, ErrAssigneeNotFound
				}
				return nil, err
			}
			if assignee.Role != RoleAgent && assignee.Role != RoleAdmin {
				return nil, ErrInvalidAssignee
			}
			if ticket.AssigneeID == nil || *ticket.AssigneeID != *patch.AssigneeID {
				id := *patch.AssigneeID
				ticket.AssigneeID = &id
				changed = true
			}
		}
	}

	if !changed {
		return ticket, nil
	}
	if err := s.tickets.Update(ctx, ticket); err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ticket, nil
}

func normalizeListQuery(query TicketListQuery) (repository.TicketListFilter, int, int, error) {
	page, err := parsePositive(query.Page, defaultPage, "page")
	if err != nil {
		return repository.TicketListFilter{}, 0, 0, err
	}
	limit, err := parsePositive(query.Limit, defaultLimit, "limit")
	if err != nil {
		return repository.TicketListFilter{}, 0, 0, err
	}
	if limit > maxLimit {
		return repository.TicketListFilter{}, 0, 0, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidInput, maxLimit)
	}

	status := strings.ToLower(strings.TrimSpace(query.Status))
	if status != "" && !validStatus(status) {
		return repository.TicketListFilter{}, 0, 0, fmt.Errorf("%w: status must be one of open, in_progress, resolved, closed", ErrInvalidInput)
	}
	priority := strings.ToLower(strings.TrimSpace(query.Priority))
	if priority != "" && !validPriority(priority) {
		return repository.TicketListFilter{}, 0, 0, fmt.Errorf("%w: priority must be one of low, medium, high, urgent", ErrInvalidInput)
	}

	search := strings.TrimSpace(query.Search)
	if utf8.RuneCountInString(search) > maxSearchLen {
		return repository.TicketListFilter{}, 0, 0, fmt.Errorf("%w: search must be at most %d characters", ErrInvalidInput, maxSearchLen)
	}

	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	if sortBy == "" {
		sortBy = "created_at"
	}
	switch sortBy {
	case "created_at", "updated_at", "priority", "status", "title":
	default:
		return repository.TicketListFilter{}, 0, 0, fmt.Errorf("%w: sort_by must be one of created_at, updated_at, priority, status, title", ErrInvalidInput)
	}

	order := strings.ToLower(strings.TrimSpace(query.Order))
	if order == "" {
		order = "desc"
	}
	if order != "asc" && order != "desc" {
		return repository.TicketListFilter{}, 0, 0, fmt.Errorf("%w: order must be asc or desc", ErrInvalidInput)
	}

	filter := repository.TicketListFilter{
		Status:   status,
		Priority: priority,
		Search:   search,
		SortBy:   sortBy,
		Order:    strings.ToUpper(order),
		Limit:    limit,
		Offset:   (page - 1) * limit,
	}

	assignee := strings.TrimSpace(query.AssigneeID)
	if strings.EqualFold(assignee, "unassigned") {
		filter.Unassigned = true
	} else if assignee != "" {
		id, err := uuid.Parse(assignee)
		if err != nil || id == uuid.Nil {
			return repository.TicketListFilter{}, 0, 0, fmt.Errorf("%w: assignee_id must be a UUID or \"unassigned\"", ErrInvalidInput)
		}
		filter.AssigneeID = &id
	}
	return filter, page, limit, nil
}

func parsePositive(raw string, fallback int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%w: %s must be an integer greater than or equal to 1", ErrInvalidInput, name)
	}
	return n, nil
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
	if !validPriority(ticket.Priority) {
		return ErrInvalidInput
	}
	return nil
}

func validPriority(priority string) bool {
	switch priority {
	case "low", "medium", "high", "urgent":
		return true
	default:
		return false
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
