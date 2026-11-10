package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
	"github.com/yourname/helpdesk/internal/repository"
)

type CommentService struct {
	comments repository.CommentRepository
	tickets  *TicketService
}

func NewCommentService(comments repository.CommentRepository, tickets *TicketService) (*CommentService, error) {
	if comments == nil || tickets == nil {
		return nil, fmt.Errorf("new comment service: dependency is nil")
	}
	return &CommentService{comments: comments, tickets: tickets}, nil
}

func (s *CommentService) CreateComment(ctx context.Context, actorID uuid.UUID, role string, ticketID uuid.UUID, body string, isInternal bool) (*model.Comment, error) {
	if _, err := s.tickets.GetTicket(ctx, actorID, role, ticketID); err != nil {
		return nil, err
	}
	if isInternal && role == RoleRequester {
		return nil, ErrForbidden
	}

	body = strings.TrimSpace(body)
	if body == "" || utf8.RuneCountInString(body) > 5000 {
		return nil, ErrInvalidInput
	}

	comment := &model.Comment{
		TicketID:   ticketID,
		AuthorID:   actorID,
		Body:       body,
		IsInternal: isInternal,
	}
	if err := s.comments.Create(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentService) ListComments(ctx context.Context, actorID uuid.UUID, role string, ticketID uuid.UUID) ([]model.Comment, error) {
	if _, err := s.tickets.GetTicket(ctx, actorID, role, ticketID); err != nil {
		return nil, err
	}

	comments, err := s.comments.ListByTicketID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if role != RoleRequester {
		return comments, nil
	}

	visible := make([]model.Comment, 0, len(comments))
	for _, comment := range comments {
		if !comment.IsInternal {
			visible = append(visible, comment)
		}
	}
	return visible, nil
}
