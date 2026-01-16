package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	"github.com/nhassl3/simplebank/internals/http/simplebank/session"
	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
	"github.com/nhassl3/simplebank/internals/lib/token"
)

const (
	opCreateUser         = "domain.CreateUser"
	opGetUser            = "domain.GetUser"
	opLoginUser          = "domain.LoginUser"
	opUpdateUserPassword = "domain.UpdateUserPassword"
	opUpdateUserFullName = "domain.UpdateUserFullName"
	opDeleteUser         = "domain.DeleteUser"
)

type UserHandler struct {
	log        *slog.Logger
	store      db.Store
	tokenMaker token.Maker
}

func NewUserHandler(log *slog.Logger, store db.Store, tokenMaker token.Maker) *UserHandler {
	return &UserHandler{
		log:        log,
		store:      store,
		tokenMaker: tokenMaker,
	}
}

// CreateUser creates new user with given parameters
func (h *UserHandler) CreateUser(ctx context.Context, in session.CreateUserRequest) (*session.AuthResponse, error) {
	log := h.log.With(slog.String("op", opCreateUser))

	hashedPassword, err := argon2id.CreateHash(in.Password, argon2id.DefaultParams)
	if err != nil {
		log.Error("failed to create hash #" + in.Username)
		return nil, sl.ErrUpLevel(opCreateUser, err)
	}

	user, err := h.store.CreateUser(ctx, db.CreateUserParams{
		Username:       in.Username,
		HashedPassword: hashedPassword,
		FullName:       in.FullName,
		Email:          in.Email,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, sl.ErrorUserAlreadyExists
		}
		log.Error("failed to create user", sl.Err(err))
		return nil, sl.ErrUpLevel(opCreateUser, err)
	}

	// Create access token (15 minutes)
	accessToken, err := h.tokenMaker.CreateToken(user.Username, false)
	if err != nil {
		log.Error("failed to create token", sl.Err(err))
		return nil, sl.ErrUpLevel(opCreateUser, err)
	}

	return &session.AuthResponse{
		Token: accessToken,
		UserResponse: session.UserResponse{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: user.PasswordChangedAt,
			CreatedAt:         user.CreatedAt,
		},
	}, nil
}

// LoginUser authenticates a user and returns an access token
func (h *UserHandler) LoginUser(ctx context.Context, in session.LoginRequest) (*session.AuthResponse, error) {
	log := h.log.With("op", opLoginUser)

	// Get user's hashed password
	user, err := h.store.GetUserPrivate(ctx, in.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoUsers
		}
		log.Error("failed to get user", sl.Err(err))
		return nil, sl.ErrUpLevel(opLoginUser, err)
	}

	// Verify password
	match, err := argon2id.ComparePasswordAndHash(in.Password, user.HashedPassword)
	if err != nil {
		log.Error("failed to compare password", sl.Err(err))
		return nil, sl.ErrUpLevel(opLoginUser, err)
	}

	if !match {
		return nil, sl.ErrorUnauthorized
	}

	// Create access token (15 minutes)
	accessToken, err := h.tokenMaker.CreateToken(in.Username, user.LevelRight.Valid)
	if err != nil {
		log.Error("failed to create token", sl.Err(err))
		return nil, sl.ErrUpLevel(opLoginUser, err)
	}

	return &session.AuthResponse{
		Token: accessToken,
		UserResponse: session.UserResponse{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: user.PasswordChangedAt,
			CreatedAt:         user.CreatedAt,
		},
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, username string) (*db.GetUserRow, error) {
	log := h.log.With("op", opGetUser)

	user, err := h.store.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoUsers
		}
		log.Error("failed to get user", sl.Err(err))
		return nil, sl.ErrUpLevel(opGetUser, err)
	}

	return &user, nil
}

func (h *UserHandler) UpdateUserPassword(ctx context.Context, in session.UpdateUserPasswordRequest) (*db.UpdatePasswordRow, error) {
	log := h.log.With("op", opUpdateUserPassword)

	hashedPassword, err := h.store.GetUserPassword(ctx, in.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoUsers
		}
		log.Error("failed to get user", sl.Err(err))
		return nil, sl.ErrUpLevel(opUpdateUserPassword, err)
	}

	if match, err := argon2id.ComparePasswordAndHash(in.Password, hashedPassword); err == nil && match {
		return nil, sl.ErrorPasswordsMatch
	} else {
		if err != nil {
			log.Error("failed to compare password", sl.Err(err))
			return nil, sl.ErrUpLevel(opUpdateUserPassword, err)
		}
	}

	newHashedPassword, err := argon2id.CreateHash(in.Password, argon2id.DefaultParams)
	if err != nil {
		log.Error("failed to create hash #" + in.Username)
		return nil, sl.ErrUpLevel(opUpdateUserPassword, err)
	}

	user, err := h.store.UpdatePassword(ctx, db.UpdatePasswordParams{
		Username:       in.Username,
		HashedPassword: newHashedPassword,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoUsers
		}
		log.Error("failed to update user password", sl.Err(err))
		return nil, sl.ErrUpLevel(opUpdateUserPassword, err)
	}

	return &user, nil
}

func (h *UserHandler) UpdateUserFullName(ctx context.Context, in session.UpdateUserFullNameRequest) (*db.UpdateNameRow, error) {
	log := h.log.With("op", opUpdateUserFullName)

	user, err := h.store.UpdateName(ctx, db.UpdateNameParams{
		Username: in.Username,
		FullName: in.FullName,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoUsers
		}
		log.Error("failed to update user full name", sl.Err(err))
		return nil, sl.ErrUpLevel(opUpdateUserFullName, err)
	}

	return &user, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, username string) error {
	log := h.log.With("op", opDeleteUser)

	if err := h.store.DeleteUser(ctx, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sl.ErrorNoUsers
		}
		log.Error("failed to delete user", sl.Err(err))
		return sl.ErrUpLevel(opDeleteUser, err)
	}

	return nil
}
