package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
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
func (maker JWTMaker) CreateToken(username string, claims map[string]string) (string, error) {
	return createToken(username, maker.duration, func(payload *Payload) (string, error) {
		return jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(maker.secretKey))
	}, claims)
}

// VerifyToken checks if the token is valid or not
func (maker JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, sl.ErrUpLevel(opVerifyToken, fmt.Errorf("%w: %s",
				ErrInvalidToken,
				fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]),
			))
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, sl.ErrUpLevel(opVerifyToken, ErrExpiredToken)
		}
		return nil, sl.ErrUpLevel(opVerifyToken, ErrInvalidToken)
	}

	if !jwtToken.Valid {
		return nil, sl.ErrUpLevel(opVerifyToken, ErrInvalidToken)
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok || payload == nil {
		return nil, sl.ErrUpLevel(opVerifyToken, fmt.Errorf("%w: %s",
			ErrInvalidToken,
			"token claims is not satisfied token does not satisfy the custom payload",
		))
	}

	return payload, nil
}
