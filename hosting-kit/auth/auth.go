package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type Claims struct {
	UserID  uuid.UUID
	Email   string
	Name    string
	IsAdmin bool
}

type Client interface {
	Authenticate(ctx context.Context, token string) (Claims, error)
}
