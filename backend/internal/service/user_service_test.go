package service

import (
	"context"
	"testing"
)

func TestGetMeOmitsPasswordHash(t *testing.T) {
	h := newTicketHarness(t)
	user := h.addUser("Requester", "requester@example.com", RoleRequester)
	svc, err := NewUserService(h.users)
	if err != nil {
		t.Fatal(err)
	}

	got, err := svc.GetMe(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.PasswordHash != "" {
		t.Fatalf("password hash leaked: %q", got.PasswordHash)
	}
	if got.Email != user.Email || got.Role != RoleRequester || got.Name != user.Name {
		t.Fatalf("profile = %+v", got)
	}
}
