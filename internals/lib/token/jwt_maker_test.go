package token

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func generateRandomBytes(length int) (string, error) {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func TestJWTMaker(t *testing.T) {
	faker := gofakeit.New(0)
	secretKey, err := generateRandomBytes(32)
	require.NoError(t, err)
	require.NotZero(t, secretKey)
	require.GreaterOrEqual(t, len(secretKey), 32)

	maker, err := NewJWTMaker(secretKey, time.Minute*15)
	require.NoError(t, err)
	require.NotNil(t, maker)

	username := faker.Username()
	duration := time.Minute

	issuedAt := time.Now()
	expireAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(username, false)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Subject)
	require.WithinDuration(t, issuedAt, payload.IssuedAt.Time, time.Second)
	require.WithinDuration(t, expireAt, payload.ExpiresAt.Time, time.Second)
}

func TestJWTMakerExpiredToken(t *testing.T) {
	secretKey, err := generateRandomBytes(32)
	require.NoError(t, err)
	require.NotZero(t, secretKey)
	require.GreaterOrEqual(t, len(secretKey), 32)

	maker, err := NewJWTMaker(secretKey, -time.Minute) // Negative duration to create expired token
	require.NoError(t, err)
	require.NotNil(t, maker)

	username := "testuser"

	token, err := maker.CreateToken(username, false)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Wait a bit to ensure token is expired
	time.Sleep(time.Second)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.Nil(t, payload)
	require.Equal(t, ErrExpiredToken, err)
}

func TestJWTMakerInvalidToken(t *testing.T) {
	secretKey, err := generateRandomBytes(32)
	require.NoError(t, err)
	require.NotZero(t, secretKey)
	require.GreaterOrEqual(t, len(secretKey), 32)

	maker, err := NewJWTMaker(secretKey, time.Minute*15)
	require.NoError(t, err)
	require.NotNil(t, maker)

	invalidToken := "invalid.token.string"

	payload, err := maker.VerifyToken(invalidToken)
	require.Error(t, err)
	require.Nil(t, payload)
	require.Equal(t, ErrInvalidToken, err)
}

func TestJWTMakerInvalidSecretKey(t *testing.T) {
	secretKey, err := generateRandomBytes(4)
	require.NoError(t, err)
	require.NotZero(t, secretKey)
	require.GreaterOrEqual(t, len(secretKey), 4)
	require.LessOrEqual(t, len(secretKey), 31)

	_, err = NewJWTMaker(secretKey, time.Minute*15)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid secret key")
}
