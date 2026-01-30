package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/nhassl3/simplebank/internals/lib/token"
	"github.com/stretchr/testify/require"
)

func TestJWTMakerCreateToken(t *testing.T) {
	faker := gofakeit.New(0)
	secretKey, err := GenerateRandomString(32)
	require.NoError(t, err)
	require.NotZero(t, secretKey)
	require.GreaterOrEqual(t, len(secretKey), 32)

	duration := 15 * time.Minute

	maker, err := token.NewJWTMaker(secretKey, duration)
	require.NoError(t, err)
	require.NotNil(t, maker)

	username := faker.Username()

	issuedAt := time.Now()
	expireAt := issuedAt.Add(duration)

	claims := make(map[string]string)
	key, value := "ad", "true"
	claims[key] = value
	jwtToken, err := maker.CreateToken(username, claims)

	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	payload, err := maker.VerifyToken(jwtToken)
	require.NoError(t, err)
	require.NotNil(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Subject)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expireAt, payload.ExpiresAt, time.Second)
	require.Equal(t, claims, payload.Claims)
	require.Equal(t, value, payload.Claims[key])
}

func TestJWTMakerExpiredToken(t *testing.T) {
	secretKey, err := GenerateRandomString(32)
	require.NoError(t, err)
	require.NotZero(t, secretKey)
	require.GreaterOrEqual(t, len(secretKey), 32)

	maker, err := token.NewJWTMaker(secretKey, -time.Minute) // Negative duration to create expired token
	require.NoError(t, err)
	require.NotNil(t, maker)

	username := "testuser"

	jwtToken, err := maker.CreateToken(username, nil)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	// Wait a bit to ensure token is expired
	time.Sleep(time.Second)

	payload, err := maker.VerifyToken(jwtToken)
	require.Error(t, err)
	require.Nil(t, payload)
	require.EqualError(t, err, "maker.VerifyToken: token has expired")
}

func TestJWTMakerInvalidToken(t *testing.T) {
	secretKey, err := GenerateRandomString(32)
	require.NoError(t, err)
	require.NotZero(t, secretKey)
	require.GreaterOrEqual(t, len(secretKey), 32)

	maker, err := token.NewJWTMaker(secretKey, time.Minute*15)
	require.NoError(t, err)
	require.NotNil(t, maker)

	invalidToken := "invalid.token.string"

	payload, err := maker.VerifyToken(invalidToken)
	require.Error(t, err)
	require.Nil(t, payload)
	require.Contains(t, err.Error(), token.ErrInvalidToken.Error())
}

func TestJWTMakerInvalidSecretKey(t *testing.T) {
	secretKey, err := GenerateRandomString(4)
	require.NoError(t, err)
	require.NotZero(t, secretKey)
	require.GreaterOrEqual(t, len(secretKey), 4)
	require.LessOrEqual(t, len(secretKey), 31)

	_, err = token.NewJWTMaker(secretKey, time.Minute*15)
	require.Error(t, err)
	require.Contains(t, err.Error(), token.ErrInvalidSecretKey.Error())
}
