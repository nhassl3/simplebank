package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
	duration  time.Duration
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string, duration time.Duration) (*JWTMaker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf(
			"%w: %s",
			ErrInvalidSecretKey,
			fmt.Sprintf("secret key must be at least %d bytes", minSecretKeySize),
		)
	}
	return &JWTMaker{
		secretKey: secretKey,
		duration:  duration,
	}, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker JWTMaker) CreateToken(username string, isAdmin bool) (string, error) {
	payload, err := NewPayload(username, isAdmin, maker.duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	return jwtToken.SignedString([]byte(maker.secretKey))
}

// VerifyToken checks if the token is valid or not
func (maker JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %s",
				ErrInvalidToken,
				fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]),
			)
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Проверяем, что токен валиден
	if !jwtToken.Valid {
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok || payload == nil {
		return nil, fmt.Errorf("%w: %s",
			ErrInvalidToken,
			"token claims is not satisfied token does not satisfy the custom payload",
		)
	}

	return payload, nil
}
