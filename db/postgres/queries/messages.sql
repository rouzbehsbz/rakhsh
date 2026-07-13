-- name: InsertMessage :exec
INSERT INTO "messages" (
    uid, 
    created_at, 
    updated_at, 
    client_id, 
    status, 
    reason, 
    is_express, 
    recipient, 
    text
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: UpdateMessage :exec
UPDATE messages
SET
    status = $1,
    reason = $2,
    updated_at = NOW()
WHERE uid = $3;

-- name: FindMessageByUid :one
SELECT * FROM "messages" WHERE client_id = $1 AND uid = $2;