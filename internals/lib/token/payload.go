package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
)

// Ошибки валидации
var (
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidToken     = errors.New("token is invalid")
	ErrTokenNotValidYet = errors.New("token is not valid yet")
	ErrInvalidTokenID   = errors.New("token ID is invalid")
	ErrInvalidSubject   = errors.New("subject is invalid")
	ErrInvalidUUID      = errors.New("invalid UUID")
	ErrInvalidSecretKey = errors.New("secret key is invalid")
)

// Payload реализует интерфейсы как для JWT, так и для PASETO
type Payload struct {
	ID        string            `json:"jti,omitempty"`
	Audience  string            `json:"aud,omitempty"`
	Issuer    string            `json:"iss,omitempty"`
	Subject   string            `json:"sub,omitempty"`
	IssuedAt  time.Time         `json:"iat,omitempty"`
	ExpiresAt time.Time         `json:"exp,omitempty"`
	NotBefore time.Time         `json:"nbf,omitempty"`
	Claims    map[string]string `json:"claims,omitempty"`
}

// Реализация интерфейса jwt. Claims для JWT

func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(p.ExpiresAt), nil
}

func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(p.IssuedAt), nil
}

func (p *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(p.NotBefore), nil
}

func (p *Payload) GetIssuer() (string, error) {
	return p.Issuer, nil
}

func (p *Payload) GetSubject() (string, error) {
	return p.Subject, nil
}

func (p *Payload) GetAudience() (jwt.ClaimStrings, error) {
	audiences := jwt.ClaimStrings{p.Audience}
	for _, value := range p.Claims {
		audiences = append(audiences, value)
	}
	return audiences, nil
}

// Valid реализует метод Valid() из интерфейса jwt. Claims
// Этот метод будет использоваться для ручной валидации в PASETO
func (p *Payload) Valid() error {
	now := time.Now()

	// Проверяем срок действия
	if now.After(p.ExpiresAt) {
		return ErrExpiredToken
	}

	// Проверяем время "не ранее"
	if now.Before(p.NotBefore) {
		return ErrTokenNotValidYet
	}

	// Проверяем обязательные поля
	if p.ID == "" {
		return ErrInvalidTokenID
	}

	if p.Subject == "" {
		return ErrInvalidSubject
	}

	return nil
}

// NewPayload возвращает новый payload для токена
func NewPayload(username string, duration time.Duration, claims map[string]string) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, ErrInvalidUUID
	}

	now := time.Now()
	return &Payload{
		ID:        tokenID.String(),
		Issuer:    "github.com/nhassl3/simplebank",
		Subject:   username,
		IssuedAt:  now,
		ExpiresAt: now.Add(duration),
		NotBefore: now,
		Claims:    claims,
	}, nil
}

// PASETO специфичные методы

// ToPasetoJSONToken конвертирует Payload в paseto.JSONToken
// PASETO автоматически валидирует стандартные поля в JSONToken
func (p *Payload) ToPasetoJSONToken() *paseto.JSONToken {
	return &paseto.JSONToken{
		Jti:        p.ID,
		Issuer:     p.Issuer,
		Subject:    p.Subject,
		Audience:   p.Audience,
		IssuedAt:   p.IssuedAt,
		Expiration: p.ExpiresAt,
		NotBefore:  p.NotBefore,
	}
}

// ToPasetoClaimsWithFooter конвертирует Payload в claims для PASETO с пользовательскими полями
func (p *Payload) ToPasetoClaimsWithFooter() (map[string]interface{}, map[string]interface{}) {
	// Основные claims (стандартные поля PASETO)
	claims := map[string]interface{}{
		"jti": p.ID,
		"iss": p.Issuer,
		"sub": p.Subject,
		"aud": p.Audience,
		"iat": p.IssuedAt.Format(time.RFC3339),
		"exp": p.ExpiresAt.Format(time.RFC3339),
		"nbf": p.NotBefore.Format(time.RFC3339),
	}

	// Footer для дополнительных claims
	footer := make(map[string]interface{})

	// Добавляем пользовательские claims в footer
	// Можно также добавить их в основной claims, но в footer они будут скрыты
	for key, value := range p.Claims {
		footer[key] = value
	}

	return claims, footer
}

// FromPasetoJSONToken создает Payload из paseto.JSONToken
func FromPasetoJSONToken(jsonToken *paseto.JSONToken, claims map[string]string) (*Payload, error) {
	payload := &Payload{
		ID:        jsonToken.Jti,
		Issuer:    jsonToken.Issuer,
		Subject:   jsonToken.Subject,
		Audience:  jsonToken.Audience,
		IssuedAt:  jsonToken.IssuedAt,
		ExpiresAt: jsonToken.Expiration,
		NotBefore: jsonToken.NotBefore,
		Claims:    claims,
	}

	// Валидируем payload после создания
	if err := payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}

// ValidateForPaseto выполняет валидацию Payload специально для PASETO
// Этот метод можно вызывать после декодирования PASETO токена
func (p *Payload) ValidateForPaseto() error {
	// Используем ту же самую валидацию, что и для JWT
	return p.Valid()
}
