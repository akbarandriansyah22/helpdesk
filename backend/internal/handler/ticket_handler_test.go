package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/middleware"
	"github.com/yourname/helpdesk/internal/model"
	"github.com/yourname/helpdesk/internal/repository"
	"github.com/yourname/helpdesk/internal/service"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestListPassesFiltersAndPagination(t *testing.T) {
	tickets := &recordingTickets{}
	handler := newTicketHandler(t, tickets, stubUsers{})
	actor := uuid.New()

	w := perform(http.MethodGet, "/api/tickets?status=open&priority=high&assignee_id=unassigned&search=Printer&sort_by=title&order=asc&page=2&limit=5", nil, actor, service.RoleAgent, handler.List)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}

	filter := tickets.filter
	if filter.Status != "open" || filter.Priority != "high" || !filter.Unassigned || filter.Search != "Printer" {
		t.Fatalf("filter = %+v", filter)
	}
	if filter.SortBy != "title" || filter.Order != "ASC" || filter.Limit != 5 || filter.Offset != 5 {
		t.Fatalf("page filter = %+v", filter)
	}
	if filter.RequesterID != nil {
		t.Fatal("agent list was scoped to a requester")
	}

	var payload struct {
		Meta struct {
			Page       int `json:"page"`
			Limit      int `json:"limit"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Meta.Page != 2 || payload.Meta.Limit != 5 || payload.Meta.Total != 0 {
		t.Fatalf("meta = %+v", payload.Meta)
	}
}

func TestRequesterListIsScopedToSelf(t *testing.T) {
	tickets := &recordingTickets{}
	handler := newTicketHandler(t, tickets, stubUsers{})
	actor := uuid.New()

	w := perform(http.MethodGet, "/api/tickets", nil, actor, service.RoleRequester, handler.List)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if tickets.filter.RequesterID == nil || *tickets.filter.RequesterID != actor {
		t.Fatalf("requester scope = %v", tickets.filter.RequesterID)
	}
}

func TestPatchRejectsIllegalTransition(t *testing.T) {
	owner := uuid.New()
	ticketID := uuid.New()
	tickets := &recordingTickets{ticket: &model.Ticket{
		ID:          ticketID,
		Title:       "Printer jam",
		Status:      "open",
		Priority:    "high",
		RequesterID: owner,
	}}
	handler := newTicketHandler(t, tickets, stubUsers{})

	w := perform(http.MethodPatch, "/api/tickets/"+ticketID.String(), []byte(`{"status":"closed"}`), uuid.New(), service.RoleAgent, handler.Update, gin.Param{Key: "id", Value: ticketID.String()})
	if w.Code != http.StatusConflict {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "open") || !strings.Contains(w.Body.String(), "closed") || !strings.Contains(w.Body.String(), "agent") {
		t.Fatalf("body %s", w.Body.String())
	}
	if tickets.updated {
		t.Fatal("illegal transition was persisted")
	}
}

func TestRequesterCannotAssign(t *testing.T) {
	owner := uuid.New()
	ticketID := uuid.New()
	tickets := &recordingTickets{ticket: &model.Ticket{
		ID:          ticketID,
		Status:      "open",
		Priority:    "low",
		RequesterID: owner,
	}}
	handler := newTicketHandler(t, tickets, stubUsers{})
	body := []byte(`{"assignee_id":"22222222-2222-4222-8222-222222222222"}`)

	w := perform(http.MethodPatch, "/api/tickets/"+ticketID.String(), body, owner, service.RoleRequester, handler.Update, gin.Param{Key: "id", Value: ticketID.String()})
	if w.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "assign") {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestGetHidesAnotherRequestersTicket(t *testing.T) {
	ticketID := uuid.New()
	tickets := &recordingTickets{ticket: &model.Ticket{
		ID:          ticketID,
		Status:      "open",
		RequesterID: uuid.New(),
	}}
	handler := newTicketHandler(t, tickets, stubUsers{})

	w := perform(http.MethodGet, "/api/tickets/"+ticketID.String(), nil, uuid.New(), service.RoleRequester, handler.Get, gin.Param{Key: "id", Value: ticketID.String()})
	if w.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestMeResponseHasNoPassword(t *testing.T) {
	userID := uuid.New()
	users := stubUsers{user: &model.User{
		ID:           userID,
		Name:         "Requester",
		Email:        "requester@example.com",
		PasswordHash: "super-secret-hash",
		Role:         service.RoleRequester,
	}}
	svc, err := service.NewUserService(users)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewUserHandler(svc)

	w := perform(http.MethodGet, "/api/me", nil, userID, service.RoleRequester, handler.Me)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "super-secret-hash") || strings.Contains(w.Body.String(), "password") {
		t.Fatalf("password leaked: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "requester@example.com") {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestWriteErrorStatusCodes(t *testing.T) {
	cases := []struct {
		err  error
		code int
	}{
		{fmt.Errorf("%w: cannot change status from open to closed as agent", service.ErrInvalidTransition), http.StatusConflict},
		{fmt.Errorf("%w: requesters cannot assign or unassign tickets", service.ErrForbidden), http.StatusForbidden},
		{fmt.Errorf("%w: page must be an integer greater than or equal to 1", service.ErrInvalidInput), http.StatusBadRequest},
		{service.ErrInvalidStatus, http.StatusBadRequest},
		{service.ErrInvalidAssignee, http.StatusBadRequest},
		{service.ErrNotFound, http.StatusNotFound},
		{service.ErrAssigneeNotFound, http.StatusNotFound},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		writeError(c, tc.err)
		if w.Code != tc.code {
			t.Errorf("%v -> %d, want %d", tc.err, w.Code, tc.code)
		}
	}
}

func newTicketHandler(t *testing.T, tickets repository.TicketRepository, users repository.UserRepository) *TicketHandler {
	t.Helper()
	svc, err := service.NewTicketService(tickets, users)
	if err != nil {
		t.Fatal(err)
	}
	return NewTicketHandler(svc)
}

func perform(method, target string, body []byte, actor uuid.UUID, role string, handle gin.HandlerFunc, params ...gin.Param) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	c.Set(middleware.ContextUserID, actor)
	c.Set(middleware.ContextRole, role)
	handle(c)
	return w
}

type recordingTickets struct {
	ticket  *model.Ticket
	filter  repository.TicketListFilter
	updated bool
}

func (r *recordingTickets) Create(context.Context, *model.Ticket) error {
	return fmt.Errorf("not implemented")
}

func (r *recordingTickets) GetByID(_ context.Context, id uuid.UUID) (*model.Ticket, error) {
	if r.ticket == nil || r.ticket.ID != id {
		return nil, sql.ErrNoRows
	}
	clone := *r.ticket
	return &clone, nil
}

func (r *recordingTickets) List(context.Context) ([]model.Ticket, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *recordingTickets) ListFiltered(_ context.Context, filter repository.TicketListFilter) ([]model.Ticket, int, error) {
	r.filter = filter
	return []model.Ticket{}, 0, nil
}

func (r *recordingTickets) Update(context.Context, *model.Ticket) error {
	r.updated = true
	return nil
}

func (r *recordingTickets) Delete(context.Context, uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *recordingTickets) Close() error { return nil }

type stubUsers struct {
	user *model.User
}

func (s stubUsers) Create(context.Context, *model.User) error { return fmt.Errorf("not implemented") }

func (s stubUsers) GetByID(context.Context, uuid.UUID) (*model.User, error) {
	if s.user == nil {
		return nil, sql.ErrNoRows
	}
	clone := *s.user
	return &clone, nil
}

func (s stubUsers) GetByEmail(context.Context, string) (*model.User, error) {
	return nil, sql.ErrNoRows
}

func (s stubUsers) List(context.Context) ([]model.User, error) { return nil, nil }

func (s stubUsers) Update(context.Context, *model.User) error {
	return fmt.Errorf("not implemented")
}

func (s stubUsers) Delete(context.Context, uuid.UUID) error { return fmt.Errorf("not implemented") }

func (s stubUsers) Close() error { return nil }
