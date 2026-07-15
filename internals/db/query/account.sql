-- name: CreateAccount :one
INSERT INTO accounts (
                      owner,
                      balance,
                      currency
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE owner=$1 AND id=$2
LIMIT 1;

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id=$1
LIMIT 1;

-- More safety when use parallel calculating and transactions
-- //name: GetAccountForUpdate :one
-- SELECT * FROM accounts WHERE id=$1 LIMIT 1 FOR UPDATE;

-- name: ListAccounts :many
SELECT * FROM accounts WHERE owner=$1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: AddAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE id=sqlc.arg(id) AND owner=sqlc.arg(owner)
RETURNING *;

-- name: AddAccountBalanceByID :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE id=sqlc.arg(id)
RETURNING *;

-- name: UpdateAccountBalance :one
UPDATE accounts
SET balance = $2
WHERE id = $1 AND owner=$3
RETURNING *;

-- name: DeleteAccount :one
DELETE FROM accounts WHERE id=$1 AND owner=$2 RETURNING id;