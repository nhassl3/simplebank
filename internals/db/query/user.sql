-- name: CreateUser :one
INSERT INTO users (
                   username,
                   hashed_password,
                   full_name,
                   email
) VALUES (
          $1, $2, $3, $4
         ) RETURNING username, full_name, email, password_changed_at, created_at;

-- name: GetUser :one
SELECT username, full_name, email, password_changed_at, created_at FROM users WHERE username = $1 LIMIT 1;

-- name: GetUserPrivate :one
SELECT hashed_password FROM users WHERE username = $1 LIMIT 1 FOR UPDATE;

-- name: UpdatePassword :one
UPDATE users SET password_changed_at = NOW(), hashed_password = $2 WHERE username = $1 RETURNING username, full_name, email, password_changed_at, created_at;

-- name: UpdateName :one
UPDATE users SET full_name = $2 WHERE username= $1 RETURNING username, full_name, email, password_changed_at, created_at;

-- name: DeleteUser :exec
DELETE FROM users WHERE username = $1;