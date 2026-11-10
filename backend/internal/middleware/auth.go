package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/service"
)

const (
	ContextUserID = "user_id"
	ContextRole   = "role"
)

func AuthMiddleware(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "auth middleware is not configured",
			})
			return
		}

		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid authorization header",
			})
			return
		}

		claims, err := auth.ParseToken(token)
		if err != nil || claims.UserID == uuid.Nil || claims.Role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func UserID(c *gin.Context) (uuid.UUID, bool) {
	value, ok := c.Get(ContextUserID)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

func Role(c *gin.Context) (string, bool) {
	value, ok := c.Get(ContextRole)
	if !ok {
		return "", false
	}
	role, ok := value.(string)
	if !ok || role == "" {
		return "", false
	}
	return role, true
}

func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	if token == "" || strings.Contains(token, " ") {
		return "", false
	}
	return token, true
}
