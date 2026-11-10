package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/helpdesk/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

type updateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=requester agent admin"`
}

func (h *UserHandler) List(c *gin.Context) {
	_, role, ok := currentActor(c)
	if !ok {
		return
	}

	users, err := h.users.GetAllUsers(c.Request.Context(), role)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusOK, users)
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	actorID, role, ok := currentActor(c)
	if !ok {
		return
	}
	userID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var request updateRoleRequest
	if !bindJSON(c, &request) {
		return
	}

	user, err := h.users.UpdateUserRole(c.Request.Context(), actorID, role, userID, request.Role)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusOK, user)
}
