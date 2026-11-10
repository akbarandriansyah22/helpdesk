package model

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID `json:"id" binding:"omitempty,uuid"`
	TicketID   uuid.UUID `json:"ticket_id" binding:"required,uuid"`
	AuthorID   uuid.UUID `json:"author_id" binding:"required,uuid"`
	Body       string    `json:"body" binding:"required,max=5000"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at" binding:"omitempty"`
}
