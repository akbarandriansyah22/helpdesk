package model

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID          uuid.UUID  `json:"id" binding:"omitempty,uuid"`
	Title       string     `json:"title" binding:"required,min=3,max=200"`
	Description string     `json:"description" binding:"required,max=5000"`
	Category    string     `json:"category" binding:"required,min=2,max=50"`
	Priority    string     `json:"priority" binding:"required,oneof=low medium high urgent"`
	Status      string     `json:"status" binding:"omitempty,oneof=open in_progress resolved closed"`
	RequesterID uuid.UUID  `json:"requester_id" binding:"required,uuid"`
	AssigneeID  *uuid.UUID `json:"assignee_id" binding:"omitempty,uuid"`
	CreatedAt   time.Time  `json:"created_at" binding:"omitempty"`
	UpdatedAt   time.Time  `json:"updated_at" binding:"omitempty"`
}
