package token

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
	"github.com/o1egl/paseto"
)

// PASETOMaker generates security design tokens with secret key
type PASETOMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
	duration     time.Duration
}

// NewPASETOMaker creates a new PASETOMaker with PASETO method of generate tokens
func NewPASETOMaker(symmetricKey []byte, duration time.Duration) (*PASETOMaker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf(
			"%w: %s",
			ErrInvalidSecretKey,
			fmt.Sprintf("invalid key size: must be exactly %d bytes", chacha20poly1305.KeySize),
		)
	}
	return &PASETOMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: symmetricKey,
		duration:     duration,
	}, nil
}

// CreateToken creates a new PASETO token for a specific username and duration
func (maker *PASETOMaker) CreateToken(username string, claims map[string]string) (string, error) {
	return createToken(username, maker.duration, func(payload *Payload) (string, error) {
		return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	}, claims)
}

// VerifyToken checks if the token is valid or not
func (maker *PASETOMaker) VerifyToken(token string) (*Payload, error) {
	// Декодируем токен с автоматической валидацией стандартных полей
	var payload *Payload
	var footer []byte

	err := maker.paseto.Decrypt(token, maker.symmetricKey, &payload, &footer)
	if err != nil {
		// PASETO автоматически валидирует exp, nbf, iat при декодировании
		return nil, sl.ErrUpLevel(opVerifyToken, fmt.Errorf("%w: %s", ErrInvalidToken, err.Error()))
	}

	// Извлекаем пользовательские claims из footer
	customClaims := make(map[string]string)
	if len(footer) > 0 {
		err = json.Unmarshal(footer, &customClaims)
		if err != nil {
			// Если footer есть, но не парсится, это не фатально
			// Можно просто логировать или игнорировать
			customClaims = nil
		}
	}

	if err = payload.Valid(); err != nil {
		return nil, sl.ErrUpLevel(opVerifyToken, err)
	}

	return payload, nil
}
