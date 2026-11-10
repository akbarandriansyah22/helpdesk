package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
	"github.com/yourname/helpdesk/internal/repository"
)

const (
	RoleRequester = "requester"
	RoleAgent     = "agent"
	RoleAdmin     = "admin"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) (*UserService, error) {
	if users == nil {
		return nil, fmt.Errorf("new user service: user repository is nil")
	}
	return &UserService{users: users}, nil
}

func (s *UserService) GetAllUsers(ctx context.Context, actorRole string) ([]model.User, error) {
	if actorRole != RoleAgent && actorRole != RoleAdmin {
		return nil, ErrForbidden
	}

	users, err := s.users.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}

func (s *UserService) UpdateUserRole(ctx context.Context, actorID uuid.UUID, actorRole string, userID uuid.UUID, role string) (*model.User, error) {
	if actorID == uuid.Nil || userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if actorRole != RoleAdmin {
		return nil, ErrForbidden
	}
	if !validRole(role) {
		return nil, ErrInvalidRole
	}
	if actorID == userID {
		return nil, ErrCannotChangeOwnRole
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if user.Role == role {
		return publicUser(user), nil
	}

	user.Role = role
	if err := s.users.Update(ctx, user); err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return publicUser(user), nil
}

func validRole(role string) bool {
	switch role {
	case RoleRequester, RoleAgent, RoleAdmin:
		return true
	default:
		return false
	}
}
