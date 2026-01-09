package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	"github.com/nhassl3/simplebank/internals/http/simplebank/requests"
	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
)

const (
	opCreateUser         = "domain.CreateUser"
	opGetUser            = "domain.GetUser"
	opUpdateUserPassword = "domain.UpdateUserPassword"
	opUpdateUserFullName = "domain.UpdateUserFullName"
	opDeleteUser         = "domain.DeleteUser"
)

type UserHandler struct {
	log   *slog.Logger
	store db.Store
}

func NewUserHandler(log *slog.Logger, store db.Store) *UserHandler {
	return &UserHandler{
		log:   log,
		store: store,
	}
}

// CreateUser creates new user with given parameters
func (h *UserHandler) CreateUser(ctx context.Context, in requests.CreateUserRequest) (*db.CreateUserRow, error) {
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

	return &user, nil
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

func (h *UserHandler) UpdateUserPassword(ctx context.Context, in requests.UpdateUserPasswordRequest) (*db.UpdatePasswordRow, error) {
	log := h.log.With("op", opUpdateUserPassword)

	hashedPassword, err := h.store.GetUserPrivate(ctx, in.Username)
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

func (h *UserHandler) UpdateUserFullName(ctx context.Context, in requests.UpdateUserFullNameRequest) (*db.UpdateNameRow, error) {
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
