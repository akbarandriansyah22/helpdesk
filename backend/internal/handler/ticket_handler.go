package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	tickets, err := h.tickets.ListTickets(c.Request.Context(), actorID, role)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusOK, tickets)
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
