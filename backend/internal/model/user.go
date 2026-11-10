package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" binding:"omitempty,uuid"`
	Name         string    `json:"name" binding:"required,min=2,max=100"`
	Email        string    `json:"email" binding:"required,email,max=255"`
	PasswordHash string    `json:"-" binding:"-"`
	Role         string    `json:"role" binding:"required,oneof=requester agent admin"`
	CreatedAt    time.Time `json:"created_at" binding:"omitempty"`
	UpdatedAt    time.Time `json:"updated_at" binding:"omitempty"`
}
