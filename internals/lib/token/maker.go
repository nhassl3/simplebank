package token

import (
	"fmt"
	"time"

	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
)

const (
	opCreateToken = "maker.CreateToken"
	opVerifyToken = "maker.VerifyToken"
)

// Maker is an interface for managing tokens
type Maker interface {
	CreateToken(username string, claims map[string]string) (token string, err error)
	VerifyToken(token string) (payload *Payload, err error)
}

// createToken already wrote function for using in another makers (e.x. PASETO, JWT)
func createToken(
	username string,
	duration time.Duration,
	createTokenFunc func(payload *Payload) (string, error),
	claims map[string]string,
) (token string, err error) {
	payload, err := NewPayload(username, duration, claims)
	if err != nil {
		return "", sl.ErrUpLevel(opCreateToken, fmt.Errorf("error creating token payload: %w", err))
	}

	token, err = createTokenFunc(payload)
	if err != nil {
		return "", sl.ErrUpLevel(opCreateToken, fmt.Errorf("error creating token: %w", err))
	}

	return
}
