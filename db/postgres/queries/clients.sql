-- name: FindClientById :one
SELECT id, name, balance FROM clients WHERE id = $1;

-- name: UpdateBalanceByAmount :exec
UPDATE clients SET balance = balance + $1 WHERE id = $2;