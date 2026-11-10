package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/helpdesk/internal/service"
)

type CommentHandler struct {
	comments *service.CommentService
}

func NewCommentHandler(comments *service.CommentService) *CommentHandler {
	return &CommentHandler{comments: comments}
}

type createCommentRequest struct {
	Body       string `json:"body" binding:"required,max=5000"`
	IsInternal bool   `json:"is_internal"`
}

func (h *CommentHandler) Create(c *gin.Context) {
	actorID, role, ok := currentActor(c)
	if !ok {
		return
	}
	ticketID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var request createCommentRequest
	if !bindJSON(c, &request) {
		return
	}

	comment, err := h.comments.CreateComment(
		c.Request.Context(),
		actorID,
		role,
		ticketID,
		request.Body,
		request.IsInternal,
	)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusCreated, comment)
}

func (h *CommentHandler) List(c *gin.Context) {
	actorID, role, ok := currentActor(c)
	if !ok {
		return
	}
	ticketID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	comments, err := h.comments.ListComments(c.Request.Context(), actorID, role, ticketID)
	if err != nil {
		writeError(c, err)
		return
	}
	writeData(c, http.StatusOK, comments)
}
