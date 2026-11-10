package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
	"github.com/yourname/helpdesk/internal/service"
)

type TicketHandler struct {
	tickets *service.TicketService
}

func NewTicketHandler(tickets *service.TicketService) *TicketHandler {
	return &TicketHandler{tickets: tickets}
}

type createTicketRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=200"`
	Description string `json:"description" binding:"required,max=5000"`
	Category    string `json:"category" binding:"required,min=2,max=50"`
	Priority    string `json:"priority" binding:"required,oneof=low medium high urgent"`
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open in_progress resolved closed"`
}

type patchTicketRequest struct {
	Status     *string       `json:"status" binding:"omitempty,oneof=open in_progress resolved closed"`
	Priority   *string       `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	AssigneeID assigneeField `json:"assignee_id"`
}

// assigneeField tells an omitted assignee_id apart from JSON null.
// Null clears the assignee. A UUID assigns that user.
type assigneeField struct {
	Set bool
	ID  *uuid.UUID
}

func (a *assigneeField) UnmarshalJSON(data []byte) error {
	a.Set = true
	if string(data) == "null" {
		a.ID = nil
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("assignee_id must be a UUID or null")
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return fmt.Errorf("assignee_id must be a UUID or null")
	}
	a.ID = &id
	return nil
}

func (h *TicketHandler) Create(c *gin.Context) {
	actorID, _, ok := currentActor(c)
	if !ok {
		return
	}

	var request createTicketRequest
	if !bindJSON(c, &request) {
		return
	}

	ticket, err := h.tickets.CreateTicket(c.Request.Context(), actorID, &model.Ticket{
		Title:       request.Title,
		Description: request.Description,
		Category:    request.Category,
		Priority:    request.Priority,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusCreated, ticket)
}

func (h *TicketHandler) List(c *gin.Context) {
	actorID, role, ok := currentActor(c)
	if !ok {
		return
	}

	result, err := h.tickets.ListTickets(c.Request.Context(), actorID, role, service.TicketListQuery{
		Status:     c.Query("status"),
		Priority:   c.Query("priority"),
		AssigneeID: c.Query("assignee_id"),
		Search:     c.Query("search"),
		SortBy:     c.Query("sort_by"),
		Order:      c.Query("order"),
		Page:       c.Query("page"),
		Limit:      c.Query("limit"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": result.Tickets,
		"meta": gin.H{
			"page":        result.Page,
			"limit":       result.Limit,
			"total":       result.Total,
			"total_pages": result.TotalPages,
		},
	})
}

func (h *TicketHandler) Update(c *gin.Context) {
	actorID, role, ok := currentActor(c)
	if !ok {
		return
	}
	ticketID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var request patchTicketRequest
	if !bindJSON(c, &request) {
		return
	}

	ticket, err := h.tickets.UpdateTicket(c.Request.Context(), actorID, role, ticketID, service.TicketUpdate{
		Status:      request.Status,
		Priority:    request.Priority,
		AssigneeSet: request.AssigneeID.Set,
		AssigneeID:  request.AssigneeID.ID,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusOK, ticket)
}

func (h *TicketHandler) Get(c *gin.Context) {
	actorID, role, ok := currentActor(c)
	if !ok {
		return
	}
	ticketID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	ticket, err := h.tickets.GetTicket(c.Request.Context(), actorID, role, ticketID)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateStatus(c *gin.Context) {
	actorID, role, ok := currentActor(c)
	if !ok {
		return
	}
	ticketID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var request updateStatusRequest
	if !bindJSON(c, &request) {
		return
	}

	ticket, err := h.tickets.UpdateStatus(c.Request.Context(), actorID, role, ticketID, request.Status)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusOK, ticket)
}
