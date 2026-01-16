package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Custom errors uses to handle with JWT tokens
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("expired token")
	ErrInvalidUUID      = errors.New("invalid uuid")
	ErrInvalidSecretKey = errors.New("invalid secret key")
)

// Payload is a custom variation of the jwt Claims with already written needed for validation methods
type Payload struct {
	IsAdmin bool `json:"is_admin"`
	jwt.RegisteredClaims
}

// NewPayload returns new payload for token with specific data
func NewPayload(username string, isAdmin bool, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, ErrInvalidUUID
	}

	now := time.Now()
	return &Payload{
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Issuer:    "github.com/nhassl3/simplebank",
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}, nil
}
