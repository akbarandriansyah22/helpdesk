package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestInternalCommentsAreHiddenFromRequester(t *testing.T) {
	h := newTicketHarness(t)
	owner := h.addUser("Owner", "owner@example.com", RoleRequester)
	other := h.addUser("Other", "other@example.com", RoleRequester)
	agent := h.addUser("Agent", "agent@example.com", RoleAgent)
	ticket := h.addTicket(owner.ID, StatusOpen, "low", "VPN access", time.Now().UTC(), nil)

	comments := newMemComments()
	svc, err := NewCommentService(comments, h.svc)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	_, err = svc.CreateComment(ctx, owner.ID, RoleRequester, ticket.ID, "internal attempt", true)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("requester internal create: got %v", err)
	}

	public, err := svc.CreateComment(ctx, owner.ID, RoleRequester, ticket.ID, "I still cannot connect", false)
	if err != nil {
		t.Fatal(err)
	}
	if public.IsInternal {
		t.Fatal("public comment was stored as internal")
	}

	internal, err := svc.CreateComment(ctx, agent.ID, RoleAgent, ticket.ID, "Checked the firewall rule", true)
	if err != nil {
		t.Fatal(err)
	}
	if !internal.IsInternal {
		t.Fatal("agent internal note was stored as public")
	}

	visible, err := svc.ListComments(ctx, owner.ID, RoleRequester, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(visible) != 1 || visible[0].Body != public.Body || visible[0].IsInternal {
		t.Fatalf("requester sees %+v", visible)
	}

	all, err := svc.ListComments(ctx, agent.ID, RoleAgent, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("agent sees %d comments", len(all))
	}

	_, err = svc.ListComments(ctx, other.ID, RoleRequester, ticket.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("other requester list: got %v", err)
	}
	_, err = svc.CreateComment(ctx, other.ID, RoleRequester, ticket.ID, "not my ticket", false)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("other requester create: got %v", err)
	}
}
