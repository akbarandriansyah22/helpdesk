package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/middleware"
	"github.com/yourname/helpdesk/internal/service"
)

func writeData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput),
		errors.Is(err, service.ErrInvalidRole),
		errors.Is(err, service.ErrInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrForbidden),
		errors.Is(err, service.ErrCannotChangeOwnRole):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrEmailTaken),
		errors.Is(err, service.ErrInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		log.Printf("handler error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return uuid.Nil, false
	}
	return id, true
}

func currentActor(c *gin.Context) (uuid.UUID, string, bool) {
	id, okID := middleware.UserID(c)
	role, okRole := middleware.Role(c)
	if !okID || !okRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, "", false
	}
	return id, role, true
}
