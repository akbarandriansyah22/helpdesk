package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/helpdesk/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request registerRequest
	if !bindJSON(c, &request) {
		return
	}

	user, err := h.auth.Register(c.Request.Context(), request.Name, request.Email, request.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request loginRequest
	if !bindJSON(c, &request) {
		return
	}

	token, user, err := h.auth.Login(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}
