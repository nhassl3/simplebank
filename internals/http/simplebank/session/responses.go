package session

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type UserResponse struct {
	Username          string             `json:"username"`
	FullName          string             `json:"full_name"`
	Email             string             `json:"email"`
	PasswordChangedAt pgtype.Timestamptz `json:"password_changed_at"`
	CreatedAt         pgtype.Timestamptz `json:"created_at"`
}

// AuthResponse is a response for create new user or log in methods
type AuthResponse struct {
	Token        string `json:"token"`
	UserResponse `json:"user"`
}
