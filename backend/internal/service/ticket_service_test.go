package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
)

func TestStatusTransitions(t *testing.T) {
	allowed := map[string]map[string]map[string]bool{
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
	statuses := []string{StatusOpen, StatusInProgress, StatusResolved, StatusClosed}
	roles := []string{RoleRequester, RoleAgent, RoleAdmin}

	for _, from := range statuses {
		for _, to := range statuses {
			if from == to {
				continue
			}
			for _, role := range roles {
				want := allowed[from][to][role]
				t.Run(role+" "+from+" to "+to, func(t *testing.T) {
					h := newTicketHarness(t)
					owner := h.addUser("Owner", "owner@example.com", RoleRequester)
					agent := h.addUser("Agent", "agent@example.com", RoleAgent)
					admin := h.addUser("Admin", "admin@example.com", RoleAdmin)
					actor := owner
					switch role {
					case RoleAgent:
						actor = agent
					case RoleAdmin:
						actor = admin
					}
					ticket := h.addTicket(owner.ID, from, "medium", "Network outage", time.Now().UTC(), nil)

					got, err := h.svc.UpdateStatus(context.Background(), actor.ID, role, ticket.ID, to)
					if want {
						if err != nil {
							t.Fatalf("expected allow, got %v", err)
						}
						if got.Status != to {
							t.Fatalf("status = %s, want %s", got.Status, to)
						}
						return
					}
					if !errors.Is(err, ErrInvalidTransition) {
						t.Fatalf("expected invalid transition, got %v", err)
					}
					if !strings.Contains(err.Error(), from) || !strings.Contains(err.Error(), to) || !strings.Contains(err.Error(), role) {
						t.Fatalf("error message %q is missing from, to, or role", err.Error())
					}
				})
			}
		}
	}
}

func TestUpdateStatusSameValueDoesNotWrite(t *testing.T) {
	h := newTicketHarness(t)
	owner := h.addUser("Owner", "owner@example.com", RoleRequester)
	created := time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)
	ticket := h.addTicket(owner.ID, StatusOpen, "low", "Same status", created, nil)

	got, err := h.svc.UpdateStatus(context.Background(), owner.ID, RoleRequester, ticket.ID, StatusOpen)
	if err != nil {
		t.Fatal(err)
	}
	if !got.UpdatedAt.Equal(created) {
		t.Fatal("unchanged status should not write the ticket")
	}
}

func TestUpdateStatusRejectsUnknownStatus(t *testing.T) {
	h := newTicketHarness(t)
	_, err := h.svc.UpdateStatus(context.Background(), uuid.New(), RoleAgent, uuid.New(), "done")
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("got %v", err)
	}
}

func TestRequesterCannotAccessAnotherTicket(t *testing.T) {
	h := newTicketHarness(t)
	owner := h.addUser("Owner", "owner@example.com", RoleRequester)
	other := h.addUser("Other", "other@example.com", RoleRequester)
	agent := h.addUser("Agent", "agent@example.com", RoleAgent)
	admin := h.addUser("Admin", "admin@example.com", RoleAdmin)
	ticket := h.addTicket(owner.ID, StatusOpen, "high", "Private ticket", time.Now().UTC(), nil)

	_, err := h.svc.GetTicket(context.Background(), other.ID, RoleRequester, ticket.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("get other ticket: got %v", err)
	}

	_, err = h.svc.UpdateStatus(context.Background(), other.ID, RoleRequester, ticket.ID, StatusClosed)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("close other ticket: got %v", err)
	}

	if _, err := h.svc.GetTicket(context.Background(), owner.ID, RoleRequester, ticket.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.svc.GetTicket(context.Background(), agent.ID, RoleAgent, ticket.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.svc.GetTicket(context.Background(), admin.ID, RoleAdmin, ticket.ID); err != nil {
		t.Fatal(err)
	}
}

func TestListTicketsFiltersPaginationAndScope(t *testing.T) {
	h := newTicketHarness(t)
	owner := h.addUser("Owner", "owner@example.com", RoleRequester)
	other := h.addUser("Other", "other@example.com", RoleRequester)
	agent := h.addUser("Agent", "agent@example.com", RoleAgent)
	base := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

	printer := h.addTicket(owner.ID, StatusOpen, "high", "Printer jam", base.Add(1*time.Hour), &agent.ID)
	offline := h.addTicket(owner.ID, StatusOpen, "low", "Printer offline", base.Add(2*time.Hour), nil)
	vpn := h.addTicket(owner.ID, StatusClosed, "high", "VPN down", base.Add(3*time.Hour), &agent.ID)
	email := h.addTicket(other.ID, StatusOpen, "high", "Email bounce", base.Add(4*time.Hour), &agent.ID)

	ctx := context.Background()
	own, err := h.svc.ListTickets(ctx, owner.ID, RoleRequester, TicketListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if own.Total != 3 || own.Page != 1 || own.Limit != 20 || own.TotalPages != 1 {
		t.Fatalf("owner page meta = %+v", own)
	}
	if ids(own.Tickets)[0] != vpn.ID {
		t.Fatalf("default sort = %v, want newest first %s", ids(own.Tickets), vpn.ID)
	}

	page, err := h.svc.ListTickets(ctx, owner.ID, RoleRequester, TicketListQuery{Page: "1", Limit: "2"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || page.TotalPages != 2 || ids(page.Tickets)[0] != vpn.ID || ids(page.Tickets)[1] != offline.ID {
		t.Fatalf("page 1 = %+v titles %v", page, titles(page.Tickets))
	}
	page2, err := h.svc.ListTickets(ctx, owner.ID, RoleRequester, TicketListQuery{Page: "2", Limit: "2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Tickets) != 1 || page2.Tickets[0].ID != printer.ID {
		t.Fatalf("page 2 = %v", titles(page2.Tickets))
	}

	openHigh, err := h.svc.ListTickets(ctx, owner.ID, RoleRequester, TicketListQuery{
		Status:   "open",
		Priority: "high",
		Search:   "pRiNtEr",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(openHigh.Tickets) != 1 || openHigh.Tickets[0].ID != printer.ID {
		t.Fatalf("filtered = %v", titles(openHigh.Tickets))
	}

	unassigned, err := h.svc.ListTickets(ctx, owner.ID, RoleRequester, TicketListQuery{AssigneeID: "unassigned"})
	if err != nil {
		t.Fatal(err)
	}
	if len(unassigned.Tickets) != 1 || unassigned.Tickets[0].ID != offline.ID {
		t.Fatalf("unassigned = %v", titles(unassigned.Tickets))
	}

	assigned, err := h.svc.ListTickets(ctx, agent.ID, RoleAgent, TicketListQuery{
		AssigneeID: agent.ID.String(),
		SortBy:     "title",
		Order:      "asc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if assigned.Total != 3 || titles(assigned.Tickets)[0] != "Email bounce" {
		t.Fatalf("agent assigned sort = %v", titles(assigned.Tickets))
	}

	everyone, err := h.svc.ListTickets(ctx, agent.ID, RoleAgent, TicketListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if everyone.Total != 4 || everyone.Tickets[0].ID != email.ID {
		t.Fatalf("agent sees %+v", titles(everyone.Tickets))
	}

	onlyOwn, err := h.svc.ListTickets(ctx, other.ID, RoleRequester, TicketListQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if onlyOwn.Total != 1 || onlyOwn.Tickets[0].ID != email.ID {
		t.Fatalf("other requester sees %v", titles(onlyOwn.Tickets))
	}

	for _, query := range []TicketListQuery{
		{Status: "nope"},
		{Priority: "critical"},
		{SortBy: "created_at;drop"},
		{Order: "sideways"},
		{Page: "0"},
		{Limit: "101"},
		{AssigneeID: "not-a-uuid"},
	} {
		if _, err := h.svc.ListTickets(ctx, agent.ID, RoleAgent, query); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("query %+v got %v", query, err)
		}
	}
}

func TestAssignTicket(t *testing.T) {
	h := newTicketHarness(t)
	owner := h.addUser("Owner", "owner@example.com", RoleRequester)
	agent := h.addUser("Agent", "agent@example.com", RoleAgent)
	otherAgent := h.addUser("Other Agent", "other.agent@example.com", RoleAgent)
	admin := h.addUser("Admin", "admin@example.com", RoleAdmin)
	ticket := h.addTicket(owner.ID, StatusOpen, "urgent", "Badge printer", time.Now().UTC(), nil)
	ctx := context.Background()

	assigned, err := h.svc.UpdateTicket(ctx, agent.ID, RoleAgent, ticket.ID, TicketUpdate{
		AssigneeSet: true,
		AssigneeID:  &otherAgent.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if assigned.AssigneeID == nil || *assigned.AssigneeID != otherAgent.ID {
		t.Fatalf("assignee = %v", assigned.AssigneeID)
	}

	self, err := h.svc.UpdateTicket(ctx, agent.ID, RoleAgent, ticket.ID, TicketUpdate{
		AssigneeSet: true,
		AssigneeID:  &agent.ID,
	})
	if err != nil || self.AssigneeID == nil || *self.AssigneeID != agent.ID {
		t.Fatalf("self assign: %v %+v", err, self)
	}

	byAdmin, err := h.svc.UpdateTicket(ctx, admin.ID, RoleAdmin, ticket.ID, TicketUpdate{
		AssigneeSet: true,
		AssigneeID:  &admin.ID,
		Priority:    strPtr("low"),
	})
	if err != nil || byAdmin.Priority != "low" || byAdmin.AssigneeID == nil || *byAdmin.AssigneeID != admin.ID {
		t.Fatalf("admin assign: %v priority %s", err, byAdmin.Priority)
	}

	_, err = h.svc.UpdateTicket(ctx, agent.ID, RoleAgent, ticket.ID, TicketUpdate{
		AssigneeSet: true,
		AssigneeID:  &owner.ID,
	})
	if !errors.Is(err, ErrInvalidAssignee) {
		t.Fatalf("assign requester: got %v", err)
	}

	missing := uuid.New()
	_, err = h.svc.UpdateTicket(ctx, agent.ID, RoleAgent, ticket.ID, TicketUpdate{
		AssigneeSet: true,
		AssigneeID:  &missing,
	})
	if !errors.Is(err, ErrAssigneeNotFound) {
		t.Fatalf("missing assignee: got %v", err)
	}

	_, err = h.svc.UpdateTicket(ctx, owner.ID, RoleRequester, ticket.ID, TicketUpdate{
		AssigneeSet: true,
		AssigneeID:  &agent.ID,
	})
	if !errors.Is(err, ErrForbidden) || !strings.Contains(err.Error(), "assign") {
		t.Fatalf("requester assign: got %v", err)
	}

	cleared, err := h.svc.UpdateTicket(ctx, admin.ID, RoleAdmin, ticket.ID, TicketUpdate{AssigneeSet: true})
	if err != nil || cleared.AssigneeID != nil {
		t.Fatalf("unassign: %v assignee %v", err, cleared.AssigneeID)
	}

	_, err = h.svc.UpdateTicket(ctx, owner.ID, RoleRequester, ticket.ID, TicketUpdate{AssigneeSet: true})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("requester unassign: got %v", err)
	}
}

type ticketHarness struct {
	users   *memUsers
	tickets *memTickets
	svc     *TicketService
}

func newTicketHarness(t *testing.T) *ticketHarness {
	t.Helper()
	h := &ticketHarness{users: newMemUsers(), tickets: newMemTickets()}
	svc, err := NewTicketService(h.tickets, h.users)
	if err != nil {
		t.Fatal(err)
	}
	h.svc = svc
	return h
}

func (h *ticketHarness) addUser(name, email, role string) model.User {
	user := model.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: "secret-hash",
		Role:         role,
	}
	h.users.put(user)
	return user
}

func (h *ticketHarness) addTicket(requester uuid.UUID, status, priority, title string, created time.Time, assignee *uuid.UUID) model.Ticket {
	ticket := model.Ticket{
		ID:          uuid.New(),
		Title:       title,
		Description: "Needs attention",
		Category:    "Hardware",
		Priority:    priority,
		Status:      status,
		RequesterID: requester,
		AssigneeID:  assignee,
		CreatedAt:   created,
		UpdatedAt:   created,
	}
	h.tickets.put(ticket)
	return ticket
}

func ids(tickets []model.Ticket) []uuid.UUID {
	out := make([]uuid.UUID, len(tickets))
	for i, ticket := range tickets {
		out[i] = ticket.ID
	}
	return out
}

func titles(tickets []model.Ticket) []string {
	out := make([]string, len(tickets))
	for i, ticket := range tickets {
		out[i] = ticket.Title
	}
	return out
}

func strPtr(value string) *string {
	return &value
}
