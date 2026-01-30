package db

import (
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	t.Helper()

	password := generateRandomPassword()

	user, args, err := createRandomUser(password)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.NotEmpty(t, args)
	require.NotZero(t, user.Username)

	require.Equal(t, args.Username, user.Username)
	require.Equal(t, args.HashedPassword, user.HashedPassword)
	require.Equal(t, args.FullName, user.FullName)
	require.Equal(t, args.Email, user.Email)

	match, err := argon2id.ComparePasswordAndHash(password, user.HashedPassword)
	require.NoError(t, err)
	require.True(t, match)
}

func TestGetUser(t *testing.T) {
	t.Helper()

	password := generateRandomPassword()

	user, args, err := createRandomUser(password)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.NotEmpty(t, args)
	require.NotZero(t, user.Username)

	u, err := store.GetUserPrivate(ctx, args.Username)
	require.NoError(t, err)
	require.NotEmpty(t, u)
	require.NotZero(t, u.Username)

	require.Equal(t, args.Username, u.Username)
	require.Equal(t, args.HashedPassword, u.HashedPassword)
	require.Equal(t, args.FullName, u.FullName)
	require.Equal(t, args.Email, u.Email)

	match, err := argon2id.ComparePasswordAndHash(password, u.HashedPassword)
	require.NoError(t, err)
	require.True(t, match)
}

func TestUpdatePassword(t *testing.T) {
	t.Helper()

	user, _, err := createRandomUser(generateRandomPassword())
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.NotZero(t, user.Username)

	newPassword := generateRandomPassword()

	hash, err := argon2id.CreateHash(newPassword, argon2id.DefaultParams)
	require.NoError(t, err)

	updatedUser, err := store.UpdatePassword(ctx, UpdatePasswordParams{
		Username:       user.Username,
		HashedPassword: hash,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.NotZero(t, updatedUser.Username)

	password, err := store.GetUserPassword(ctx, updatedUser.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, user.Username, updatedUser.Username)
	require.NotEqual(t, user.HashedPassword, password)
	require.Equal(t, hash, password)

	ok, err := argon2id.ComparePasswordAndHash(newPassword, password)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestDeleteUser(t *testing.T) {
	t.Helper()

	password := generateRandomPassword()
	user, args, err := createRandomUser(password)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.NotEmpty(t, args)
	require.NotZero(t, user.Username)

	require.NoError(t, store.DeleteUser(ctx, args.Username))

	_, err = store.GetUser(ctx, args.Username)
	require.Error(t, err)
	require.EqualError(t, err, "no rows in result set")
}

func TestUpdateFullName(t *testing.T) {
	t.Helper()

	user, args, err := createRandomUser(generateRandomPassword())
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.NotEmpty(t, args)
	require.NotZero(t, user.Username)

	updatedUser, err := store.UpdateName(ctx, UpdateNameParams{
		Username: user.Username,
		FullName: "New Full Name",
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.NotEqual(t, user.FullName, updatedUser.FullName)
	require.Equal(t, user.Email, updatedUser.Email)
	require.Equal(t, "New Full Name", updatedUser.FullName)

	require.WithinDuration(t, user.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)
}

func generateRandomPassword() string {
	return faker.Password(true, true, true, false, false, 32)
}

func createRandomUser(password string) (user User, args CreateUserParams, err error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return
	}

	args = CreateUserParams{
		Username:       faker.Username(),
		HashedPassword: hash,
		FullName:       faker.Name(),
		Email:          faker.Email(),
	}

	createdUser, err := store.CreateUser(ctx, args)
	if err != nil {
		return
	}

	user = User{
		Username:          createdUser.Username,
		HashedPassword:    hash,
		FullName:          createdUser.FullName,
		Email:             createdUser.Email,
		CreatedAt:         createdUser.CreatedAt,
		PasswordChangedAt: createdUser.PasswordChangedAt,
	}

	return
}
