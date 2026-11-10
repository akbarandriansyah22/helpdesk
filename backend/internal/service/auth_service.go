package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourname/helpdesk/internal/model"
	"github.com/yourname/helpdesk/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenTTL     = 24 * time.Hour
	tokenIssuer  = "helpdesk"
	minPassword  = 8
	maxPassword  = 72
	minNameRunes = 2
	maxNameRunes = 100
	// dummyPasswordHash is a valid bcrypt hash used when the email is unknown,
	// so login does not return faster for a missing account.
	dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
)

type AuthClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(users repository.UserRepository, jwtSecret string) (*AuthService, error) {
	if users == nil {
		return nil, fmt.Errorf("new auth service: user repository is nil")
	}
	if strings.TrimSpace(jwtSecret) == "" {
		return nil, fmt.Errorf("new auth service: jwt secret is empty")
	}
	return &AuthService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
	}, nil
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) < minNameRunes || utf8.RuneCountInString(name) > maxNameRunes {
		return nil, ErrInvalidInput
	}

	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	existing, err := s.users.GetByEmail(ctx, email)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         RoleRequester,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return publicUser(user), nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	email, err := normalizeEmail(email)
	if err != nil || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if isNotFound(err) {
			_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(password))
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.signToken(user)
	if err != nil {
		return "", nil, err
	}
	return token, publicUser(user), nil
}

func (s *AuthService) ParseToken(tokenString string) (*AuthClaims, error) {
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return s.jwtSecret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}
	return claims, nil
}

func (s *AuthService) signToken(user *model.User) (string, error) {
	now := time.Now().UTC()
	claims := AuthClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", ErrInvalidInput
	}
	return email, nil
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < minPassword || len(password) > maxPassword {
		return ErrInvalidInput
	}
	return nil
}

func publicUser(user *model.User) *model.User {
	if user == nil {
		return nil
	}
	clone := *user
	clone.PasswordHash = ""
	return &clone
}
